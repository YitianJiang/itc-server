package kite

import (
	"strconv"
	"strings"
)

// encode host from string to uint64
func encodeAddr(addr string) uint64 {
	hostPort := strings.Split(addr, ":")
	if len(hostPort) != 2 {
		return 0
	}

	secs := strings.Split(hostPort[0], ".")
	if len(secs) != 4 {
		return 0
	}
	var host uint64
	for _, sec := range secs {
		v, err := strconv.Atoi(sec)
		if err != nil {
			return 0
		}
		host = host<<8 + uint64(v)
	}

	return host
}

func decodeAddr(code uint64) string {
	if code == 0 {
		return ""
	}

	host := code
	secs := make([]string, 4)
	for i := 3; i >= 0; i-- {
		sec := host & 0xff
		secs[i] = strconv.Itoa(int(sec))
		host >>= 8
	}

	return strings.Join(secs, ".")
}
