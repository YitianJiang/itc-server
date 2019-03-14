package tos

import (
	"math/rand"
	"sort"
	"sync"
	"time"
)

type allocInternalAddr struct {
	ipPort         string
	weight         int
	soFarWeightSum int
}

type addrAllocator struct {
	rwMutex     sync.RWMutex
	timestamp   int64
	addrList    []allocInternalAddr
	totalWeight int
	threshHold  int
}

func newAddrAllocator(addrList []addr, threshHold int) *addrAllocator {
	allocator := &addrAllocator{
		timestamp:  time.Now().Unix(),
		threshHold: threshHold,
	}
	allocator.init(addrList)
	return allocator
}

func (allocator *addrAllocator) getOneAddr() string {
	allocator.rwMutex.RLock()
	defer allocator.rwMutex.RUnlock()
	if len(allocator.addrList) == 0 {
		return ""
	}
	randWeight := rand.Intn(int(allocator.totalWeight))
	hitIdx := sort.Search(len(allocator.addrList), func(i int) bool { return allocator.addrList[i].soFarWeightSum > randWeight })
	return allocator.addrList[hitIdx].ipPort
}

func (allocator *addrAllocator) rmOneAddr(ipPort string) {
	allocator.rwMutex.Lock()
	defer allocator.rwMutex.Unlock()

	newAddrList := make([]allocInternalAddr, len(allocator.addrList))
	found := false
	curIdx := 0
	soFarWeightSum := 0

	for _, addr := range allocator.addrList {
		if addr.ipPort == ipPort {
			found = true
			continue
		}
		soFarWeightSum = soFarWeightSum + addr.weight
		newAddrList[curIdx].soFarWeightSum = soFarWeightSum
		newAddrList[curIdx].ipPort = addr.ipPort
		newAddrList[curIdx].weight = addr.weight
		curIdx = curIdx + 1
	}

	// cannot rm too many addr
	if !found || len(allocator.addrList) <= allocator.threshHold {
		return
	}
	allocator.addrList = newAddrList[:curIdx]
	allocator.totalWeight = soFarWeightSum

}

func (allocator *addrAllocator) getAddrList() []addr {
	allocator.rwMutex.RLock()
	defer allocator.rwMutex.RUnlock()
	addrList := make([]addr, len(allocator.addrList))
	for idx, addr := range allocator.addrList {
		addrList[idx].ipPort = addr.ipPort
		addrList[idx].weight = addr.weight
	}
	return addrList
}

func (allocator *addrAllocator) init(addrList []addr) {
	allocator.Shuffle(len(addrList), func(i, j int) {
		addrList[i], addrList[j] = addrList[j], addrList[i]
	})

	soFarWeightSum := 0
	allocator.addrList = make([]allocInternalAddr, len(addrList))
	for idx, addr := range addrList {
		soFarWeightSum = soFarWeightSum + addr.weight
		allocator.addrList[idx].soFarWeightSum = soFarWeightSum
		allocator.addrList[idx].ipPort = addr.ipPort
		allocator.addrList[idx].weight = addr.weight
	}
	allocator.totalWeight = soFarWeightSum
}

func (allocator *addrAllocator) Shuffle(n int, swap func(i, j int)) {
	rand.Seed(time.Now().UnixNano())
	i := n - 1
	for ; i > 0; i-- {
		j := int(rand.Int31n(int32(i + 1)))
		swap(i, j)
	}
}
