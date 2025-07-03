package metrics

import runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"

type PsiMetrics struct {
	Total  uint64  `json:"total,omitempty"`
	Avg10  float64 `json:"avg10,omitempty"`
	Avg60  float64 `json:"avg60,omitempty"`
	Avg300 float64 `json:"avg300,omitempty"`
}

func extractPsiMetrics(psiData *runtimeapi.PsiData) PsiMetrics {
	if psiData == nil {
		return PsiMetrics{}
	}

	return PsiMetrics{
		Total:  psiData.Total,
		Avg10:  psiData.Avg10,
		Avg60:  psiData.Avg60,
		Avg300: psiData.Avg300,
	}
}

func SafePsiSomeCpuMetrics(stats *runtimeapi.ContainerStats) PsiMetrics {
	if stats.Cpu == nil || stats.Cpu.Psi == nil {
		return PsiMetrics{}
	}
	return extractPsiMetrics(stats.Cpu.Psi.Some)
}

func SafePsiFullCpuMetrics(stats *runtimeapi.ContainerStats) PsiMetrics {
	if stats.Cpu == nil || stats.Cpu.Psi == nil {
		return PsiMetrics{}
	}
	return extractPsiMetrics(stats.Cpu.Psi.Full)
}

func SafePsiSomeMemoryMetrics(stats *runtimeapi.ContainerStats) PsiMetrics {
	if stats.Memory == nil || stats.Memory.Psi == nil {
		return PsiMetrics{}
	}
	return extractPsiMetrics(stats.Memory.Psi.Some)
}

func SafePsiFullMemoryMetrics(stats *runtimeapi.ContainerStats) PsiMetrics {
	if stats.Memory == nil || stats.Memory.Psi == nil {
		return PsiMetrics{}
	}
	return extractPsiMetrics(stats.Memory.Psi.Full)
}

func SafePsiSomeIoMetrics(stats *runtimeapi.ContainerStats) PsiMetrics {
	if stats.Io == nil || stats.Io.Psi == nil {
		return PsiMetrics{}
	}
	return extractPsiMetrics(stats.Io.Psi.Some)
}

func SafePsiFullIoMetrics(stats *runtimeapi.ContainerStats) PsiMetrics {
	if stats.Io == nil || stats.Io.Psi == nil {
		return PsiMetrics{}
	}
	return extractPsiMetrics(stats.Io.Psi.Full)
}
