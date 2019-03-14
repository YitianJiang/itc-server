package tos

import (
	"net"
	"sync"
	"time"
)

const EXPIRE_GAP = 600
const DETECT_GAP = 30
const CONN_TIMEOUT = 1

var globalMutex sync.Mutex
var globalDetector *addrDetector

func getDetector() *addrDetector {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	if globalDetector == nil {
		globalDetector = &addrDetector{}
		go globalDetector.loopDetect()
	}
	return globalDetector
}

type toDetectAddr struct {
	ipPort     string
	timestatmp int64
}

type addrDetector struct {
	//telnetClient
	rwMutex          sync.RWMutex
	toDetectAddrList []toDetectAddr
}

func (detector *addrDetector) addOneAddr(ipPort string) {
	detector.rwMutex.Lock()
	defer detector.rwMutex.Unlock()

	for _, addr := range detector.toDetectAddrList {
		if addr.ipPort == ipPort {
			return
		}
	}
	toAddAddr := toDetectAddr{
		ipPort:     ipPort,
		timestatmp: time.Now().Unix(),
	}
	detector.toDetectAddrList = append(detector.toDetectAddrList, toAddAddr)
}

func (detector *addrDetector) rmOneAddr(ipPort string) {
	detector.rmAddres([]string{ipPort})
}

func (detector *addrDetector) rmAddres(ipPorts []string) {
	if len(ipPorts) == 0 {
		return
	}

	detector.rwMutex.Lock()
	defer detector.rwMutex.Unlock()
	ipPortsSet := make(map[string]struct{})
	for _, ipPort := range ipPorts {
		ipPortsSet[ipPort] = struct{}{}
	}

	newToDetectList := make([]toDetectAddr, 0)
	for _, addr := range detector.toDetectAddrList {
		_, ok := ipPortsSet[addr.ipPort]
		if ok {
			continue
		}
		newToDetectList = append(newToDetectList, addr)
	}
	detector.toDetectAddrList = newToDetectList
}

func (detector *addrDetector) getAddrList() []string {
	detector.rwMutex.RLock()
	defer detector.rwMutex.RUnlock()

	ipPorts := make([]string, len(detector.toDetectAddrList))
	for i, addr := range detector.toDetectAddrList {
		ipPorts[i] = addr.ipPort
	}
	return ipPorts
}

//TODO catch panic and recover
func (detector *addrDetector) loopDetect() {
	for {
		toDetectIpPort := detector.getOneToDetect(time.Now().Unix())
		if toDetectIpPort == "" {
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}

		if !detect(toDetectIpPort) {
			detector.addOneAddr(toDetectIpPort)
		}
	}
}

func (detector *addrDetector) getOneToDetect(nowUnix int64) string {
	detector.rwMutex.Lock()
	defer detector.rwMutex.Unlock()
	if len(detector.toDetectAddrList) == 0 {
		return ""
	}

	startIdx := 0
	selectedAddr := toDetectAddr{}
	iterAddr := toDetectAddr{}
	for ; startIdx < len(detector.toDetectAddrList); startIdx++ {
		iterAddr = detector.toDetectAddrList[startIdx]
		if iterAddr.timestatmp < nowUnix-EXPIRE_GAP {
			continue
		}
		if iterAddr.timestatmp > nowUnix-DETECT_GAP {
			break
		}
		selectedAddr = iterAddr
		startIdx = startIdx + 1
		break
	}

	detector.toDetectAddrList = detector.toDetectAddrList[startIdx:]
	return selectedAddr.ipPort
}

func detect(ipPort string) bool {
	conn, err := net.DialTimeout("tcp", ipPort, time.Duration(CONN_TIMEOUT)*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
