package webrtcvad

import (
	"math"
	"testing"
)

// TestLevinsonDurbin 测试Levinson-Durbin算法
func TestLevinsonDurbin(t *testing.T) {
	// 创建一个简单的自相关序列
	// 对应一个AR(2)过程
	autoCorr := []float64{1.0, 0.8, 0.5, 0.2}
	order := 2

	arCoeffs, predError := LevinsonDurbin(autoCorr, order)

	if len(arCoeffs) != order+1 {
		t.Errorf("Expected %d coefficients, got %d", order+1, len(arCoeffs))
	}

	// 第一个系数应该是1
	if arCoeffs[0] != 1.0 {
		t.Errorf("First coefficient should be 1.0, got %.3f", arCoeffs[0])
	}

	// 预测误差应该是正数
	if predError <= 0 {
		t.Errorf("Prediction error should be positive, got %.3f", predError)
	}

	// 预测误差应该小于原始方差
	if predError >= autoCorr[0] {
		t.Errorf("Prediction error %.3f should be < variance %.3f",
			predError, autoCorr[0])
	}

	t.Logf("AR coefficients: %v", arCoeffs)
	t.Logf("Prediction error: %.6f", predError)
}

// TestLPCAnalysis 测试LPC分析
func TestLPCAnalysis(t *testing.T) {
	// 创建一个简单的语音样本（正弦波）
	length := 256
	signal := make([]int16, length)

	for i := 0; i < length; i++ {
		signal[i] = int16(1000 * math.Sin(2*math.Pi*float64(i)/32.0))
	}

	order := 10
	lpcCoeffs, gain := LPCAnalysis(signal, length, order)

	if len(lpcCoeffs) != order+1 {
		t.Errorf("Expected %d LPC coefficients, got %d", order+1, len(lpcCoeffs))
	}

	// 第一个系数应该是1
	if lpcCoeffs[0] != 1.0 {
		t.Errorf("First LPC coefficient should be 1.0, got %.3f", lpcCoeffs[0])
	}

	// 增益应该是正数
	if gain <= 0 {
		t.Error("Gain should be positive")
	}

	t.Logf("LPC coefficients (first 5): [%.3f, %.3f, %.3f, %.3f, %.3f]",
		lpcCoeffs[0], lpcCoeffs[1], lpcCoeffs[2], lpcCoeffs[3], lpcCoeffs[4])
	t.Logf("Gain: %.6f", gain)
}

// TestLPCSynthesis 测试LPC合成
func TestLPCSynthesis(t *testing.T) {
	// 创建简单的激励信号
	length := 128
	excitation := make([]int16, length)
	output := make([]int16, length)

	// 脉冲激励
	excitation[0] = 1000
	for i := 1; i < length; i++ {
		excitation[i] = 0
	}

	// 使用简单的LPC系数
	lpcCoeffs := []float64{1.0, -0.9, 0.5}

	LPCSynthesis(excitation, lpcCoeffs, output)

	// 输出应该非零
	foundNonZero := false
	for i := 0; i < length; i++ {
		if output[i] != 0 {
			foundNonZero = true
			break
		}
	}

	if !foundNonZero {
		t.Error("LPC synthesis output is all zeros")
	}

	t.Logf("LPC synthesis output (first 10): %v", output[:10])
}

// TestARFilter 测试AR滤波器
func TestARFilter(t *testing.T) {
	order := 3
	ar := NewARFilter(order)

	if ar.order != order {
		t.Errorf("Expected order %d, got %d", order, ar.order)
	}

	// 设置系数
	coeffs := []float64{1.0, -0.8, 0.5, -0.2}
	ar.SetCoefficients(coeffs)

	// 创建输入信号
	length := 100
	input := make([]int16, length)
	output := make([]int16, length)

	for i := 0; i < length; i++ {
		input[i] = int16(500 * math.Sin(2*math.Pi*float64(i)/16.0))
	}

	// 应用滤波器
	ar.Filter(input, output)

	// 验证输出
	foundNonZero := false
	for i := 0; i < length; i++ {
		if output[i] != 0 {
			foundNonZero = true
			break
		}
	}

	if !foundNonZero {
		t.Error("AR filter output is all zeros")
	}

	t.Logf("AR filter output (first 10): %v", output[:10])
}

// TestComputeParcorCoefficients 测试PARCOR系数计算
func TestComputeParcorCoefficients(t *testing.T) {
	// 创建自相关序列
	autoCorr := []float64{1.0, 0.9, 0.7, 0.5, 0.3}
	order := 3

	parcor := ComputeParcorCoefficients(autoCorr, order)

	if len(parcor) != order {
		t.Errorf("Expected %d PARCOR coefficients, got %d", order, len(parcor))
	}

	// PARCOR系数应该在-1到1之间（稳定性条件）
	for i, k := range parcor {
		if k < -1.0 || k > 1.0 {
			t.Errorf("PARCOR[%d]=%.3f is outside [-1, 1]", i, k)
		}
	}

	t.Logf("PARCOR coefficients: %v", parcor)
}

// TestARFilterInt16 测试定点AR滤波器
func TestARFilterInt16(t *testing.T) {
	length := 64
	order := 4

	input := make([]int16, length)
	output := make([]int16, length)
	coeffs := make([]int16, order)
	state := make([]int16, order)

	// 创建测试信号
	for i := 0; i < length; i++ {
		input[i] = int16(i * 10)
	}

	// 设置系数（Q15格式）
	coeffs[0] = 26214  // ~0.8 in Q15
	coeffs[1] = -16384 // ~-0.5 in Q15
	coeffs[2] = 9830   // ~0.3 in Q15
	coeffs[3] = -3277  // ~-0.1 in Q15

	// 应用滤波器
	ARFilterInt16(input, output, coeffs, state, order)

	// 验证输出
	foundNonZero := false
	for i := 0; i < length; i++ {
		if output[i] != 0 {
			foundNonZero = true
			break
		}
	}

	if !foundNonZero {
		t.Error("AR filter int16 output is all zeros")
	}

	t.Logf("AR filter int16 output (first 10): %v", output[:10])
}

// TestPredictionError 测试预测误差计算
func TestPredictionError(t *testing.T) {
	length := 50

	signal := make([]int16, length)
	predicted := make([]int16, length)

	// 创建信号和预测
	for i := 0; i < length; i++ {
		signal[i] = int16(i * 10)
		predicted[i] = int16(i*10 + 5) // 预测误差为5
	}

	mse := PredictionError(signal, predicted, length)

	// MSE应该接近25（5^2）
	expectedMSE := 25.0
	if math.Abs(mse-expectedMSE) > 1.0 {
		t.Errorf("Expected MSE ~%.1f, got %.1f", expectedMSE, mse)
	}

	t.Logf("Mean squared error: %.6f", mse)
}

// TestLPCRoundTrip 测试LPC分析-合成往返
func TestLPCRoundTrip(t *testing.T) {
	// 创建原始信号
	length := 256
	original := make([]int16, length)

	for i := 0; i < length; i++ {
		original[i] = int16(1000 * math.Sin(2*math.Pi*float64(i)/32.0))
	}

	// LPC分析
	order := 12
	lpcCoeffs, gain := LPCAnalysis(original, length, order)

	// 计算残差（激励）
	excitation := make([]int16, length)
	ar := NewARFilter(order)
	ar.SetCoefficients(lpcCoeffs)
	ar.Filter(original, excitation)

	// LPC合成
	reconstructed := make([]int16, length)
	LPCSynthesis(excitation, lpcCoeffs, reconstructed)

	// 计算重建误差
	mse := PredictionError(original, reconstructed, length)

	t.Logf("LPC order: %d, Gain: %.6f", order, gain)
	t.Logf("Reconstruction MSE: %.6f", mse)
	t.Logf("Original[0]=%d, Reconstructed[0]=%d", original[0], reconstructed[0])
}

// BenchmarkLevinsonDurbin 基准测试Levinson-Durbin
func BenchmarkLevinsonDurbin(b *testing.B) {
	autoCorr := make([]float64, 17) // order=16
	for i := range autoCorr {
		autoCorr[i] = 1.0 / float64(i+1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LevinsonDurbin(autoCorr, 16)
	}
}

// BenchmarkLPCAnalysis 基准测试LPC分析
func BenchmarkLPCAnalysis(b *testing.B) {
	length := 256
	signal := make([]int16, length)

	for i := 0; i < length; i++ {
		signal[i] = int16(math.Sin(float64(i)*0.1) * 1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LPCAnalysis(signal, length, 12)
	}
}

// BenchmarkARFilterInt16 基准测试定点AR滤波器
func BenchmarkARFilterInt16(b *testing.B) {
	length := 512
	order := 10

	input := make([]int16, length)
	output := make([]int16, length)
	coeffs := make([]int16, order)
	state := make([]int16, order)

	for i := 0; i < length; i++ {
		input[i] = int16(i % 1000)
	}

	for i := 0; i < order; i++ {
		coeffs[i] = int16((i + 1) * 1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ARFilterInt16(input, output, coeffs, state, order)
	}
}

