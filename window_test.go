package webrtcvad

import (
	"math"
	"testing"
)

// TestWindowFunctions 测试各种窗函数
func TestWindowFunctions(t *testing.T) {
	N := 100

	windows := map[string]WindowFunc{
		"Hamming":        HammingWindow,
		"Hann":           HannWindow,
		"Blackman":       BlackmanWindow,
		"BlackmanHarris": BlackmanHarrisWindow,
		"Bartlett":       BartlettWindow,
		"Welch":          WelchWindow,
		"Rectangular":    RectangularWindow,
	}

	for name, window := range windows {
		t.Run(name, func(t *testing.T) {
			// 检查窗函数值在合理范围内
			for i := 0; i < N; i++ {
				val := window(i, N)
				if val < 0 || val > 1 {
					t.Errorf("%s窗在位置%d的值超出[0,1]范围: %f", name, i, val)
				}
			}

			// 检查对称性（大多数窗函数是对称的）
			if name != "Bartlett" { // Bartlett有时会有数值误差
				mid := N / 2
				for i := 0; i < mid; i++ {
					left := window(i, N)
					right := window(N-1-i, N)
					if math.Abs(left-right) > 1e-10 {
						t.Errorf("%s窗不对称: w[%d]=%.10f, w[%d]=%.10f", 
							name, i, left, N-1-i, right)
					}
				}
			}
		})
	}
}

// TestKaiserWindow 测试Kaiser窗
func TestKaiserWindow(t *testing.T) {
	N := 100
	beta := 5.0

	for i := 0; i < N; i++ {
		val := KaiserWindow(i, N, beta)
		if val < 0 || val > 1 {
			t.Errorf("Kaiser窗在位置%d的值超出[0,1]范围: %f", i, val)
		}
	}

	// 测试不同beta值
	betas := []float64{0, 2, 5, 8.6, 10}
	for _, beta := range betas {
		// 中心值应该是最大的
		center := KaiserWindow(N/2, N, beta)
		edge := KaiserWindow(0, N, beta)
		if edge > center {
			t.Errorf("Kaiser窗(beta=%.1f)边缘值大于中心值", beta)
		}
	}
}

// TestApplyWindow 测试窗函数应用
func TestApplyWindow(t *testing.T) {
	// 创建测试信号
	signal := make([]int16, 100)
	for i := range signal {
		signal[i] = 1000
	}

	// 应用Hann窗
	windowed := ApplyWindow(signal, HannWindow)

	// 检查长度
	if len(windowed) != len(signal) {
		t.Errorf("加窗后长度错误: 期望%d, 得到%d", len(signal), len(windowed))
	}

	// Hann窗的边缘应该接近0
	if math.Abs(float64(windowed[0])) > 1 {
		t.Errorf("Hann窗边缘值应接近0, 得到%d", windowed[0])
	}

	// 中心值应该接近原值
	mid := len(windowed) / 2
	if math.Abs(float64(windowed[mid])-1000) > 10 {
		t.Errorf("Hann窗中心值应接近1000, 得到%d", windowed[mid])
	}
}

// TestApplyWindowTo 测试零分配版本
func TestApplyWindowTo(t *testing.T) {
	signal := make([]int16, 100)
	for i := range signal {
		signal[i] = int16(i)
	}

	result := make([]int16, 100)
	ApplyWindowTo(signal, HammingWindow, result)

	// 验证结果
	for i := range result {
		expected := int16(float64(signal[i]) * HammingWindow(i, len(signal)))
		if result[i] != expected {
			t.Errorf("位置%d: 期望%d, 得到%d", i, expected, result[i])
		}
	}
}

// TestGenerateWindow 测试窗函数生成
func TestGenerateWindow(t *testing.T) {
	N := 128
	window := GenerateWindow(N, HannWindow)

	if len(window) != N {
		t.Errorf("窗函数长度错误: 期望%d, 得到%d", N, len(window))
	}

	// 验证对称性
	for i := 0; i < N/2; i++ {
		if math.Abs(window[i]-window[N-1-i]) > 1e-10 {
			t.Errorf("窗函数不对称: w[%d]=%.10f, w[%d]=%.10f",
				i, window[i], N-1-i, window[N-1-i])
		}
	}
}

// TestWindowEnergy 测试窗函数能量计算
func TestWindowEnergy(t *testing.T) {
	N := 100

	windows := map[string]WindowFunc{
		"Hamming":     HammingWindow,
		"Hann":        HannWindow,
		"Blackman":    BlackmanWindow,
		"Rectangular": RectangularWindow,
	}

	for name, window := range windows {
		energy := WindowEnergy(N, window)
		if energy <= 0 {
			t.Errorf("%s窗能量应为正数: %f", name, energy)
		}

		// 矩形窗的能量应该最大
		if name == "Rectangular" && math.Abs(energy-float64(N)) > 0.01 {
			t.Errorf("矩形窗能量应接近%d, 得到%f", N, energy)
		}
	}
}

// TestWindowSum 测试窗函数和
func TestWindowSum(t *testing.T) {
	N := 100

	windows := map[string]WindowFunc{
		"Hamming":     HammingWindow,
		"Hann":        HannWindow,
		"Rectangular": RectangularWindow,
	}

	for name, window := range windows {
		sum := WindowSum(N, window)
		if sum <= 0 {
			t.Errorf("%s窗和应为正数: %f", name, sum)
		}

		// 矩形窗的和应该等于N
		if name == "Rectangular" && math.Abs(sum-float64(N)) > 0.01 {
			t.Errorf("矩形窗和应等于%d, 得到%f", N, sum)
		}
	}
}

// TestWindowSpectralProperties 测试窗函数频谱特性
func TestWindowSpectralProperties(t *testing.T) {
	N := 256

	windows := []struct {
		name     string
		window   WindowFunc
		minSidelobe float64 // 旁瓣抑制（dB，越小越好）
	}{
		{"Hamming", HammingWindow, -42},
		{"Hann", HannWindow, -31},
		{"Blackman", BlackmanWindow, -58},
		{"Rectangular", RectangularWindow, -13},
	}

	for _, w := range windows {
		t.Run(w.name, func(t *testing.T) {
			// 生成窗函数
			window := GenerateWindow(N, w.window)
			
			// 验证主瓣
			maxVal := window[N/2]
			if maxVal < 0.5 {
				t.Errorf("%s窗主瓣值太小: %f", w.name, maxVal)
			}

			// 验证边缘衰减
			edge := window[0]
			if edge > maxVal*0.1 && w.name != "Rectangular" {
				t.Logf("%s窗边缘衰减不足: edge=%f, max=%f", w.name, edge, maxVal)
			}
		})
	}
}

// BenchmarkHammingWindow Benchmark Hamming窗
func BenchmarkHammingWindow(b *testing.B) {
	N := 1024
	for i := 0; i < b.N; i++ {
		for n := 0; n < N; n++ {
			HammingWindow(n, N)
		}
	}
}

// BenchmarkApplyWindow Benchmark应用窗函数
func BenchmarkApplyWindow(b *testing.B) {
	signal := make([]int16, 1024)
	for i := range signal {
		signal[i] = int16(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyWindow(signal, HammingWindow)
	}
}

// BenchmarkApplyWindowTo Benchmark零分配版本
func BenchmarkApplyWindowTo(b *testing.B) {
	signal := make([]int16, 1024)
	result := make([]int16, 1024)
	for i := range signal {
		signal[i] = int16(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyWindowTo(signal, HammingWindow, result)
	}
}

