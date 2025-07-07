package converter

import (
	"github.com/jinzhu/copier"
	"github.com/kubensage/kubensage-agent/pkg/metrics"
	pb "github.com/kubensage/kubensage-agent/proto/gen"
)

func ConvertToProto(m *metrics.Metrics) (*pb.Metrics, error) {
	var out pb.Metrics
	if err := copier.Copy(&out, m); err != nil {
		return nil, err
	}
	return &out, nil
}

/*func ConvertToProto(m *metrics.Metrics) (*pb.Metrics, error) {
	psiCpuSomeData := &pb.PsiData{
		Total:  m.NodeMetrics.PsiCpuMetrics.Some.Total,
		Avg10:  m.NodeMetrics.PsiCpuMetrics.Some.Avg10,
		Avg60:  m.NodeMetrics.PsiCpuMetrics.Some.Avg60,
		Avg300: m.NodeMetrics.PsiCpuMetrics.Some.Avg300,
	}
	psiCpuFullData := &pb.PsiData{
		Total:  m.NodeMetrics.PsiCpuMetrics.Full.Total,
		Avg10:  m.NodeMetrics.PsiCpuMetrics.Full.Avg10,
		Avg60:  m.NodeMetrics.PsiCpuMetrics.Full.Avg60,
		Avg300: m.NodeMetrics.PsiCpuMetrics.Full.Avg300,
	}
	psiCpuMetrics := &pb.PsiMetrics{
		Some: psiCpuSomeData,
		Full: psiCpuFullData,
	}
	return &pb.Metrics{
		NodeMetrics: &pb.NodeMetrics{
			Hostname:        m.NodeMetrics.Hostname,
			Uptime:          m.NodeMetrics.Uptime,
			BootTime:        m.NodeMetrics.BootTime,
			Procs:           m.NodeMetrics.Procs,
			Os:              m.NodeMetrics.OS,
			Platform:        m.NodeMetrics.Platform,
			PlatformFamily:  m.NodeMetrics.PlatformFamily,
			PlatformVersion: m.NodeMetrics.PlatformVersion,
			KernelVersion:   m.NodeMetrics.KernelVersion,
			KernelArch:      m.NodeMetrics.KernelArch,
			HostId:          m.NodeMetrics.HostID,
			CpuModel:        m.NodeMetrics.CPUModel,
			CpuCores:        m.NodeMetrics.CPUCores,
			CpuUsagePercent: m.NodeMetrics.CPUUsagePercent,
			TotalMemory:     m.NodeMetrics.TotalMemory,
			FreeMemory:      m.NodeMetrics.FreeMemory,
			UsedMemory:      m.NodeMetrics.UsedMemory,
			MemoryUsedPerc:  m.NodeMetrics.MemoryUsedPerc,
			PsiCpuMetrics:   psiCpuMetrics,
			/*PsiMemoryMetrics:  nil,
			PsiIoMetrics:      nil,
			NetworkInterfaces: nil,
		},
		PodMetrics: []*pb.PodMetrics{},
	}, nil
}*/
