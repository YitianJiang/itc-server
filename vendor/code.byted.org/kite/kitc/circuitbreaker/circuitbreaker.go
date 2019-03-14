package circuit

import (
	"sync"
	"time"
)

// State changes between CLOSED, OPEN, HALFOPEN
// [CLOSED] -->- tripped ----> [OPEN]<-------+
//    ^                          |           ^
//    |                          v           |
//    +                          |      detect fail
//    |                          |           |
//    |                    cooling timeout   |
//    ^                          |           ^
//    |                          v           |
//    +--- detect succeed --<-[HALFOPEN]-->--+
//
// The behaviors of each stateus:
// =================================================================================================
// |           | [Succeed]                  | [Fail or Timeout]       | [IsAllowed]                |
// |================================================================================================
// | [CLOSED]  | do nothing                 | if tripped, become OPEN | allow                      |
// |================================================================================================
// | [OPEN]    | do nothing                 | do nothing              | if cooling timeout, allow; |
// |           |                            |                         | else reject                |
// |================================================================================================
// |           |increase halfopenSuccess,   |                         | if detect timeout, allow;  |
// |[HALFOPEN] |if(halfopenSuccess >=       | become OPEN             | else reject                |
// |           | DEFAULT_HALFOPEN_SUCCESSES)|                         |                            |
// |           |     become CLOSED          |                         |                            |
// =================================================================================================
type State int

func (s State) String() string {
	switch s {
	case OPEN:
		return "OPEN"
	case HALFOPEN:
		return "HALFOPEN"
	case CLOSED:
		return "CLOSED"
	}
	return "INVALID"
}

const (
	OPEN     State = iota
	HALFOPEN State = iota
	CLOSED   State = iota
)

const (
	// cooling timeout is the time the breaker stay in OPEN before becoming HALFOPEN
	DEFAULT_COOLING_TIMEOUT = time.Second * 5

	// detect timeout is the time interval between every detect in HALFOPEN
	DEFAULT_DETECT_TIMEOUT = time.Millisecond * 200

	// halfopen success is the threshold when the breaker is in HALFOPEN;
	// after secceeding consecutively this times, it will change its state from HALFOPEN to CLOSED;
	DEFAULT_HALFOPEN_SUCCESSES = 2
)

// TripFunc is a function called by a Breaker when error appear and
// determines whether the breaker should trip.
type TripFunc func(Metricser) bool

// StateChangeHandler .
type StateChangeHandler func(oldState, newState State, m Metricser)

// Breaker is the base of a circuit breaker.
type Breaker struct {
	Metricser // metircs all success, error and timeout within some time
	sync.RWMutex

	state           State     // state now
	openTime        time.Time // the time when the breaker become OPEN recently
	lastRetryTime   time.Time // last retry time when in HALFOPEN state
	halfopenSuccess int       // consecutive successes when HALFOPEN
	isFixed         bool

	options Options

	now func() time.Time // for test
}

// Options for Breaker
type Options struct {
	// parameters for metricser
	BucketTime time.Duration // the time each bucket holds
	BucketNums int           // the number of buckets the breaker have

	// parameters for breaker
	CoolingTimeout time.Duration // fixed when create
	DetectTimeout  time.Duration // fixed when create

	ShouldTrip         TripFunc           // can be nil
	StateChangeHandler StateChangeHandler // can be nil

	// for test
	now func() time.Time
}

// NewBreaker creates a base breaker with a specified options
func NewBreaker(options Options) (*Breaker, error) {
	if options.now == nil {
		options.now = time.Now
	}

	if options.BucketTime <= 0 {
		options.BucketTime = DEFAULT_BUCKET_TIME
	}

	if options.BucketNums <= 0 {
		options.BucketNums = DEFAULT_BUCKET_NUMS
	}

	if options.CoolingTimeout <= 0 {
		options.CoolingTimeout = DEFAULT_COOLING_TIMEOUT
	}

	if options.DetectTimeout <= 0 {
		options.DetectTimeout = DEFAULT_DETECT_TIMEOUT
	}

	metricser, err := NewWindowWithOptions(options.BucketTime, options.BucketNums)
	if err != nil {
		return nil, err
	}

	breaker := &Breaker{
		Metricser: metricser,
		now:       options.now,
		state:     CLOSED,
	}

	breaker.options = Options{
		BucketTime:         options.BucketTime,
		BucketNums:         options.BucketNums,
		CoolingTimeout:     options.CoolingTimeout,
		DetectTimeout:      options.DetectTimeout,
		ShouldTrip:         options.ShouldTrip,
		StateChangeHandler: options.StateChangeHandler,
		now:                options.now,
	}

	return breaker, nil
}

// Succeed records a success and decreases the concurrency counter by one
func (b *Breaker) Succeed() {
	b.Lock()
	switch b.state {
	case OPEN: // do nothing
	case HALFOPEN:
		b.halfopenSuccess++
		if b.halfopenSuccess == DEFAULT_HALFOPEN_SUCCESSES {
			if b.options.StateChangeHandler != nil {
				b.options.StateChangeHandler(HALFOPEN, CLOSED, b.Metricser)
			}
			b.Metricser.Reset()
			b.state = CLOSED
		}
	case CLOSED:
		b.Metricser.Succeed()
	}
	b.Unlock() // don't use defer
}

func (b *Breaker) error(isTimeout bool, trip TripFunc) {
	b.Lock()
	if isTimeout {
		b.Metricser.Timeout()
	} else {
		b.Metricser.Fail()
	}

	switch b.state {
	case OPEN: // do nothing
	case HALFOPEN: // become OPEN
		if b.options.StateChangeHandler != nil {
			b.options.StateChangeHandler(HALFOPEN, OPEN, b.Metricser)
		}
		b.openTime = time.Now()
		b.state = OPEN
	case CLOSED: // call ShouldTrip
		if trip != nil && trip(b) {
			// become OPEN and set the open time
			if b.options.StateChangeHandler != nil {
				b.options.StateChangeHandler(CLOSED, OPEN, b.Metricser)
			}
			b.openTime = time.Now()
			b.state = OPEN
		}
	}
	b.Unlock() // don't use defer
}

// Fail records a failure and decreases the concurrency counter by one
func (b *Breaker) Fail() {
	b.error(false, b.options.ShouldTrip)
}

// FailWithTrip .
func (b *Breaker) FailWithTrip(trip TripFunc) {
	b.error(false, trip)
}

// Timeout records a timeout and decreases the concurrency counter by one
func (b *Breaker) Timeout() {
	b.error(true, b.options.ShouldTrip)
}

// TimeoutWithTrip .
func (b *Breaker) TimeoutWithTrip(trip TripFunc) {
	b.error(true, trip)
}

// IsAllowed .
func (b *Breaker) IsAllowed() bool {
	return b.isAllowed()
}

// IsAllowed .
func (b *Breaker) isAllowed() bool {
	b.Lock()
	switch b.state {
	case OPEN:
		now := time.Now()
		if b.openTime.Add(b.options.CoolingTimeout).After(now) {
			b.Unlock()
			return false
		}
		// cooling timeout, then become HALFOPEN
		b.state = HALFOPEN
		b.halfopenSuccess = 0
		b.lastRetryTime = now
	case HALFOPEN:
		now := time.Now()
		if b.lastRetryTime.Add(b.options.DetectTimeout).After(now) {
			b.Unlock()
			return false
		}
		b.lastRetryTime = now
	case CLOSED:
	}

	b.Unlock() // don't use defer
	return true
}

// State returns the breaker's state now
func (b *Breaker) State() State {
	b.Lock()
	state := b.state
	b.Unlock()
	return state
}

// Reset resets this breaker
func (b *Breaker) Reset() {
	b.Lock()
	b.Metricser.Reset()
	b.state = CLOSED
	// don't change concurrency counter anyway
	b.Unlock()
}

// ThresholdTripFunc .
func ThresholdTripFunc(threshold int64) TripFunc {
	return func(m Metricser) bool {
		return m.Failures()+m.Timeouts() >= threshold
	}
}

// ConsecutiveTripFunc .
func ConsecutiveTripFunc(threshold int64) TripFunc {
	return func(m Metricser) bool {
		return m.ConseErrors() >= threshold
	}
}

// RateTripFunc .
func RateTripFunc(rate float64, minSamples int64) TripFunc {
	return func(m Metricser) bool {
		samples := m.Samples()
		return samples >= minSamples && m.ErrorRate() >= rate
	}
}
