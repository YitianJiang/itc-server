package tos

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"code.byted.org/gopkg/consul"
)

const FAST_FAIL_CNT = 3
const REFRESH_TIME_GAP = 10
const DOWN_WEIGHT_SLEEP_GAP = 60
const INIT_WEIGHT = 100
const DOWN_WEIGHT_THRESHOLD = 80
const DEFAULT_SERVICE_NAME = "toutiao.tos.tosapi"

type addr struct {
	ipPort string
	weight int
}

type addrManager struct {
	mutex             sync.Mutex
	detector          *addrDetector
	alloctor          *addrAllocator
	statsor           *addrStatsor
	underWeightAddrTS map[string]int64
	cluster           string
	idc               string
	lastRefreshTS     int64
}

func (man *addrManager) getAddr() string {
	var allocator *addrAllocator
	var statsor *addrStatsor
	man.mutex.Lock()
	if time.Now().Unix()-man.lastRefreshTS > REFRESH_TIME_GAP {
		man.refresh()
	}
	allocator = man.alloctor
	statsor = man.statsor
	man.mutex.Unlock()

	ipPort := allocator.getOneAddr()
	statsor.cnt(ipPort)

	return ipPort
}

func (man *addrManager) cntSucc(ipPort string) {
	var statsor *addrStatsor
	man.mutex.Lock()
	statsor = man.statsor
	man.mutex.Unlock()
	statsor.cntSucc(ipPort)
}

func (man *addrManager) cntFail(ipPort string) {
	var statsor *addrStatsor
	var allocator *addrAllocator

	man.mutex.Lock()
	statsor = man.statsor
	allocator = man.alloctor
	man.mutex.Unlock()

	failInARow := statsor.cntFail(ipPort)
	if failInARow >= FAST_FAIL_CNT {
		allocator.rmOneAddr(ipPort)
		man.detector.addOneAddr(ipPort)
	}
}

func (man *addrManager) fastCntFail(ipPort string) {
	var allocator *addrAllocator

	man.mutex.Lock()
	allocator = man.alloctor
	man.mutex.Unlock()

	allocator.rmOneAddr(ipPort)
	man.detector.addOneAddr(ipPort)
}

func newAddrManager(cluster, idc string) (*addrManager, error) {
	detector := getDetector()

	man := &addrManager{
		detector:          detector,
		alloctor:          newAddrAllocator([]addr{}, 0),
		statsor:           newAddrStatsor([]addr{}),
		underWeightAddrTS: make(map[string]int64),
		cluster:           cluster,
		idc:               idc,
		lastRefreshTS:     0,
	}

	return man, nil
}

var testAddr = os.Getenv("TEST_TOSAPI_ADDR")

func getEndpoints(idc, cluster string) (consul.Endpoints, error) {
	if testAddr != "" {
		addrList := strings.Split(testAddr, ";")
		if len(addrList) == 0 {
			return consul.Endpoints{}, nil
		}

		ret := make(consul.Endpoints, len(addrList))
		for i, addr := range addrList {
			ret[i] = consul.Endpoint{Addr: addr}
		}
		return ret, nil
	}
	name := DEFAULT_SERVICE_NAME
	if idc != "" {
		name += ".service." + idc
	}
	return consul.Lookup(name, consul.WithCluster(cluster))
}

func (man *addrManager) cleanExpireUnderWeightAddr(candiAddrList consul.Endpoints) {
	curAddrSet := make(map[string]struct{})
	for _, addr := range candiAddrList {
		curAddrSet[addr.Addr] = struct{}{}
	}

	nowUnix := time.Now().Unix()
	for ipPort, timestamp := range man.underWeightAddrTS {
		if timestamp < nowUnix-DOWN_WEIGHT_SLEEP_GAP {
			delete(man.underWeightAddrTS, ipPort)
			continue
		}
		if _, ok := curAddrSet[ipPort]; !ok {
			delete(man.underWeightAddrTS, ipPort)
		}
	}
}

func (man *addrManager) cleanDetectorNotExistAddr(candiAddrList consul.Endpoints, inDetectAddrList []string) {
	existAddrMap := make(map[string]struct{})
	for _, addr := range candiAddrList {
		existAddrMap[addr.Addr] = struct{}{}
	}

	notExistAddrList := make([]string, 0)
	for _, addr := range inDetectAddrList {
		if _, ok := existAddrMap[addr]; !ok {
			notExistAddrList = append(notExistAddrList, addr)
		}
	}
	if len(notExistAddrList) == 0 {
		return
	}
	man.detector.rmAddres(notExistAddrList)
}

func (man *addrManager) calNewAddrList(candiAddrList consul.Endpoints) []addr {
	// we have addr
	//   1. above threshhold, will survive in next round
	//   2. below threshhold, need to sleep a long time
	//   3. fast fail, need to be detect
	man.cleanExpireUnderWeightAddr(candiAddrList)

	prevRoundAddrList := man.alloctor.getAddrList()
	prevRoundAddr2Weight := make(map[string]int)
	for _, addr := range prevRoundAddrList {
		prevRoundAddr2Weight[addr.ipPort] = addr.weight
	}

	inDetectAddrList := man.detector.getAddrList()
	inDetectAddrMap := make(map[string]struct{})
	for _, addr := range inDetectAddrList {
		inDetectAddrMap[addr] = struct{}{}
	}
	man.cleanDetectorNotExistAddr(candiAddrList, inDetectAddrList)

	chosenAddrList := make([]addr, 0)
	underWeightAddrList := make([]addr, 0)
	for _, endpoint := range candiAddrList {
		_, ok := man.underWeightAddrTS[endpoint.Addr]
		if ok {
			continue
		}

		_, ok = inDetectAddrMap[endpoint.Addr]
		if ok { // still in detection recover state, should not expose to user
			inDetectAddrList = append(inDetectAddrList, endpoint.Addr)
			continue
		}

		curWeight, ok := prevRoundAddr2Weight[endpoint.Addr]
		if !ok {
			chosenAddrList = append(chosenAddrList, addr{
				ipPort: endpoint.Addr,
				weight: INIT_WEIGHT,
			})
			continue
		}

		succRatio := man.statsor.succRatio(endpoint.Addr)
		if succRatio == 100 {
			chosenAddrList = append(chosenAddrList, addr{
				ipPort: endpoint.Addr,
				weight: INIT_WEIGHT,
			})
		} else if succRatio*curWeight >= 100*DOWN_WEIGHT_THRESHOLD {
			chosenAddrList = append(chosenAddrList, addr{
				ipPort: endpoint.Addr,
				weight: succRatio * curWeight / 100,
			})
		} else {
			underWeightAddrList = append(underWeightAddrList, addr{endpoint.Addr, succRatio * curWeight})
		}
	}

	// selected addr list should at lease have half addres of all addres
	// to prevent that half server cannot handle all rate
	if len(chosenAddrList)*2 < len(candiAddrList) {
		if len(underWeightAddrList)+len(chosenAddrList) < len(candiAddrList)/2 {
			for ipPort, _ := range man.underWeightAddrTS {
				if len(underWeightAddrList)+len(chosenAddrList) > len(candiAddrList)/2 {
					break
				}
				underWeightAddrList = append(underWeightAddrList, addr{ipPort, DOWN_WEIGHT_THRESHOLD})
				delete(man.underWeightAddrTS, ipPort)
			}
		}

		sort.Slice(underWeightAddrList, func(i, j int) bool {
			return underWeightAddrList[i].weight >= underWeightAddrList[j].weight
		})

		i := 0
		for i = 0; i+1+len(chosenAddrList) < len(candiAddrList)/2 && i < len(underWeightAddrList); i++ {
			chosenAddrList = append(chosenAddrList, underWeightAddrList[i])
		}

		underWeightAddrList = underWeightAddrList[i:]
	}

	nowUnix := time.Now().Unix()
	for _, addr := range underWeightAddrList {
		if _, ok := man.underWeightAddrTS[addr.ipPort]; !ok {
			man.underWeightAddrTS[addr.ipPort] = nowUnix
		}
	}

	return chosenAddrList
}

func (man *addrManager) refresh() {
	man.lastRefreshTS = time.Now().Unix()
	endpoints, err := getEndpoints(man.idc, man.cluster)
	if err != nil || len(endpoints) == 0 {
		man.lastRefreshTS = time.Now().Unix()
		return
	}

	newAddrList := man.calNewAddrList(endpoints)
	if len(newAddrList) == 0 {
		man.lastRefreshTS = time.Now().Unix()
		fmt.Printf("got 0 addr list")
		return
	}
	if len(newAddrList) <= len(endpoints)/2 {
		man.lastRefreshTS = time.Now().Unix()
		return
	}

	man.alloctor = newAddrAllocator(newAddrList, len(endpoints)/2)
	man.statsor = newAddrStatsor(newAddrList)
	man.lastRefreshTS = time.Now().Unix()
}
