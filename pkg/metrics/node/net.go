package node

import (
	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/net"
	"strings"
)

func netUsage(stat net.IOCountersStat) *gen.NetUsage {
	return &gen.NetUsage{
		TotalBytesSent:       stat.BytesSent,
		TotalBytesReceived:   stat.BytesRecv,
		TotalPacketsSent:     stat.PacketsSent,
		TotalPacketsReceived: stat.PacketsRecv,
		TotalErrIn:           stat.Errin,
		TotalErrOut:          stat.Errout,
		TotalDropIn:          stat.Dropin,
		TotalDropOut:         stat.Dropout,
		TotalFifoErrIn:       stat.Fifoin,
		TotalFifoErrOut:      stat.Fifoout,
	}
}

func networkInterfaces(interfaces net.InterfaceStatList) []*gen.InterfaceStat {
	networkInterfaces := make([]*gen.InterfaceStat, 0, len(interfaces))

	for _, stat := range interfaces {
		addresses := make([]string, 0, len(stat.Addrs))
		for _, addr := range stat.Addrs {
			addresses = append(addresses, addr.Addr)
		}

		networkInterfaces = append(networkInterfaces, &gen.InterfaceStat{
			Index:        int32(stat.Index),
			Mtu:          int32(stat.MTU),
			Name:         stat.Name,
			HardwareAddr: stat.HardwareAddr,
			Flags:        stat.Flags,
			Addrs:        addresses,
		})
	}
	return networkInterfaces
}

func getPrimaryIPs(interfaces []*gen.InterfaceStat) (string, string) {
	skipPrefixes := []string{"lo", "cali", "veth", "docker", "br-", "tunl", "flannel"}

	var ipv4Addr, ipv6Addr string

	for _, iface := range interfaces {
		skip := false
		for _, prefix := range skipPrefixes {
			if strings.HasPrefix(iface.Name, prefix) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		isLoopback := false
		for _, flag := range iface.Flags {
			if strings.ToLower(flag) == "loopback" {
				isLoopback = true
				break
			}
		}
		if isLoopback {
			continue
		}

		for _, addr := range iface.Addrs {
			ip := strings.Split(addr, "/")[0]

			if strings.HasPrefix(ip, "127.") || ip == "::1" {
				continue
			}

			if strings.Contains(ip, ":") {
				if ipv6Addr == "" {
					ipv6Addr = ip
				}
			} else {
				if ipv4Addr == "" {
					ipv4Addr = ip
				}
			}

			// Stop if both addresses found
			if ipv4Addr != "" && ipv6Addr != "" {
				return ipv4Addr, ipv6Addr
			}
		}
	}
	return ipv4Addr, ipv6Addr
}
