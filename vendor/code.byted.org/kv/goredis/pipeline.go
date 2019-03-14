package goredis

import (
	"errors"
	"sync"
	"time"

	redis "code.byted.org/kv/redis-v6"
)

type Pipeline struct {
	redis.Pipeliner
	c                  *Client
	cluster            string
	psm                string
	metricsServiceName string
	name               string
}

var ppool = &sync.Pool{New: func() interface{} { return make(map[string]int, 5) }}

// func (p *Pipeline) SetPSM

// Exec executes all previously queued commands using one client-server roundtrip.
//
// Exec returns list of commands and error.
// miss is not a error in pipeline.
// you should use Cmder.Err() == redis.Nil to find whether miss occur or not
//
// After Commit, you should use Close() to close the pipeline releasing open resources.
func (p *Pipeline) Exec() (ret []redis.Cmder, rerr error) {
	// degredate
	/*cmds := p.Pipeliner.Cmds()
	notDegCmds := make([]redis.Cmder, 0, len(cmds))
	for _, c := range cmds {
		if cmdDegredated(p.metricsServiceName, c.Name()) {
			c.SetErr(ErrDegradated)
		} else {
			notDegCmds = append(notDegCmds, c)
		}
	}
	p.Pipeliner.SetCmds(notDegCmds)
	*/
	// hack for stress tag
	var stressCmds []redis.Cmder
	if prefix, ok := isStressTest(p.c.ctx); ok {
		stressCmds = make([]redis.Cmder, 0, len(p.Pipeliner.Cmds()))
		for _, c := range p.Pipeliner.Cmds() {
			stressCmds = append(stressCmds, convertStressCMD(prefix, c))
		}
		// modify pipeline's cmds
		p.Pipeliner.SetCmds(stressCmds)
	}

	start := time.Now().UnixNano()
	cmder, _ := p.Pipeliner.Exec()
	latency := (time.Now().UnixNano() - start) / 1000

	var resErr error

	pipelineCmdNum := len(cmder)
	if pipelineCmdNum == 0 {
		resErr = errors.New("pipeline cmd num is 0")
	}

	cmdErrorCounter := ppool.Get().(map[string]int)
	cmdSuccessCounter := ppool.Get().(map[string]int)
	for k, _ := range cmdErrorCounter {
		delete(cmdErrorCounter, k)
	}
	for k, _ := range cmdSuccessCounter {
		delete(cmdSuccessCounter, k)
	}
	for _, res := range cmder {
		cmdStr := res.Name()
		// one or more miss occur in pipeline, and we think miss is not a error in pipeline
		if res.Err() != nil && res.Err() != redis.Nil {
			counter, ok := cmdErrorCounter[cmdStr]
			if ok {
				cmdErrorCounter[cmdStr] = counter + 1
			} else {
				cmdErrorCounter[cmdStr] = 1
			}
			if resErr == nil {
				resErr = res.Err()
			}
		} else {
			counter, ok := cmdSuccessCounter[cmdStr]
			if ok {
				cmdSuccessCounter[cmdStr] = counter + 1
			} else {
				cmdSuccessCounter[cmdStr] = 1
			}
		}
	}
	// Aggregate pipeline cmd metrics by cmdStr
	// separate cmd
	for cmdStr, counter := range cmdSuccessCounter {
		addCallMetrics(p.c.ctx, cmdStr, -1, nil, p.c.cluster, p.c.psm, p.c.metricsServiceName, counter)
	}
	for cmdStr, counter := range cmdErrorCounter {
		addCallMetrics(p.c.ctx, cmdStr, -1, resErr, p.c.cluster, p.c.psm, p.c.metricsServiceName, counter)
	}
	addCallMetrics(p.c.ctx, "pipeline", latency, resErr, p.c.cluster, p.c.psm, p.c.metricsServiceName, 1)

	if pipelineCmdNum > 500 {
		addCallMetrics(p.c.ctx, "big_pipeline", -1, nil, p.c.cluster, p.c.psm, p.c.metricsServiceName, 1)
	}
	ppool.Put(cmdErrorCounter)
	ppool.Put(cmdSuccessCounter)
	return cmder, resErr
}
