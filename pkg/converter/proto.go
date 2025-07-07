package converter

import (
	"github.com/kubensage/kubensage-agent/pkg/metrics"
	pb "github.com/kubensage/kubensage-agent/proto/gen"
)

func convertPsiData(data metrics.PsiData) *pb.PsiData {
	return &pb.PsiData{
		Total:  data.Total,
		Avg10:  data.Avg10,
		Avg60:  data.Avg60,
		Avg300: data.Avg300,
	}
}

func convertPsiMetrics(m metrics.PsiMetrics) *pb.PsiMetrics {
	return &pb.PsiMetrics{
		Some: convertPsiData(m.Some),
		Full: convertPsiData(m.Full),
	}
}

func ConvertToProto(m *metrics.Metrics) (*pb.Metrics, error) {
	node := m.NodeMetrics

	// Convert PSI
	psiCpuMetrics := convertPsiMetrics(node.PsiCpuMetrics)
	psiMemoryMetrics := convertPsiMetrics(node.PsiMemoryMetrics)
	psiIoMetrics := convertPsiMetrics(node.PsiIoMetrics)

	// Convert network interfaces
	var networkInterfaces []*pb.InterfaceStat
	for _, iface := range node.NetworkInterfaces {
		var ifaceAddresses []*pb.InterfaceAddr
		for _, ifaceAddr := range iface.Addrs {
			ifaceAddresses = append(ifaceAddresses, &pb.InterfaceAddr{Addr: ifaceAddr.String()})
		}

		networkInterfaces = append(networkInterfaces, &pb.InterfaceStat{
			Index:        int32(iface.Index),
			Mtu:          int32(iface.MTU),
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr,
			Flags:        iface.Flags,
			Addrs:        ifaceAddresses,
		})
	}

	var podsMetrics []*pb.PodMetrics

	for _, pm := range m.PodMetrics {
		var containerMetrics []*pb.ContainerMetrics

		for _, cm := range pm.ContainerMetrics {

			cpuMetrics := &pb.CpuMetrics{
				Timestamp:            uint64(cm.CpuMetrics.Timestamp),
				UsageNanoCores:       uint64(cm.CpuMetrics.UsageNanoCores),
				UsageCoreNanoSeconds: uint64(cm.CpuMetrics.UsageCoreNanoSeconds),
			}

			memoryMetrics := &pb.MemoryMetrics{
				Timestamp:       uint64(cm.MemoryMetrics.Timestamp),
				WorkingSetBytes: uint64(cm.MemoryMetrics.WorkingSetBytes),
				AvailableBytes:  uint64(cm.MemoryMetrics.AvailableBytes),
				UsageBytes:      uint64(cm.MemoryMetrics.UsageBytes),
				RssBytes:        uint64(cm.MemoryMetrics.RssBytes),
				PageFaults:      uint64(cm.MemoryMetrics.PageFaults),
				MajorPageFaults: uint64(cm.MemoryMetrics.MajorPageFaults),
			}

			fileSystemMetrics := &pb.FileSystemMetrics{
				Timestamp:  uint64(cm.FileSystemMetrics.Timestamp),
				Mountpoint: cm.FileSystemMetrics.Mountpoint,
				UsedBytes:  uint64(cm.FileSystemMetrics.UsedBytes),
				InodesUsed: uint64(cm.FileSystemMetrics.InodesUsed),
			}

			swapMetrics := &pb.SwapMetrics{
				Timestamp:      uint64(cm.SwapMetrics.Timestamp),
				AvailableBytes: uint64(cm.SwapMetrics.SwapAvailableBytes),
				UsageBytes:     uint64(cm.SwapMetrics.SwapUsageBytes),
			}

			containerMetric := &pb.ContainerMetrics{
				Id:                cm.Id,
				Name:              cm.Name,
				Image:             cm.Image,
				CreatedAt:         cm.CreatedAt,
				State:             cm.State,
				Attempt:           cm.Attempt,
				CpuMetrics:        cpuMetrics,
				MemoryMetrics:     memoryMetrics,
				FileSystemMetrics: fileSystemMetrics,
				SwapMetrics:       swapMetrics,
			}

			containerMetrics = append(containerMetrics, containerMetric)
		}

		podMetrics := pb.PodMetrics{
			Id:               pm.Id,
			Uid:              pm.Uid,
			Name:             pm.Name,
			Namespace:        pm.Namespace,
			CreatedAt:        pm.CreatedAt,
			State:            pm.State,
			Attempt:          pm.Attempt,
			ContainerMetrics: containerMetrics,
		}

		podsMetrics = append(podsMetrics, &podMetrics)
	}

	// Compose result
	return &pb.Metrics{
		NodeMetrics: &pb.NodeMetrics{
			Hostname:          node.Hostname,
			Uptime:            node.Uptime,
			BootTime:          node.BootTime,
			Procs:             node.Procs,
			Os:                node.OS,
			Platform:          node.Platform,
			PlatformFamily:    node.PlatformFamily,
			PlatformVersion:   node.PlatformVersion,
			KernelVersion:     node.KernelVersion,
			KernelArch:        node.KernelArch,
			HostId:            node.HostID,
			CpuModel:          node.CPUModel,
			CpuCores:          node.CPUCores,
			CpuUsagePercent:   node.CPUUsagePercent,
			TotalMemory:       node.TotalMemory,
			FreeMemory:        node.FreeMemory,
			UsedMemory:        node.UsedMemory,
			MemoryUsedPerc:    node.MemoryUsedPerc,
			PsiCpuMetrics:     psiCpuMetrics,
			PsiMemoryMetrics:  psiMemoryMetrics,
			PsiIoMetrics:      psiIoMetrics,
			NetworkInterfaces: networkInterfaces,
		},
		PodMetrics: podsMetrics,
	}, nil
}
