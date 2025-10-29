package webrtcvad

import (
	"testing"
)

// TestAbsGeneric 测试泛型绝对值
func TestAbsGeneric(t *testing.T) {
	// 测试int16
	if Abs[int16](-100) != 100 {
		t.Error("Abs[int16](-100) 失败")
	}
	if Abs[int16](100) != 100 {
		t.Error("Abs[int16](100) 失败")
	}

	// 测试int32
	if Abs[int32](-1000) != 1000 {
		t.Error("Abs[int32](-1000) 失败")
	}
	if Abs[int32](1000) != 1000 {
		t.Error("Abs[int32](1000) 失败")
	}

	// 测试int
	if Abs(-12345) != 12345 {
		t.Error("Abs(-12345) 失败")
	}

	// 测试零
	if Abs[int16](0) != 0 {
		t.Error("Abs[int16](0) 失败")
	}
}

// TestMinMaxGeneric 测试泛型最小/最大值
func TestMinMaxGeneric(t *testing.T) {
	// 测试Min
	if Min(3, 5) != 3 {
		t.Error("Min(3, 5) 失败")
	}
	if Min[int16](10, 20) != 10 {
		t.Error("Min[int16](10, 20) 失败")
	}
	if Min(-5, 5) != -5 {
		t.Error("Min(-5, 5) 失败")
	}

	// 测试Max
	if Max(3, 5) != 5 {
		t.Error("Max(3, 5) 失败")
	}
	if Max[int16](10, 20) != 20 {
		t.Error("Max[int16](10, 20) 失败")
	}
	if Max(-5, 5) != 5 {
		t.Error("Max(-5, 5) 失败")
	}
}

// TestClampGeneric 测试泛型限幅
func TestClampGeneric(t *testing.T) {
	// 正常范围内
	if Clamp(50, 0, 100) != 50 {
		t.Error("Clamp(50, 0, 100) 失败")
	}

	// 小于最小值
	if Clamp(-10, 0, 100) != 0 {
		t.Error("Clamp(-10, 0, 100) 失败")
	}

	// 大于最大值
	if Clamp(150, 0, 100) != 100 {
		t.Error("Clamp(150, 0, 100) 失败")
	}

	// 边界值
	if Clamp(0, 0, 100) != 0 {
		t.Error("Clamp(0, 0, 100) 失败")
	}
	if Clamp(100, 0, 100) != 100 {
		t.Error("Clamp(100, 0, 100) 失败")
	}
}

// TestMinMaxSlice 测试泛型切片最小/最大值
func TestMinMaxSlice(t *testing.T) {
	data := []int{3, 1, 4, 1, 5, 9, 2, 6}

	// 最小值
	min := MinSlice(data)
	if min != 1 {
		t.Errorf("MinSlice 失败: 期望1, 得到%d", min)
	}

	// 最大值
	max := MaxSlice(data)
	if max != 9 {
		t.Errorf("MaxSlice 失败: 期望9, 得到%d", max)
	}

	// 空切片
	emptyMin := MinSlice([]int{})
	if emptyMin != 0 {
		t.Errorf("MinSlice空切片 失败: 期望0, 得到%d", emptyMin)
	}

	// int16切片
	data16 := []int16{100, 200, 50, 150}
	min16 := MinSlice(data16)
	if min16 != 50 {
		t.Errorf("MinSlice[int16] 失败: 期望50, 得到%d", min16)
	}
}

// TestSumAverage 测试泛型求和和平均值
func TestSumAverage(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}

	// 求和
	sum := Sum(data)
	if sum != 15 {
		t.Errorf("Sum 失败: 期望15, 得到%d", sum)
	}

	// 平均值
	avg := Average(data)
	if avg != 3.0 {
		t.Errorf("Average 失败: 期望3.0, 得到%f", avg)
	}

	// float64切片
	dataFloat := []float64{1.5, 2.5, 3.5}
	sumFloat := Sum(dataFloat)
	if sumFloat != 7.5 {
		t.Errorf("Sum[float64] 失败: 期望7.5, 得到%f", sumFloat)
	}
}

// TestBackwardCompatibility 测试向后兼容的特化版本
func TestBackwardCompatibility(t *testing.T) {
	// AbsInt16
	if AbsInt16(-100) != 100 {
		t.Error("AbsInt16(-100) 失败")
	}

	// AbsInt32
	if AbsInt32(-1000) != 1000 {
		t.Error("AbsInt32(-1000) 失败")
	}

	// MinInt16
	if MinInt16(10, 20) != 10 {
		t.Error("MinInt16(10, 20) 失败")
	}

	// MaxInt16
	if MaxInt16(10, 20) != 20 {
		t.Error("MaxInt16(10, 20) 失败")
	}

	// MinInt32
	if MinInt32(100, 200) != 100 {
		t.Error("MinInt32(100, 200) 失败")
	}

	// MaxInt32
	if MaxInt32(100, 200) != 200 {
		t.Error("MaxInt32(100, 200) 失败")
	}
}

// TestGenericWithDifferentTypes 测试泛型支持多种类型
func TestGenericWithDifferentTypes(t *testing.T) {
	// int8
	if Abs[int8](-10) != 10 {
		t.Error("Abs[int8] 失败")
	}

	// int64
	if Abs[int64](-100000) != 100000 {
		t.Error("Abs[int64] 失败")
	}

	// uint (无符号，但Min/Max应该工作)
	if Min[uint](10, 20) != 10 {
		t.Error("Min[uint] 失败")
	}

	// float32
	if Min[float32](3.14, 2.71) != 2.71 {
		t.Error("Min[float32] 失败")
	}

	// string
	if Min("apple", "banana") != "apple" {
		t.Error("Min[string] 失败")
	}
}

// BenchmarkAbsGeneric Benchmark泛型绝对值
func BenchmarkAbsGeneric(b *testing.B) {
	val := int16(-100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Abs(val)
	}
}

// BenchmarkAbsInt16Specialized Benchmark特化版本
func BenchmarkAbsInt16Specialized(b *testing.B) {
	val := int16(-100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AbsInt16(val)
	}
}

// BenchmarkMinGeneric Benchmark泛型Min
func BenchmarkMinGeneric(b *testing.B) {
	a, b2 := int16(10), int16(20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Min(a, b2)
	}
}

// BenchmarkMinSlice Benchmark切片最小值
func BenchmarkMinSlice(b *testing.B) {
	data := make([]int16, 1000)
	for i := range data {
		data[i] = int16(i % 1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MinSlice(data)
	}
}
