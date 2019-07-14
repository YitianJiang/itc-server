package mysql

import (
	"code.byted.org/kite/kitc/circuitbreaker"
	"code.byted.org/gopkg/context"
	"code.byted.org/gopkg/logs"
	"database/sql/driver"
)

var (
	driverCbPanel *circuit.Panel
)

func init() {
	driverCbPanel, _ = circuit.NewPanel(CBChangeHandler, circuit.Options{})
}

func cbOpen(r MysqlReqMeta) bool {
	if rConfig := getRemoteConfig(r); !rConfig.CBIsOpen {
		return false
	}

	breaker := driverCbPanel.GetBreaker(r.String())
	if breaker != nil {
		return !breaker.IsAllowed()
	}

	return false
}

func doCBMetrics(ctx context.Context, r MysqlReqMeta, err error) {
	if err == driver.ErrSkip {
		return
	}
	// cb and degradation err do not metrics
	errCode := getMysqlErrCode(err)
	if errCode == ErrNotAllowedByServiceCBCode || errCode == ErrForbiddenByDegradationCode {
		return
	}

	cbKey := r.String()
	if errCode != 0 {
		config := getRemoteConfig(r)
		driverCbPanel.FailWithTrip(cbKey, circuit.RateTripFunc(config.CBErrRate, int64(config.CBMinSample)))
	} else {
		driverCbPanel.Succeed(cbKey)
	}
}

func CBChangeHandler(key string, oldState, newState circuit.State, m circuit.Metricser) {
	logs.Warnf("[mysql-driver] cb change handler: %s: %s -> %s, (succ: %d, err: %d, tmout: %d, rate: %f)",
		key, oldState, newState, m.Successes(), m.Failures(), m.Timeouts(), m.ErrorRate())
}
