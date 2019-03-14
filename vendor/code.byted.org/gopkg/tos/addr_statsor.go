package tos

import (
	"sync/atomic"
	"time"
)

type statRecord struct {
	totalCnt      int32
	succCnt       int32
	failCnt       int32
	failInARowCnt int32
}

type addrStatsor struct {
	addr2Stats map[string]*statRecord
	timestamp  int64
}

func newAddrStatsor(addrList []addr) *addrStatsor {
	statsor := &addrStatsor{
		addr2Stats: make(map[string]*statRecord),
		timestamp:  time.Now().Unix(),
	}
	for _, addr := range addrList {
		statsor.addr2Stats[addr.ipPort] = &statRecord{}
	}
	return statsor
}

func (statsor *addrStatsor) cnt(ipPort string) {
	record, ok := statsor.addr2Stats[ipPort]
	if !ok {
		return
	}
	atomic.AddInt32(&record.totalCnt, 1)
}

func (statsor *addrStatsor) cntSucc(ipPort string) {
	record, ok := statsor.addr2Stats[ipPort]
	if !ok {
		return
	}
	atomic.StoreInt32(&record.failInARowCnt, 0)
	atomic.AddInt32(&record.succCnt, 1)
}

// return fail cnt in a row
func (statsor *addrStatsor) cntFail(ipPort string) int32 {
	record, ok := statsor.addr2Stats[ipPort]
	if !ok {
		return 0
	}
	atomic.AddInt32(&record.failCnt, 1)
	return atomic.AddInt32(&record.failInARowCnt, 1)
}

func (statsor *addrStatsor) succRatio(ipPort string) int {
	record, ok := statsor.addr2Stats[ipPort]
	if !ok {
		return 100
	}

	totalCnt := int(atomic.LoadInt32(&record.totalCnt))
	failCnt := int(atomic.LoadInt32(&record.failCnt))
	if totalCnt == 0 || failCnt == 0 {
		return 100
	}

	failRatio := failCnt * 100 / totalCnt
	if failRatio >= 100 {
		return 0
	}

	if 100-failRatio >= 98 {
		return 100
	}
	return 100 - failRatio
}
