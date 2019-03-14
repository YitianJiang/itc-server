package circuit

import "sync"

// PanelStateChangeHandler .
type PanelStateChangeHandler func(key string, oldState, newState State, m Metricser)

// Panel manages a batch of circuitbreakers
// TODO(zhangyuanjia): remove unused breaker
type Panel struct {
	sync.RWMutex
	breakers       map[string]*Breaker
	defaultOptions Options
	changeHandler  PanelStateChangeHandler
}

// NewPanel .
func NewPanel(changeHandler PanelStateChangeHandler,
	defaultOptions Options) (*Panel, error) {
	_, err := NewBreaker(defaultOptions)
	if err != nil {
		return nil, err
	}

	return &Panel{
		breakers:       make(map[string]*Breaker),
		defaultOptions: defaultOptions,
		changeHandler:  changeHandler,
	}, nil
}

// GetBreaker .
func (p *Panel) GetBreaker(key string) *Breaker {
	p.RLock()
	cb, ok := p.breakers[key]
	p.RUnlock()

	if ok {
		return cb
	}

	op := p.defaultOptions
	if p.changeHandler != nil {
		op.StateChangeHandler = func(oldState, newState State, m Metricser) {
			p.changeHandler(key, oldState, newState, m)
		}
	}
	cb, _ = NewBreaker(op)
	p.Lock()
	_, ok = p.breakers[key]
	if ok == false {
		p.breakers[key] = cb
	} else {
		cb = p.breakers[key]
	}
	p.Unlock()

	return cb
}

// RemoveAllBreakers .
func (p *Panel) RemoveAllBreakers() {
	p.Lock()
	p.breakers = make(map[string]*Breaker, 30)
	p.Unlock()
}

// DumpBreakers .
func (p *Panel) DumpBreakers() map[string]*Breaker {
	breakers := make(map[string]*Breaker)
	p.RLock()
	for k, b := range p.breakers {
		breakers[k] = b
	}
	p.RUnlock()
	return breakers
}

// Succeed .
func (p *Panel) Succeed(key string) {
	p.GetBreaker(key).Succeed()
}

// Fail .
func (p *Panel) Fail(key string) {
	p.GetBreaker(key).Fail()
}

// FailWithTrip .
func (p *Panel) FailWithTrip(key string, trip TripFunc) {
	p.GetBreaker(key).FailWithTrip(trip)
}

// Timeout .
func (p *Panel) Timeout(key string) {
	p.GetBreaker(key).Timeout()
}

// TimeoutWithTrip .
func (p *Panel) TimeoutWithTrip(key string, trip TripFunc) {
	p.GetBreaker(key).TimeoutWithTrip(trip)
}

// IsAllowed .
func (p *Panel) IsAllowed(key string) bool {
	return p.GetBreaker(key).IsAllowed()
}
