package webrtcvad

import "math"

// correlation.go 实现信号的互相关和自相关功能
//
// 提供两种API版本：
// 1. 常规版本：自动分配返回数组（便捷）
// 2. To版本：使用预分配数组（零分配，高性能）

// CrossCorrelationTo 计算互相关（零分配版本）
//
// 参数:
//   - seq1: 第一个输入序列
//   - seq2: 第二个输入序列
//   - dimSeq: 序列长度
//   - dimCrossCorrelation: 互相关输出长度
//   - rightShifts: 右移位数（用于防止溢出）
//   - stepSeq2: seq2的步长（通常为1）
//   - result: 预分配的结果数组（长度应 >= dimCrossCorrelation）
//
// 注意：此函数不分配内存，适合高频调用场景
func CrossCorrelationTo(seq1, seq2 []int16, dimSeq int,
	dimCrossCorrelation int, rightShifts int, stepSeq2 int, result []int32) {

	seq2Ptr := 0
	for i := 0; i < dimCrossCorrelation; i++ {
		var corr int32 = 0
		
		// 4路循环展开优化
		j := 0
		for ; j+3 < dimSeq; j += 4 {
			if seq2Ptr+j+3 < len(seq2) && j+3 < len(seq1) {
				corr += (int32(seq1[j]) * int32(seq2[seq2Ptr+j])) >> uint(rightShifts)
				corr += (int32(seq1[j+1]) * int32(seq2[seq2Ptr+j+1])) >> uint(rightShifts)
				corr += (int32(seq1[j+2]) * int32(seq2[seq2Ptr+j+2])) >> uint(rightShifts)
				corr += (int32(seq1[j+3]) * int32(seq2[seq2Ptr+j+3])) >> uint(rightShifts)
			} else {
				break
			}
		}
		
		// 处理剩余
		for ; j < dimSeq; j++ {
			if seq2Ptr+j < len(seq2) && j < len(seq1) {
				corr += (int32(seq1[j]) * int32(seq2[seq2Ptr+j])) >> uint(rightShifts)
			}
		}
		
		seq2Ptr += stepSeq2
		result[i] = corr
	}
}

// CrossCorrelation 计算两个序列的互相关
//
// 参数:
//   - seq1: 第一个输入序列
//   - seq2: 第二个输入序列
//   - dimSeq: 序列长度
//   - dimCrossCorrelation: 互相关输出长度
//   - rightShifts: 右移位数（用于防止溢出）
//   - stepSeq2: seq2的步长（通常为1）
//
// 返回:
//   - crossCorrelation: 互相关结果数组
//
// 公式:
//
//	crossCorrelation[i] = Σ(seq1[j] * seq2[i*stepSeq2 + j]) >> rightShifts
func CrossCorrelation(seq1, seq2 []int16, dimSeq int,
	dimCrossCorrelation int, rightShifts int, stepSeq2 int) []int32 {

	crossCorrelation := make([]int32, dimCrossCorrelation)
	CrossCorrelationTo(seq1, seq2, dimSeq, dimCrossCorrelation, rightShifts, stepSeq2, crossCorrelation)
	return crossCorrelation
}

// AutoCorrelationTo 计算自相关（零分配版本）
//
// 参数:
//   - seq: 输入序列
//   - dimSeq: 序列长度
//   - dimAutoCorrelation: 自相关输出长度（通常 <= dimSeq）
//   - rightShifts: 右移位数
//   - result: 预分配的结果数组
func AutoCorrelationTo(seq []int16, dimSeq int,
	dimAutoCorrelation int, rightShifts int, result []int32) {
	CrossCorrelationTo(seq, seq, dimSeq, dimAutoCorrelation, rightShifts, 1, result)
}

// AutoCorrelation 计算信号的自相关
//
// 这是CrossCorrelation的特殊情况，其中seq1 == seq2
//
// 参数:
//   - seq: 输入序列
//   - dimSeq: 序列长度
//   - dimAutoCorrelation: 自相关输出长度（通常 <= dimSeq）
//   - rightShifts: 右移位数
//
// 返回:
//   - autoCorrelation: 自相关结果数组
func AutoCorrelation(seq []int16, dimSeq int,
	dimAutoCorrelation int, rightShifts int) []int32 {

	return CrossCorrelation(seq, seq, dimSeq, dimAutoCorrelation, rightShifts, 1)
}

// CrossCorrelationNorm 归一化互相关
//
// 计算互相关并进行能量归一化
//
// 参数:
//   - seq1: 第一个序列（长度 dimSeq）
//   - seq2: 第二个序列（长度 >= dimSeq + dimCrossCorrelation - 1）
//   - dimSeq: 序列长度
//   - dimCrossCorrelation: 互相关输出长度
//   - scale: 缩放因子
//
// 返回:
//   - 归一化的互相关数组
func CrossCorrelationNorm(seq1, seq2 []int16, dimSeq int,
	dimCrossCorrelation int, scale int) []float64 {

	result := make([]float64, dimCrossCorrelation)

	// 计算seq1的能量
	var energy1 float64 = 0
	for i := 0; i < dimSeq; i++ {
		energy1 += float64(seq1[i]) * float64(seq1[i])
	}

	if energy1 == 0 {
		return result
	}

	// 计算每个延迟的互相关和能量
	for lag := 0; lag < dimCrossCorrelation; lag++ {
		var corr float64 = 0
		var energy2 float64 = 0

		for i := 0; i < dimSeq; i++ {
			val1 := float64(seq1[i])
			val2 := float64(seq2[lag+i])
			corr += val1 * val2
			energy2 += val2 * val2
		}

		// 归一化
		if energy2 > 0 {
			norm := corr / (energy1 * energy2)
			result[lag] = norm * float64(scale)
		}
	}

	return result
}

// CrossCorrelationWithLag 计算带特定延迟的互相关
//
// 参数:
//   - seq1: 第一个序列
//   - seq2: 第二个序列
//   - dimSeq: 序列长度
//   - lag: 延迟（样本数）
//   - rightShifts: 右移位数
//
// 返回:
//   - 互相关值
func CrossCorrelationWithLag(seq1, seq2 []int16, dimSeq int, lag int, rightShifts int) int32 {
	var corr int32 = 0

	if lag >= 0 {
		// 正延迟：seq2相对seq1向右移动
		maxLen := dimSeq - lag
		if maxLen > 0 {
			for i := 0; i < maxLen; i++ {
				corr += (int32(seq1[i]) * int32(seq2[i+lag])) >> uint(rightShifts)
			}
		}
	} else {
		// 负延迟：seq2相对seq1向左移动
		lag = -lag
		maxLen := dimSeq - lag
		if maxLen > 0 {
			for i := 0; i < maxLen; i++ {
				corr += (int32(seq1[i+lag]) * int32(seq2[i])) >> uint(rightShifts)
			}
		}
	}

	return corr
}

// FindPeakCorrelation 在互相关结果中查找峰值
//
// 参数:
//   - correlation: 互相关数组
//
// 返回:
//   - peakIndex: 峰值索引
//   - peakValue: 峰值
func FindPeakCorrelation(correlation []int32) (int, int32) {
	if len(correlation) == 0 {
		return -1, 0
	}

	peakIndex := 0
	peakValue := correlation[0]

	for i := 1; i < len(correlation); i++ {
		if correlation[i] > peakValue {
			peakValue = correlation[i]
			peakIndex = i
		}
	}

	return peakIndex, peakValue
}

// NormalizedCrossCorrelation 归一化互相关（返回-1到1之间的值）
//
// 计算归一化相关系数，类似于Pearson相关系数
//
// 参数:
//   - seq1, seq2: 输入序列
//   - length: 序列长度
//
// 返回:
//   - 归一化相关系数（-1.0 到 1.0）
func NormalizedCrossCorrelation(seq1, seq2 []int16, length int) float64 {
	if length == 0 {
		return 0
	}

	// 计算均值
	var sum1, sum2 int32 = 0, 0
	for i := 0; i < length; i++ {
		sum1 += int32(seq1[i])
		sum2 += int32(seq2[i])
	}
	mean1 := float64(sum1) / float64(length)
	mean2 := float64(sum2) / float64(length)

	// 计算协方差和方差
	var covariance, var1, var2 float64 = 0, 0, 0
	for i := 0; i < length; i++ {
		diff1 := float64(seq1[i]) - mean1
		diff2 := float64(seq2[i]) - mean2
		covariance += diff1 * diff2
		var1 += diff1 * diff1
		var2 += diff2 * diff2
	}

	// 归一化：使用标准差的乘积
	if var1 > 0 && var2 > 0 {
		return covariance / math.Sqrt(var1*var2)
	}

	return 0
}

