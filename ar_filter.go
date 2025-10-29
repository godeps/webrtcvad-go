package webrtcvad

import "math"

// ar_filter.go 实现自回归(AR)滤波器和相关功能
// AR滤波器常用于语音编码和预测

// ARFilter AR滤波器结构
type ARFilter struct {
	order       int       // 滤波器阶数
	coeffs      []float64 // AR系数
	state       []float64 // 滤波器状态
}

// NewARFilter 创建新的AR滤波器
//
// order: 滤波器阶数
func NewARFilter(order int) *ARFilter {
	return &ARFilter{
		order:  order,
		coeffs: make([]float64, order+1),
		state:  make([]float64, order),
	}
}

// SetCoefficients 设置AR滤波器系数
//
// coeffs: AR系数数组（长度应为order+1）
func (ar *ARFilter) SetCoefficients(coeffs []float64) {
	copy(ar.coeffs, coeffs)
}

// Filter 应用AR滤波器
//
// input: 输入信号
// output: 输出信号
//
// AR滤波器方程：
//
//	y[n] = x[n] - Σ(a[k] * y[n-k])  k=1..order
func (ar *ARFilter) Filter(input []int16, output []int16) {
	length := len(input)

	for n := 0; n < length; n++ {
		// 当前输入
		var y float64 = float64(input[n])

		// 减去AR预测部分
		for k := 0; k < ar.order; k++ {
			y -= ar.coeffs[k+1] * ar.state[k]
		}

		// 更新状态（移位）
		for k := ar.order - 1; k > 0; k-- {
			ar.state[k] = ar.state[k-1]
		}
		ar.state[0] = y

		// 输出
		output[n] = int16(y)
	}
}

// LevinsonDurbin Levinson-Durbin算法
//
// 从自相关系数计算AR系数（用于线性预测）
//
// 参数:
//   - autoCorr: 自相关系数 [R(0), R(1), ..., R(order)]
//   - order: AR模型阶数
//
// 返回:
//   - arCoeffs: AR系数 [1, -a1, -a2, ..., -a_order]
//   - predictionError: 预测误差功率
func LevinsonDurbin(autoCorr []float64, order int) ([]float64, float64) {
	if len(autoCorr) < order+1 {
		return nil, 0
	}

	arCoeffs := make([]float64, order+1)
	arCoeffs[0] = 1.0

	if autoCorr[0] == 0 {
		return arCoeffs, 0
	}

	// 初始化
	predictionError := autoCorr[0]
	reflectionCoeffs := make([]float64, order)

	// Levinson-Durbin递归
	for m := 0; m < order; m++ {
		// 计算反射系数
		var sum float64 = autoCorr[m+1]
		for k := 0; k < m; k++ {
			sum += arCoeffs[k+1] * autoCorr[m-k]
		}
		reflectionCoeffs[m] = -sum / predictionError

		// 更新AR系数
		arCoeffs[m+1] = reflectionCoeffs[m]
		for k := 0; k < m; k++ {
			tmp := arCoeffs[k+1]
			arCoeffs[k+1] = tmp + reflectionCoeffs[m]*arCoeffs[m-k]
		}

		// 更新预测误差
		predictionError *= (1.0 - reflectionCoeffs[m]*reflectionCoeffs[m])

		// 检查稳定性
		if predictionError <= 0 {
			break
		}
	}

	return arCoeffs, predictionError
}

// LPCAnalysis 线性预测编码(LPC)分析
//
// 计算LPC系数，这是AR模型的特殊应用
//
// 参数:
//   - signal: 输入信号
//   - length: 信号长度
//   - order: LPC阶数
//
// 返回:
//   - lpcCoeffs: LPC系数
//   - gain: 增益
func LPCAnalysis(signal []int16, length int, order int) ([]float64, float64) {
	// 计算自相关
	autoCorr := make([]float64, order+1)
	for lag := 0; lag <= order; lag++ {
		var sum float64 = 0
		for n := 0; n < length-lag; n++ {
			sum += float64(signal[n]) * float64(signal[n+lag])
		}
		autoCorr[lag] = sum
	}

	// 使用Levinson-Durbin算法计算LPC系数
	lpcCoeffs, predError := LevinsonDurbin(autoCorr, order)

	// 计算增益
	var gain float64 = 0
	if predError > 0 && autoCorr[0] > 0 {
		gain = math.Sqrt(predError / autoCorr[0])
	}

	return lpcCoeffs, gain
}

// LPCSynthesis LPC合成滤波器
//
// 使用LPC系数合成语音信号
//
// 参数:
//   - excitation: 激励信号（残差）
//   - lpcCoeffs: LPC系数
//   - output: 输出合成信号
func LPCSynthesis(excitation []int16, lpcCoeffs []float64, output []int16) {
	order := len(lpcCoeffs) - 1
	length := len(excitation)
	state := make([]float64, order)

	for n := 0; n < length; n++ {
		// 计算输出
		var y float64 = float64(excitation[n])

		// 添加AR预测部分
		for k := 0; k < order && k < n; k++ {
			y -= lpcCoeffs[k+1] * state[k]
		}

		// 更新状态
		for k := order - 1; k > 0; k-- {
			state[k] = state[k-1]
		}
		state[0] = y

		output[n] = int16(y)
	}
}

// ComputeParcorCoefficients 计算偏自相关系数(PARCOR)
//
// 偏自相关系数（也称为反射系数）在语音编码中很有用
//
// 参数:
//   - autoCorr: 自相关系数
//   - order: 模型阶数
//
// 返回:
//   - parcor: 偏自相关系数
func ComputeParcorCoefficients(autoCorr []float64, order int) []float64 {
	if len(autoCorr) < order+1 {
		return nil
	}

	parcor := make([]float64, order)
	arCoeffs := make([]float64, order+1)
	arCoeffs[0] = 1.0

	if autoCorr[0] == 0 {
		return parcor
	}

	predictionError := autoCorr[0]

	for m := 0; m < order; m++ {
		// 计算反射系数（偏自相关系数）
		var sum float64 = autoCorr[m+1]
		for k := 0; k < m; k++ {
			sum += arCoeffs[k+1] * autoCorr[m-k]
		}
		parcor[m] = -sum / predictionError

		// 更新AR系数
		arCoeffs[m+1] = parcor[m]
		for k := 0; k < m; k++ {
			tmp := arCoeffs[k+1]
			arCoeffs[k+1] = tmp + parcor[m]*arCoeffs[m-k]
		}

		// 更新预测误差
		predictionError *= (1.0 - parcor[m]*parcor[m])

		if predictionError <= 0 {
			break
		}
	}

	return parcor
}

// ARFilterInt16 定点AR滤波器（使用int16）
//
// 高效的定点实现，用于实时处理
//
// 参数:
//   - input: 输入信号
//   - output: 输出信号
//   - coeffs: AR系数（Q15格式）
//   - state: 滤波器状态
//   - order: 滤波器阶数
func ARFilterInt16(input []int16, output []int16, coeffs []int16, state []int16, order int) {
	length := len(input)

	for n := 0; n < length; n++ {
		// 当前输入
		var y int32 = int32(input[n]) << 15

		// 减去AR预测部分（定点运算）
		for k := 0; k < order; k++ {
			y -= (int32(coeffs[k]) * int32(state[k]))
		}

		// 归一化
		y = (y + (1 << 14)) >> 15

		// 更新状态（移位）
		for k := order - 1; k > 0; k-- {
			state[k] = state[k-1]
		}
		state[0] = int16(y)

		// 输出
		output[n] = int16(y)
	}
}

// PredictionError 计算预测误差
//
// 参数:
//   - signal: 原始信号
//   - predicted: 预测信号
//   - length: 信号长度
//
// 返回:
//   - MSE: 均方误差
func PredictionError(signal []int16, predicted []int16, length int) float64 {
	var sumSquaredError float64 = 0

	for i := 0; i < length; i++ {
		err := float64(signal[i]) - float64(predicted[i])
		sumSquaredError += err * err
	}

	return sumSquaredError / float64(length)
}

