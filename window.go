package webrtcvad

import "math"

// window.go 窗函数库
// 提供常用的窗函数用于信号处理和频谱分析

// WindowFunc 窗函数类型
// 参数: n - 当前样本索引, N - 总样本数
// 返回: 窗函数值
type WindowFunc func(n, N int) float64

// HammingWindow Hamming窗
//
// 公式: w(n) = 0.54 - 0.46 * cos(2π*n/(N-1))
//
// 特点:
//   - 旁瓣抑制: -42 dB
//   - 主瓣宽度: 8π/N
//   - 适用场景: 语音信号分析、谱估计
func HammingWindow(n, N int) float64 {
	if N <= 1 {
		return 1.0
	}
	return 0.54 - 0.46*math.Cos(2*math.Pi*float64(n)/float64(N-1))
}

// HannWindow Hann窗（也称Hanning窗）
//
// 公式: w(n) = 0.5 * (1 - cos(2π*n/(N-1)))
//
// 特点:
//   - 旁瓣抑制: -31 dB
//   - 主瓣宽度: 8π/N
//   - 适用场景: FFT分析、频谱泄漏抑制
func HannWindow(n, N int) float64 {
	if N <= 1 {
		return 1.0
	}
	return 0.5 * (1 - math.Cos(2*math.Pi*float64(n)/float64(N-1)))
}

// BlackmanWindow Blackman窗
//
// 公式: w(n) = 0.42 - 0.5*cos(2π*n/(N-1)) + 0.08*cos(4π*n/(N-1))
//
// 特点:
//   - 旁瓣抑制: -58 dB（最好）
//   - 主瓣宽度: 12π/N（较宽）
//   - 适用场景: 高精度谱分析、滤波器设计
func BlackmanWindow(n, N int) float64 {
	if N <= 1 {
		return 1.0
	}
	a0, a1, a2 := 0.42, 0.5, 0.08
	x := 2 * math.Pi * float64(n) / float64(N-1)
	result := a0 - a1*math.Cos(x) + a2*math.Cos(2*x)
	// 防止浮点误差导致负值
	if result < 0 {
		return 0
	}
	return result
}

// BlackmanHarrisWindow Blackman-Harris窗
//
// 公式: w(n) = a0 - a1*cos(2π*n/(N-1)) + a2*cos(4π*n/(N-1)) - a3*cos(6π*n/(N-1))
//
// 特点:
//   - 旁瓣抑制: -92 dB（非常好）
//   - 主瓣宽度: 16π/N
//   - 适用场景: 高精度应用、动态范围要求高的场合
func BlackmanHarrisWindow(n, N int) float64 {
	if N <= 1 {
		return 1.0
	}
	a0, a1, a2, a3 := 0.35875, 0.48829, 0.14128, 0.01168
	x := 2 * math.Pi * float64(n) / float64(N-1)
	return a0 - a1*math.Cos(x) + a2*math.Cos(2*x) - a3*math.Cos(3*x)
}

// BartlettWindow Bartlett窗（三角窗）
//
// 公式: w(n) = 1 - |n - (N-1)/2| / ((N-1)/2)
//
// 特点:
//   - 旁瓣抑制: -25 dB
//   - 主瓣宽度: 8π/N
//   - 适用场景: 简单的平滑处理
func BartlettWindow(n, N int) float64 {
	if N <= 1 {
		return 1.0
	}
	return 1.0 - math.Abs(float64(n)-float64(N-1)/2.0)/(float64(N-1)/2.0)
}

// WelchWindow Welch窗（抛物窗）
//
// 公式: w(n) = 1 - ((n - (N-1)/2) / ((N-1)/2))²
//
// 特点:
//   - 类似Bartlett但更平滑
//   - 适用场景: 功率谱估计
func WelchWindow(n, N int) float64 {
	if N <= 1 {
		return 1.0
	}
	x := (float64(n) - float64(N-1)/2.0) / (float64(N-1) / 2.0)
	return 1.0 - x*x
}

// KaiserWindow Kaiser窗
//
// 参数:
//   - n: 当前样本索引
//   - N: 总样本数
//   - beta: 形状参数（通常2.0-10.0）
//
// 特点:
//   - 可调节旁瓣抑制（通过beta参数）
//   - beta=0: 等同于矩形窗
//   - beta≈5: 类似Hamming窗
//   - beta≈8.6: 类似Blackman窗
func KaiserWindow(n, N int, beta float64) float64 {
	if N <= 1 {
		return 1.0
	}
	x := 2.0*float64(n)/float64(N-1) - 1.0
	return bessel0(beta*math.Sqrt(1-x*x)) / bessel0(beta)
}

// bessel0 零阶修正贝塞尔函数（用于Kaiser窗）
func bessel0(x float64) float64 {
	sum := 1.0
	term := 1.0
	for i := 1; i < 50; i++ {
		term *= (x / 2.0) / float64(i)
		term *= (x / 2.0) / float64(i)
		sum += term
		if term < 1e-10 {
			break
		}
	}
	return sum
}

// RectangularWindow 矩形窗（无窗）
//
// 公式: w(n) = 1
//
// 特点:
//   - 最窄主瓣
//   - 最大旁瓣（-13 dB）
//   - 适用场景: 不需要窗函数的情况
func RectangularWindow(n, N int) float64 {
	return 1.0
}

// ApplyWindow 对信号应用窗函数
//
// 参数:
//   - signal: 输入信号
//   - window: 窗函数
//
// 返回:
//   - 应用窗函数后的信号
func ApplyWindow(signal []int16, window WindowFunc) []int16 {
	N := len(signal)
	result := make([]int16, N)
	for i := 0; i < N; i++ {
		result[i] = int16(float64(signal[i]) * window(i, N))
	}
	return result
}

// ApplyWindowFloat64 对float64信号应用窗函数
//
// 参数:
//   - signal: 输入信号
//   - window: 窗函数
//
// 返回:
//   - 应用窗函数后的信号
func ApplyWindowFloat64(signal []float64, window WindowFunc) []float64 {
	N := len(signal)
	result := make([]float64, N)
	for i := 0; i < N; i++ {
		result[i] = signal[i] * window(i, N)
	}
	return result
}

// ApplyWindowTo 对信号应用窗函数（零分配版本）
//
// 参数:
//   - signal: 输入信号
//   - window: 窗函数
//   - result: 预分配的结果数组
func ApplyWindowTo(signal []int16, window WindowFunc, result []int16) {
	N := len(signal)
	for i := 0; i < N; i++ {
		result[i] = int16(float64(signal[i]) * window(i, N))
	}
}

// GenerateWindow 生成窗函数数组
//
// 参数:
//   - N: 窗函数长度
//   - window: 窗函数
//
// 返回:
//   - 窗函数值数组
func GenerateWindow(N int, window WindowFunc) []float64 {
	result := make([]float64, N)
	for i := 0; i < N; i++ {
		result[i] = window(i, N)
	}
	return result
}

// WindowEnergy 计算窗函数的能量
//
// 用于归一化增益
func WindowEnergy(N int, window WindowFunc) float64 {
	var energy float64
	for i := 0; i < N; i++ {
		w := window(i, N)
		energy += w * w
	}
	return energy
}

// WindowSum 计算窗函数的和
//
// 用于幅度校正
func WindowSum(N int, window WindowFunc) float64 {
	var sum float64
	for i := 0; i < N; i++ {
		sum += window(i, N)
	}
	return sum
}

