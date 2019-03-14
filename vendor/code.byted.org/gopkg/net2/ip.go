package net2

import (
	"net"
)

var (
	localIP    net.IP
	localIPStr string

	privateNets []net.IPNet
)

const UnknownIPAddr = "-"

func init() {
	for _, s := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"} {
		_, n, _ := net.ParseCIDR(s)
		privateNets = append(privateNets, *n)
	}

	// get all network interfaces
	netIfaces, err := net.Interfaces()
	if err != nil {
		return
	}

	// get all private ips from non-loopback active net interfaces
	var ips []net.IP
	for _, netIface := range netIfaces {
		if netIface.Flags&net.FlagLoopback != 0 {
			// skip all Loopback interface
			continue
		}
		if netIface.Flags&net.FlagUp == 0 {
			// skip interface not UP
			continue
		}
		addrs, err := netIface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ipnet.IP.IsLoopback() {
				continue
			}
			ip := ipnet.IP
			if IsPrivateIP(ip) {
				ips = append(ips, ip)
			}
		}
	}

	// set local ip by priority of privateNets
	for _, pnet := range privateNets {
		for _, ip := range ips {
			if pnet.Contains(ip) {
				localIP = ip
				localIPStr = ip.String()
				return
			}
		}
	}
}

func GetLocalIP() net.IP {
	return localIP
}

func GetLocalIPStr() string {
	return localIPStr
}

// deprecated: use GetLocalIP or GetLocalIPStr
func GetLocalIp() string {
	if localIPStr == "" {
		return UnknownIPAddr
	}
	return localIPStr
}

func IsPrivateIP(ip net.IP) bool {
	for _, n := range privateNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
