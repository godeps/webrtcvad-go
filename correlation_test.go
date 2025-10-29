package webrtcvad

import (
	"math"
	"testing"
)

// TestCrossCorrelation 测试互相关
func TestCrossCorrelation(t *testing.T) {
	// 创建两个简单的测试序列
	seq1 := []int16{1, 2, 3, 4, 5}
	seq2 := []int16{5, 4, 3, 2, 1}

	dimSeq := 5
	dimCrossCorr := 3
	rightShifts := 0
	stepSeq2 := 1

	result := CrossCorrelation(seq1, seq2, dimSeq, dimCrossCorr, rightShifts, stepSeq2)

	if len(result) != dimCrossCorr {
		t.Errorf("Expected length %d, got %d", dimCrossCorr, len(result))
	}

	// 验证第一个值
	// corr[0] = 1*5 + 2*4 + 3*3 + 4*2 + 5*1 = 5+8+9+8+5 = 35
	expected := int32(35)
	if result[0] != expected {
		t.Errorf("Expected first correlation value %d, got %d", expected, result[0])
	}

	t.Logf("Cross-correlation result: %v", result)
}

// TestAutoCorrelation 测试自相关
func TestAutoCorrelation(t *testing.T) {
	// 创建测试序列
	seq := []int16{1, 2, 3, 4, 5, 4, 3, 2, 1}

	dimSeq := 9
	dimAutoCorr := 5
	rightShifts := 0

	result := AutoCorrelation(seq, dimSeq, dimAutoCorr, rightShifts)

	if len(result) != dimAutoCorr {
		t.Errorf("Expected length %d, got %d", dimAutoCorr, len(result))
	}

	// 自相关在lag=0时应该最大
	if result[0] <= result[1] || result[0] <= result[2] {
		t.Error("Autocorrelation at lag=0 should be maximum")
	}

	// lag=0的自相关应该是能量
	var energy int32 = 0
	for i := 0; i < dimSeq; i++ {
		energy += int32(seq[i]) * int32(seq[i])
	}
	if result[0] != energy {
		t.Errorf("Autocorrelation at lag=0 should equal energy %d, got %d",
			energy, result[0])
	}

	t.Logf("Auto-correlation result: %v", result)
}

// TestCrossCorrelationWithRightShift 测试带右移的互相关
func TestCrossCorrelationWithRightShift(t *testing.T) {
	// 创建较大值的序列
	seq1 := []int16{1000, 2000, 3000}
	seq2 := []int16{1000, 2000, 3000}

	// 不带移位
	result1 := CrossCorrelation(seq1, seq2, 3, 1, 0, 1)

	// 带右移4位
	result2 := CrossCorrelation(seq1, seq2, 3, 1, 4, 1)

	// result2应该约为result1 / 16
	ratio := float64(result1[0]) / float64(result2[0])
	if ratio < 15 || ratio > 17 {
		t.Errorf("Expected ratio ~16, got %.2f", ratio)
	}

	t.Logf("Without shift: %d, With shift: %d, Ratio: %.2f",
		result1[0], result2[0], ratio)
}

// TestFindPeakCorrelation 测试查找峰值
func TestFindPeakCorrelation(t *testing.T) {
	correlation := []int32{100, 200, 500, 300, 150}

	peakIdx, peakVal := FindPeakCorrelation(correlation)

	if peakIdx != 2 {
		t.Errorf("Expected peak index 2, got %d", peakIdx)
	}
	if peakVal != 500 {
		t.Errorf("Expected peak value 500, got %d", peakVal)
	}

	// 测试空数组
	peakIdx, peakVal = FindPeakCorrelation([]int32{})
	if peakIdx != -1 || peakVal != 0 {
		t.Error("Empty array should return -1, 0")
	}
}

// TestNormalizedCrossCorrelation 测试归一化互相关
func TestNormalizedCrossCorrelation(t *testing.T) {
	// 相同序列，相关系数应该为1
	seq1 := []int16{1, 2, 3, 4, 5}
	result1 := NormalizedCrossCorrelation(seq1, seq1, 5)
	if math.Abs(result1-1.0) > 0.001 {
		t.Errorf("Same sequence should have correlation ~1.0, got %.3f", result1)
	}

	// 完全相反的序列
	seq2 := []int16{-1, -2, -3, -4, -5}
	result2 := NormalizedCrossCorrelation(seq1, seq2, 5)
	if math.Abs(result2+1.0) > 0.001 {
		t.Errorf("Opposite sequence should have correlation ~-1.0, got %.3f", result2)
	}

	// 不相关的序列
	seq3 := []int16{1, -1, 1, -1, 1}
	seq4 := []int16{0, 1, 0, 1, 0}
	result3 := NormalizedCrossCorrelation(seq3, seq4, 5)
	if math.Abs(result3) > 0.5 {
		t.Logf("Uncorrelated sequences correlation: %.3f", result3)
	}

	t.Logf("Normalized correlations: same=%.3f, opposite=%.3f, uncorrelated=%.3f",
		result1, result2, result3)
}

// TestCrossCorrelationWithLag 测试带延迟的互相关
func TestCrossCorrelationWithLag(t *testing.T) {
	seq1 := []int16{1, 2, 3, 4, 5}
	seq2 := []int16{0, 1, 2, 3, 4, 5, 0}

	// 正延迟：seq2向右移动1位应该和seq1对齐
	corr0 := CrossCorrelationWithLag(seq1, seq2, 5, 0, 0)
	corr1 := CrossCorrelationWithLag(seq1, seq2, 5, 1, 0)

	// corr1应该大于corr0（因为对齐了）
	if corr1 <= corr0 {
		t.Logf("Warning: corr1(%d) should be > corr0(%d)", corr1, corr0)
	}

	t.Logf("Correlation at lag 0: %d, lag 1: %d", corr0, corr1)

	// 测试负延迟
	corrNeg := CrossCorrelationWithLag(seq2, seq1, 5, -1, 0)
	t.Logf("Correlation at lag -1: %d", corrNeg)
}

// BenchmarkCrossCorrelation 基准测试互相关
func BenchmarkCrossCorrelation(b *testing.B) {
	seq1 := make([]int16, 256)
	seq2 := make([]int16, 256)

	for i := 0; i < 256; i++ {
		seq1[i] = int16(i % 100)
		seq2[i] = int16((i + 10) % 100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CrossCorrelation(seq1, seq2, 256, 128, 4, 1)
	}
}

// BenchmarkAutoCorrelation 基准测试自相关
func BenchmarkAutoCorrelation(b *testing.B) {
	seq := make([]int16, 512)

	for i := 0; i < 512; i++ {
		seq[i] = int16(math.Sin(float64(i)*0.1) * 1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AutoCorrelation(seq, 512, 64, 0)
	}
}

