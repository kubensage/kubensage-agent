package node

import (
	"github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
	"strings"
)

// buildNetUsage converts a gopsutil net.IOCountersStat into a gen.NetUsage proto message.
//
// It extracts cumulative network statistics such as bytes sent/received, packets,
// errors, and drops. Designed to be used for node-level interface aggregation.
//
// Parameters:
//   - stat: net.IOCountersStat with network counters from gopsutil
//   - logger: Logger used for debug messages
//
// Returns:
//   - *gen.NetUsage containing summarized network metrics
func buildNetUsage(
	stat net.IOCountersStat,
	logger *zap.Logger,
) *gen.NetUsage {
	logger.Debug("Start buildNetUsage")

	netUsage := &gen.NetUsage{
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

	logger.Debug("Finish buildNetUsage")

	return netUsage
}

// ListNetworkInterfaces maps gopsutil network interfaces into proto InterfaceStat messages.
//
// Each interface includes its name, index, MTU, hardware address, flags,
// and assigned IP addresses. This function helps serialize interface metadata
// for transmission or storage.
//
// Parameters:
//   - interfaces: List of network interfaces from gopsutil
//   - logger: Logger for debug tracing
//
// Returns:
//   - []*gen.InterfaceStat representing all valid system interfaces
func listNetworkInterfaces(
	interfaces net.InterfaceStatList,
	logger *zap.Logger,
) []*gen.InterfaceStat {
	logger.Debug("Start listNetworkInterfaces")

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

	logger.Debug("Finish listNetworkInterfaces")

	return networkInterfaces
}

// buildPrimaryIPs determines the primary IPv4 and IPv6 addresses from a list of interfaces.
//
// It skips loopback, virtual, and known non-routable interface prefixes (e.g. `veth`, `cali`, etc.).
// The first found global-scope IPv4 and IPv6 addresses are returned.
//
// Parameters:
//   - interfaces: List of proto InterfaceStat objects
//   - logger: Logger for debug tracing
//
// Returns:
//   - ipv4Addr: Primary IPv4 address as string (empty if not found)
//   - ipv6Addr: Primary IPv6 address as string (empty if not found)
func buildPrimaryIPs(
	interfaces []*gen.InterfaceStat,
	logger *zap.Logger,
) (string, string) {
	logger.Debug("Start buildPrimaryIPs")

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

	logger.Debug("Finish buildPrimaryIPs")

	return ipv4Addr, ipv6Addr
}
