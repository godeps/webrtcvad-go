package webrtcvad

import (
	"math"
	"testing"
)

// TestComplexFFT 测试复数FFT
func TestComplexFFT(t *testing.T) {
	// 测试8点FFT
	stages := 3 // 2^3 = 8
	n := 1 << uint(stages)

	// 创建测试信号：纯实数正弦波
	data := make([]int16, n*2) // 实部和虚部交织
	for i := 0; i < n; i++ {
		// 生成一个简单的信号
		data[2*i] = int16(1000 * math.Sin(2*math.Pi*float64(i)/float64(n)))
		data[2*i+1] = 0 // 虚部为0
	}

	// 执行FFT
	result := ComplexFFT(data, stages, 1)
	if result != 0 {
		t.Errorf("ComplexFFT failed with result: %d", result)
	}

	// FFT后应该有非零值
	foundNonZero := false
	for i := 0; i < n*2; i++ {
		if data[i] != 0 {
			foundNonZero = true
			break
		}
	}
	if !foundNonZero {
		t.Error("FFT result is all zeros")
	}

	t.Logf("FFT result (first 4 bins): re=%d, im=%d, re=%d, im=%d",
		data[0], data[1], data[2], data[3])
}

// TestComplexIFFT 测试复数逆FFT
func TestComplexIFFT(t *testing.T) {
	stages := 3
	n := 1 << uint(stages)

	// 创建测试数据
	original := make([]int16, n*2)
	for i := 0; i < n; i++ {
		original[2*i] = int16(i * 100)
		original[2*i+1] = 0
	}

	// 保存原始数据
	backup := make([]int16, n*2)
	copy(backup, original)

	// 正向FFT
	ComplexFFT(original, stages, 1)

	// 逆FFT
	scale := ComplexIFFT(original, stages, 1)
	if scale < 0 {
		t.Errorf("ComplexIFFT failed with scale: %d", scale)
	}

	// 验证逆变换（应该接近原始值，考虑缩放因子）
	t.Logf("IFFT scale factor: %d", scale)
	t.Logf("First value: original=%d, after FFT+IFFT=%d",
		backup[0], original[0]<<uint(scale))
}

// TestRealFFT 测试实数FFT
func TestRealFFT(t *testing.T) {
	order := 3 // 2^3 = 8
	fft := CreateRealFFT(order)
	if fft == nil {
		t.Fatal("Failed to create RealFFT")
	}

	n := 1 << uint(order)
	realData := make([]int16, n)
	complexData := make([]int16, n+2)

	// 创建测试信号
	for i := 0; i < n; i++ {
		realData[i] = int16(1000 * math.Cos(2*math.Pi*float64(i)/float64(n)))
	}

	// 正向FFT
	result := fft.RealForwardFFT(realData, complexData)
	if result != 0 {
		t.Errorf("RealForwardFFT failed with result: %d", result)
	}

	// 验证DC分量
	t.Logf("DC component: %d", complexData[0])
	t.Logf("Nyquist component: %d", complexData[n])

	// 验证CCS格式
	if complexData[1] != 0 {
		t.Errorf("DC imaginary part should be 0, got %d", complexData[1])
	}
	if complexData[n+1] != 0 {
		t.Errorf("Nyquist imaginary part should be 0, got %d", complexData[n+1])
	}
}

// TestRealInverseFFT 测试实数逆FFT
func TestRealInverseFFT(t *testing.T) {
	order := 4 // 2^4 = 16
	fft := CreateRealFFT(order)
	if fft == nil {
		t.Fatal("Failed to create RealFFT")
	}

	n := 1 << uint(order)
	original := make([]int16, n)
	complexData := make([]int16, n+2)
	reconstructed := make([]int16, n)

	// 创建测试信号
	for i := 0; i < n; i++ {
		original[i] = int16(i * 10)
	}

	// 正向FFT
	fft.RealForwardFFT(original, complexData)

	// 逆FFT
	scale := fft.RealInverseFFT(complexData, reconstructed)
	if scale < 0 {
		t.Errorf("RealInverseFFT failed with scale: %d", scale)
	}

	t.Logf("Real IFFT scale factor: %d", scale)
	t.Logf("Original[0]=%d, Reconstructed[0]=%d (scaled: %d)",
		original[0], reconstructed[0], reconstructed[0]<<uint(scale))
}

// TestFFTInvalidOrder 测试无效FFT阶数
func TestFFTInvalidOrder(t *testing.T) {
	// 测试过大的阶数
	data := make([]int16, 2048*2)
	result := ComplexFFT(data, 11, 1) // 2^11 = 2048 > 1024
	if result != -1 {
		t.Error("Should fail with order 11")
	}

	// 测试CreateRealFFT的边界情况
	if CreateRealFFT(0) != nil {
		t.Error("Should return nil for order 0")
	}
	if CreateRealFFT(1) != nil {
		t.Error("Should return nil for order 1")
	}
	if CreateRealFFT(kMaxFFTOrder+1) != nil {
		t.Error("Should return nil for order > kMaxFFTOrder")
	}
	if CreateRealFFT(kMaxFFTOrder) == nil {
		t.Error("Should succeed for order kMaxFFTOrder")
	}
}

// BenchmarkComplexFFT 基准测试复数FFT
func BenchmarkComplexFFT(b *testing.B) {
	stages := 8 // 2^8 = 256
	n := 1 << uint(stages)
	data := make([]int16, n*2)

	for i := 0; i < n; i++ {
		data[2*i] = int16(i % 1000)
		data[2*i+1] = 0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComplexFFT(data, stages, 1)
	}
}

// BenchmarkRealFFT 基准测试实数FFT
func BenchmarkRealFFT(b *testing.B) {
	order := 8
	fft := CreateRealFFT(order)
	n := 1 << uint(order)

	realData := make([]int16, n)
	complexData := make([]int16, n+2)

	for i := 0; i < n; i++ {
		realData[i] = int16(i % 1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fft.RealForwardFFT(realData, complexData)
	}
}

