package metrics

import "github.com/kubensage/kubensage-agent/proto/gen"

type RingBuffer struct {
	data     []*gen.Metrics
	capacity int
	start    int
	size     int
}

func NewMetricsRingBuffer(cap int) *RingBuffer {
	return &RingBuffer{
		data:     make([]*gen.Metrics, cap),
		capacity: cap,
	}
}

func (b *RingBuffer) Add(m *gen.Metrics) {
	idx := (b.start + b.size) % b.capacity
	b.data[idx] = m
	if b.size < b.capacity {
		b.size++
	} else {
		b.start = (b.start + 1) % b.capacity // overwrite oldest
	}
}

func (b *RingBuffer) Pop() *gen.Metrics {
	if b.size == 0 {
		return nil
	}
	m := b.data[b.start]
	b.data[b.start] = nil
	b.start = (b.start + 1) % b.capacity
	b.size--
	return m
}

func (b *RingBuffer) Len() int {
	return b.size
}
