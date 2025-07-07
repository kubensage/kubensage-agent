package converter

import (
	pb "github.com/kubensage/kubensage-agent/proto/gen"
	"github.com/shirou/gopsutil/v3/net"
)

func convertNetworkInterfaces(interfaces []net.InterfaceStat) []*pb.InterfaceStat {
	var result []*pb.InterfaceStat

	for _, iface := range interfaces {
		var addresses []*pb.InterfaceAddr

		for _, a := range iface.Addrs {
			addresses = append(addresses, &pb.InterfaceAddr{Addr: a.String()})
		}

		result = append(result, &pb.InterfaceStat{
			Index:        int32(iface.Index),
			Mtu:          int32(iface.MTU),
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr,
			Flags:        iface.Flags,
			Addrs:        addresses,
		})
	}

	return result
}
