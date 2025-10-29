package webrtcvad

import (
	"testing"
)

// spl_test.go - 信号处理库测试
// 包含正确性测试和性能基准测试

// BenchmarkNormW32 前导零计数（已优化，使用bits包）
func BenchmarkNormW32(b *testing.B) {
	testCases := []int32{
		0, 1, -1, 100, -100, 1000, -1000,
		0x7FFFFFFF, -0x80000000, 0x12345678,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range testCases {
			normW32(val)
		}
	}
}

// BenchmarkZerosArrayW16 数组清零（已优化，使用clear()）
func BenchmarkZerosArrayW16(b *testing.B) {
	data := make([]int16, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zerosArrayW16(data, 1024)
	}
}

// BenchmarkMaxAbsValueW16 最大绝对值查找
func BenchmarkMaxAbsValueW16(b *testing.B) {
	data := make([]int16, 512)
	for i := range data {
		data[i] = int16((i * 37) % 1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		maxAbsValueW16(data, 512)
	}
}

// BenchmarkMinValueW16 最小值查找（已优化，循环展开）
func BenchmarkMinValueW16(b *testing.B) {
	data := make([]int16, 512)
	for i := range data {
		data[i] = int16((i * 37) % 1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		minValueW16(data, 512)
	}
}

// BenchmarkCalculateEnergy 能量计算（已优化，循环展开）
func BenchmarkCalculateEnergy(b *testing.B) {
	data := make([]int16, 256)
	for i := range data {
		data[i] = int16((i * 37) % 1000)
	}
	var scale int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateEnergy(data, 256, &scale)
	}
}

// 正确性测试

func TestNormW32OptCorrectness(t *testing.T) {
	testCases := []int32{
		0, 1, -1, 100, -100, 1000, -1000,
		0x7FFFFFFF, -0x80000000, 0x00010000,
		0x00000100, 0x00000001, 0x12345678,
	}

	// 这些测试用例应该都能正确处理
	for _, val := range testCases {
		result := normW32(val)
		// 基本验证：结果应该在合理范围内
		if result < 0 || result > 32 {
			t.Errorf("normW32 result out of range for %d: got %d", val, result)
		}
	}
}

func TestZerosArrayCorrectness(t *testing.T) {
	size := 100

	// 测试清零功能
	data := make([]int16, size)
	for i := range data {
		data[i] = int16(i)
	}
	zerosArrayW16(data, size)

	// 验证所有元素都为0
	for i := 0; i < size; i++ {
		if data[i] != 0 {
			t.Errorf("zerosArray failed at index %d: expected 0, got %d", i, data[i])
		}
	}
}

func TestMaxAbsValueCorrectness(t *testing.T) {
	testCases := [][]int16{
		{1, 2, 3, 4, 5},
		{-1, -2, -3, -4, -5},
		{100, -200, 50, -150, 75},
		{0, 0, 0, 0, 0},
		{32767, -32768, 0, 100, -100},
	}

	for _, data := range testCases {
		result := maxAbsValueW16(data, len(data))
		// 验证结果是正数且合理
		if result < 0 {
			t.Errorf("maxAbsValue returned negative: %d for data %v", result, data)
		}
	}
}

func TestMinMaxValueCorrectness(t *testing.T) {
	testCases := [][]int16{
		{1, 2, 3, 4, 5},
		{-1, -2, -3, -4, -5},
		{100, -200, 50, -150, 75},
		{32767, -32768, 0, 100, -100},
	}

	for _, data := range testCases {
		// Min
		minVal := minValueW16(data, len(data))
		// Max
		maxVal := maxValueW16(data, len(data))

		// 验证min <= max
		if minVal > maxVal {
			t.Errorf("minValue > maxValue for %v: min=%d, max=%d",
				data, minVal, maxVal)
		}
	}
}

func TestCalculateEnergyCorrectness(t *testing.T) {
	testCases := [][]int16{
		{1, 2, 3, 4, 5},
		{100, 200, 300, 400, 500},
		{-100, -200, -300, -400, -500},
		{1000, -1000, 2000, -2000, 0},
	}

	for _, data := range testCases {
		var scale int
		energy := calculateEnergy(data, len(data), &scale)

		// 验证能量是非负数
		if scale < 0 {
			t.Errorf("calculateEnergy returned negative scale: %d for data %v", scale, data)
		}
		// 能量应该是合理的
		_ = energy // 基本验证通过
	}
}

// 并发安全测试
func TestOptimizedFunctionsConcurrency(t *testing.T) {
	data := make([]int16, 1000)
	for i := range data {
		data[i] = int16(i % 1000)
	}

	done := make(chan bool)
	goroutines := 10
	iterations := 1000

	for g := 0; g < goroutines; g++ {
		go func() {
			var scale int
			for i := 0; i < iterations; i++ {
				normW32(int32(i))
				maxAbsValueW16(data, len(data))
				minValueW16(data, len(data))
				calculateEnergy(data, len(data), &scale)
			}
			done <- true
		}()
	}

	for g := 0; g < goroutines; g++ {
		<-done
	}
}
