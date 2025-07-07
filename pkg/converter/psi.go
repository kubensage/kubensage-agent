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
