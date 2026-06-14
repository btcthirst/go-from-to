package performance_test

import (
	"fmt"
	"testing"

	perf "go-from-to/patterns/03-performance"
)

// ================================================================
// sync.Pool
// ================================================================

func BenchmarkWithPool(b *testing.B) {
	data := []byte("hello benchmark data")
	for b.Loop() {
		perf.WithPool(data)
	}
}

func BenchmarkWithoutPool(b *testing.B) {
	data := []byte("hello benchmark data")
	for b.Loop() {
		perf.WithoutPool(data)
	}
}

// ================================================================
// Slice pre-allocation
// ================================================================

func BenchmarkSliceNoPrealloc(b *testing.B) {
	for b.Loop() {
		perf.SliceNoPrealloc(1000)
	}
}

func BenchmarkSliceWithPrealloc(b *testing.B) {
	for b.Loop() {
		perf.SliceWithPrealloc(1000)
	}
}

func BenchmarkSliceDirectIndex(b *testing.B) {
	for b.Loop() {
		perf.SliceDirectIndex(1000)
	}
}

// ================================================================
// String concatenation
// ================================================================

func BenchmarkConcatNaive(b *testing.B) {
	for b.Loop() {
		perf.ConcatNaive(100)
	}
}

func BenchmarkConcatBuilder(b *testing.B) {
	for b.Loop() {
		perf.ConcatBuilder(100)
	}
}

func BenchmarkConcatJoin(b *testing.B) {
	parts := make([]string, 100)
	for i := range parts {
		parts[i] = fmt.Sprintf("item%d", i)
	}
	b.ResetTimer()
	for b.Loop() {
		perf.ConcatJoin(parts)
	}
}

// ================================================================
// Map pre-allocation
// ================================================================

func BenchmarkMapNoPrealloc(b *testing.B) {
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key%d", i)
	}
	b.ResetTimer()
	for b.Loop() {
		perf.MapNoPrealloc(keys)
	}
}

func BenchmarkMapWithPrealloc(b *testing.B) {
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key%d", i)
	}
	b.ResetTimer()
	for b.Loop() {
		perf.MapWithPrealloc(keys)
	}
}

// ================================================================
// Interface boxing
// ================================================================

func BenchmarkBoxedSum(b *testing.B) {
	values := make([]any, 1000)
	for i := range values {
		values[i] = float64(i)
	}
	b.ResetTimer()
	for b.Loop() {
		perf.BoxedSum(values)
	}
}

func BenchmarkTypedSum(b *testing.B) {
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i)
	}
	b.ResetTimer()
	for b.Loop() {
		perf.TypedSum(values)
	}
}
