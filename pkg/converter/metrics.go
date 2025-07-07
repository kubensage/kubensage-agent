package converter

import (
	"github.com/kubensage/kubensage-agent/pkg/metrics"
	pb "github.com/kubensage/kubensage-agent/proto/gen"
)

func ConvertToProto(m *metrics.Metrics) (*pb.Metrics, error) {
	node := m.NodeMetrics

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
			PsiCpuMetrics:     convertPsiMetrics(node.PsiCpuMetrics),
			PsiMemoryMetrics:  convertPsiMetrics(node.PsiMemoryMetrics),
			PsiIoMetrics:      convertPsiMetrics(node.PsiIoMetrics),
			NetworkInterfaces: convertNetworkInterfaces(node.NetworkInterfaces),
		},
		PodMetrics: convertPodMetrics(m.PodMetrics),
	}, nil
}
