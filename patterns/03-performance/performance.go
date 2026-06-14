package performance

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// ================================================================
// sync.Pool — повторне використання об'єктів
// ================================================================

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// WithPool використовує Pool: без зайвих алокацій при повторних викликах
func WithPool(data []byte) string {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	buf.Write(data)
	return buf.String()
}

// WithoutPool: кожен виклик — нова алокація
func WithoutPool(data []byte) string {
	var buf bytes.Buffer
	buf.Write(data)
	return buf.String()
}

// ================================================================
// Slice pre-allocation
// ================================================================

// SliceNoPrealloc: append без capacity — багато реалокацій
func SliceNoPrealloc(n int) []int {
	var result []int
	for i := range n {
		result = append(result, i*2)
	}
	return result
}

// SliceWithPrealloc: одна алокація
func SliceWithPrealloc(n int) []int {
	result := make([]int, 0, n)
	for i := range n {
		result = append(result, i*2)
	}
	return result
}

// SliceDirectIndex: нуль реалокацій, нуль зростання
func SliceDirectIndex(n int) []int {
	result := make([]int, n)
	for i := range n {
		result[i] = i * 2
	}
	return result
}

// ================================================================
// strings.Builder vs конкатенація
// ================================================================

// ConcatNaive: O(n²) — кожен += копіює весь рядок
func ConcatNaive(n int) string {
	s := ""
	for i := range n {
		s += fmt.Sprintf("item%d,", i)
	}
	return s
}

// ConcatBuilder: O(n) — один буфер
func ConcatBuilder(n int) string {
	var sb strings.Builder
	sb.Grow(n * 8) // підказка: уникнути внутрішніх реалокацій Builder
	for i := range n {
		fmt.Fprintf(&sb, "item%d,", i)
	}
	return sb.String()
}

// ConcatJoin: найшвидше коли є slice рядків
func ConcatJoin(parts []string) string {
	return strings.Join(parts, ",")
}

// ================================================================
// Map pre-allocation
// ================================================================

func MapNoPrealloc(keys []string) map[string]int {
	m := make(map[string]int)
	for i, k := range keys {
		m[k] = i
	}
	return m
}

func MapWithPrealloc(keys []string) map[string]int {
	m := make(map[string]int, len(keys)) // підказуємо розмір
	for i, k := range keys {
		m[k] = i
	}
	return m
}

// ================================================================
// Interface boxing
// ================================================================

// BoxedSum: кожен елемент упаковується в interface{} → heap
func BoxedSum(values []any) float64 {
	var sum float64
	for _, v := range values {
		if f, ok := v.(float64); ok {
			sum += f
		}
	}
	return sum
}

// TypedSum: без боксингу
func TypedSum(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum
}
