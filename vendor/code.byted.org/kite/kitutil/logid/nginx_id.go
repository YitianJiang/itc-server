package logid

import (
	"encoding/hex"
	"math/rand"
	"net"
	"time"
)

const (
	_DEFAULT_IP = "000000000000"
)

var localIP []byte = findIP()

func init() {
	rand.Seed(time.Now().Unix())
}

func GetNginxID() string {
	ret := [33]byte{}
	t := time.Now()
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	int2byte(ret[:4], year)
	int2byte(ret[4:6], int(month))
	int2byte(ret[6:8], day)
	int2byte(ret[8:10], hour)
	int2byte(ret[10:12], minute)
	int2byte(ret[12:14], second)
	copy(ret[14:26], localIP)
	ms := t.UnixNano() / 1e6 % 1000
	int2byte(ret[26:29], int(ms))
	u32 := rand.Uint32()
	u32 >>= 16
	src := []byte{byte(u32 & 0xff), byte((u32 >> 8) & 0xff)}
	hex.Encode(ret[29:33], src)
	for idx := 29; idx < 33; idx++ {
		ret[idx] = upper(ret[idx])
	}
	return string(ret[:])
}

func upper(c byte) byte {
	val := uint8(c)
	if val >= 97 && val <= 122 {
		return byte(val - 32)
	}
	return c
}

func int2byte(bs []byte, val int) {
	size := 10
	l := len(bs) - 1
	for idx := l; idx >= 0; idx-- {
		bs[idx] = byte(uint(val%size) + uint('0'))
		val = val / size
	}
}

func findIP() []byte {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []byte(_DEFAULT_IP)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				return formatIP(ip)
			}
		}
	}
	return []byte(_DEFAULT_IP)
}

func formatIP(ip string) []byte {
	realIP := []byte(_DEFAULT_IP)
	idx := 0
	for i := len(ip) - 1; i >= 0; i-- {
		c := ip[i]
		if c == '.' {
			idx = (((idx - 1) / 3) + 1) * 3
			continue
		}
		realIP[idx] = byte(c)
		idx++
	}
	reverse(realIP)
	return realIP
}

func reverse(bs []byte) {
	i, j := 0, len(bs)-1
	for i < j {
		bs[i], bs[j] = bs[j], bs[i]
		i++
		j--
	}
}
