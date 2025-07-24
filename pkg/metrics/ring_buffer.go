package metrics

import "github.com/kubensage/kubensage-agent/proto/gen"

type RingBuffer struct {
	data     []*gen.Metrics
	capacity int
	start    int
	size     int
}

// NewMetricsRingBuffer creates and returns a new RingBuffer with the specified capacity.
// The buffer stores pointers to gen.Metrics and operates in a circular manner.
// When full, new entries overwrite the oldest ones.
func NewMetricsRingBuffer(cap int) *RingBuffer {
	return &RingBuffer{
		data:     make([]*gen.Metrics, cap),
		capacity: cap,
	}
}

// Add inserts a new *gen.Metrics item into the ring buffer.
// If the buffer is full, it overwrites the oldest entry.
func (b *RingBuffer) Add(m *gen.Metrics) {
	idx := (b.start + b.size) % b.capacity
	b.data[idx] = m
	if b.size < b.capacity {
		b.size++
	} else {
		b.start = (b.start + 1) % b.capacity // overwrite oldest
	}
}

// Pop removes and returns the oldest *gen.Metrics item from the buffer.
// If the buffer is empty, it returns nil.
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

// Readd reinserts an element that was just removed via Pop()
// back into the position it was originally extracted from, if space is available.
//
// Returns true if the operation is successful, or false if the buffer is full.
func (b *RingBuffer) Readd(m *gen.Metrics) bool {
	if b.size == b.capacity {
		return false // buffer is full, cannot reinsert
	}

	// Move the start index back (circularly)
	b.start = (b.start - 1 + b.capacity) % b.capacity
	b.data[b.start] = m
	b.size++
	return true
}

// Len returns the current number of elements in the ring buffer.
func (b *RingBuffer) Len() int {
	return b.size
}
