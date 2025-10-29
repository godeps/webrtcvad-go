package webrtcvad

// vad_filterbank.go 实现VAD使用的滤波器组

// 滤波器常量

// LogOfEnergy中使用的常量
const (
	kLogConst         = 24660 // 160*log10(2)，Q9定点数
	kLogEnergyIntPart = 14336 // 14，Q10定点数
)

// HighPassFilter使用的系数，Q14定点数
var (
	kHpZeroCoefs = [3]int16{6631, -13262, 6631}
	kHpPoleCoefs = [3]int16{16384, -7756, 5620}
)

// 全通滤波器系数，上部和下部，Q15定点数
// Upper: 0.64, Lower: 0.17
var kAllPassCoefsQ15 = [2]int16{20972, 5571}

// SplitFilter中除法调整的偏移向量
var kOffsetVector = [6]int16{368, 368, 272, 176, 176, 176}

// highPassFilter 高通滤波，截止频率在80 Hz（如果dataIn以500 Hz采样）
//
// 参数：
//   - dataIn：输入音频数据（500 Hz采样）
//   - dataLength：输入和输出数据的长度
//   - filterState：滤波器状态（输入/输出）
//   - dataOut：输出音频数据，频率区间 80 - 250 Hz
func highPassFilter(dataIn []int16, dataLength int, filterState []int16, dataOut []int16) {
	var tmp32 int32

	for i := 0; i < dataLength; i++ {
		// 全零点部分（滤波器系数为Q14）
		tmp32 = int32(kHpZeroCoefs[0]) * int32(dataIn[i])
		tmp32 += int32(kHpZeroCoefs[1]) * int32(filterState[0])
		tmp32 += int32(kHpZeroCoefs[2]) * int32(filterState[1])
		filterState[1] = filterState[0]
		filterState[0] = dataIn[i]

		// 全极点部分（滤波器系数为Q14）
		tmp32 -= int32(kHpPoleCoefs[1]) * int32(filterState[2])
		tmp32 -= int32(kHpPoleCoefs[2]) * int32(filterState[3])
		filterState[3] = filterState[2]
		filterState[2] = int16(tmp32 >> 14)
		dataOut[i] = filterState[2]
	}
}

// allPassFilter 对dataIn进行全通滤波，用于将信号分割为两个频带（低通 vs 高通）之前
//
// 注意：dataIn和dataOut不能对应同一地址
//
// 参数：
//   - dataIn：输入音频信号，Q0格式
//   - dataLength：输入和输出数据的长度
//   - filterCoefficient：滤波器系数，Q15格式
//   - filterState：滤波器状态，Q(-1)格式（输入/输出）
//   - dataOut：输出音频信号，Q(-1)格式
func allPassFilter(dataIn []int16, dataLength int, filterCoefficient int16,
	filterState *int16, dataOut []int16) {

	var (
		tmp16   int16
		tmp32   int32
		state32 int32 = int32(*filterState) * (1 << 16) // Q15
	)

	for i := 0; i < dataLength; i++ {
		tmp32 = state32 + int32(filterCoefficient)*int32(dataIn[i*2])
		tmp16 = int16(tmp32 >> 16) // Q(-1)
		dataOut[i] = tmp16
		state32 = (int32(dataIn[i*2]) * (1 << 14)) -
			int32(filterCoefficient)*int32(tmp16) // Q14
		state32 *= 2 // Q15
	}

	*filterState = int16(state32 >> 16) // Q(-1)
}

// splitFilter 将dataIn分割为hpDataOut和lpDataOut
// 分别对应上半部（高通）和下半部（低通）
//
// 参数：
//   - dataIn：要分割为两个频带的输入音频数据
//   - dataLength：dataIn的长度
//   - upperState：上部滤波器状态，Q(-1)格式（输入/输出）
//   - lowerState：下部滤波器状态，Q(-1)格式（输入/输出）
//   - hpDataOut：频谱上半部分的输出音频数据，长度为 dataLength / 2
//   - lpDataOut：频谱下半部分的输出音频数据，长度为 dataLength / 2
func splitFilter(dataIn []int16, dataLength int, upperState, lowerState *int16,
	hpDataOut, lpDataOut []int16) {

	halfLength := dataLength >> 1 // 降采样因子为2
	var tmpOut int16

	// 全通滤波上分支
	allPassFilter(dataIn[0:], halfLength, kAllPassCoefsQ15[0], upperState, hpDataOut)

	// 全通滤波下分支
	allPassFilter(dataIn[1:], halfLength, kAllPassCoefsQ15[1], lowerState, lpDataOut)

	// 生成LP和HP信号
	for i := 0; i < halfLength; i++ {
		tmpOut = hpDataOut[i]
		hpDataOut[i] -= lpDataOut[i]
		lpDataOut[i] += tmpOut
	}
}

// logOfEnergy 计算dataIn的能量（dB），如果必要也更新总能量totalEnergy
//
// 参数：
//   - dataIn：用于能量计算的输入音频数据
//   - dataLength：输入数据的长度
//   - offset：添加到logEnergy的偏移值
//   - totalEnergy：用dataIn的能量更新的外部能量（输入/输出）
//     注意：只有当totalEnergy <= kMinEnergy时才更新
//   - logEnergy：10 * log10("dataIn的能量")，Q4格式（输出）
func logOfEnergy(dataIn []int16, dataLength int, offset int16,
	totalEnergy *int16, logEnergy *int16) {

	// totRshifts累积在energy上执行的右移次数
	var totRshifts int = 0
	// energy将被归一化为15位。我们使用无符号整数，因为最终会屏蔽小数部分
	var energy uint32 = 0

	energy = uint32(calculateEnergy(dataIn, dataLength, &totRshifts))

	if energy != 0 {
		// 根据构造，归一化为15位等价于无符号32位值的17个前导零
		normalizingRshifts := 17 - normU32(energy)
		// 在15位表示中，前导位是2^14。Q10中的log2(2^14)是(14 << 10)
		// 这是我们初始化log2Energy的值。详细推导见下文
		var log2Energy int16 = kLogEnergyIntPart

		totRshifts += normalizingRshifts
		// 将energy归一化为15位
		// totRshifts现在是归一化后在energy上执行的右移总数
		// 这意味着energy在Q(-totRshifts)中
		if normalizingRshifts < 0 {
			energy <<= uint(-normalizingRshifts)
		} else {
			energy >>= uint(normalizingRshifts)
		}

		// 计算Q4中dataIn的能量（dB）
		//
		// 详细计算推导：
		// 10 * log10("真实能量") 的 Q4 = 2^4 * 10 * log10("真实能量") =
		// 160 * log10(energy * 2^totRshifts) =
		// 160 * log10(2) * log2(energy * 2^totRshifts) =
		// 160 * log10(2) * (log2(energy) + log2(2^totRshifts)) =
		// (160 * log10(2)) * (log2(energy) + totRshifts) =
		// kLogConst * (log2Energy + totRshifts)

		// 计算小数部分并添加到log2Energy
		log2Energy += int16((energy & 0x00003FFF) >> 4)

		// kLogConst在Q9中，log2Energy在Q10中，totRshifts在Q0中
		// 注意我们上面的推导已经考虑了Q4中的输出
		*logEnergy = int16((int32(kLogConst)*int32(log2Energy))>>19) +
			int16((int32(totRshifts)*kLogConst)>>9)

		if *logEnergy < 0 {
			*logEnergy = 0
		}
	} else {
		*logEnergy = offset
		return
	}

	*logEnergy += offset

	// 如果totalEnergy未超过kMinEnergy，用dataIn的能量更新近似totalEnergy
	// totalEnergy在vad_core.c的GmmProbability()中用作能量指示器
	if *totalEnergy <= kMinEnergy {
		if totRshifts >= 0 {
			// 根据构造，我们知道Q0中的energy > kMinEnergy
			// 所以添加一个任意值使totalEnergy超过kMinEnergy
			*totalEnergy += kMinEnergy + 1
		} else {
			// 根据构造，energy由15位表示，因此任何右移的energy都适合int16
			// 此外，只要kMinEnergy < 8192，添加值到totalEnergy是环绕安全的
			*totalEnergy += int16(energy >> uint(-totRshifts)) // Q0
		}
	}
}

// calculateFeatures 计算VAD特征向量
//
// 从8kHz采样的音频中提取6个频带的能量特征
//
// 参数：
//   - self：VAD实例
//   - dataIn：输入音频数据（8kHz采样）
//   - dataLength：数据长度（80, 160或240样本，对应10, 20或30ms）
//   - features：输出特征向量（6个频带的对数能量）
//
// 返回：总能量
func calculateFeatures(self *vadInst, dataIn []int16, dataLength int, features []int16) int16 {
	var totalEnergy int16 = 0

	// 我们期望dataLength为80、160或240样本，对应8 kHz下的10、20或30 ms
	// 因此，第一次分割后的中间降采样数据最多有120个样本
	// 第二次分割后最多有60个样本
	var (
		hp120          [120]int16
		lp120          [120]int16
		hp60           [60]int16
		lp60           [60]int16
		halfDataLength int = dataLength >> 1
		length         int = halfDataLength // dataLength / 2，对应带宽 = 2000 Hz（降采样后）
	)

	// 初始化第一个SplitFilter的变量
	frequencyBand := 0
	inPtr := dataIn      // [0 - 4000] Hz
	hpOutPtr := hp120[:] // [2000 - 4000] Hz
	lpOutPtr := lp120[:] // [0 - 2000] Hz

	// 在2000 Hz分割并降采样
	splitFilter(inPtr, dataLength, &self.upperState[frequencyBand],
		&self.lowerState[frequencyBand], hpOutPtr, lpOutPtr)

	// 对于上频带（2000 Hz - 4000 Hz），在3000 Hz分割并降采样
	frequencyBand = 1
	inPtr = hp120[:]   // [2000 - 4000] Hz
	hpOutPtr = hp60[:] // [3000 - 4000] Hz
	lpOutPtr = lp60[:] // [2000 - 3000] Hz
	splitFilter(inPtr, length, &self.upperState[frequencyBand],
		&self.lowerState[frequencyBand], hpOutPtr, lpOutPtr)

	// 3000 Hz - 4000 Hz的能量
	length >>= 1 // dataLength / 4 <=> 带宽 = 1000 Hz

	logOfEnergy(hp60[:], length, kOffsetVector[5], &totalEnergy, &features[5])

	// 2000 Hz - 3000 Hz的能量
	logOfEnergy(lp60[:], length, kOffsetVector[4], &totalEnergy, &features[4])

	// 对于下频带（0 Hz - 2000 Hz），在1000 Hz分割并降采样
	frequencyBand = 2
	inPtr = lp120[:]        // [0 - 2000] Hz
	hpOutPtr = hp60[:]      // [1000 - 2000] Hz
	lpOutPtr = lp60[:]      // [0 - 1000] Hz
	length = halfDataLength // dataLength / 2 <=> 带宽 = 2000 Hz
	splitFilter(inPtr, length, &self.upperState[frequencyBand],
		&self.lowerState[frequencyBand], hpOutPtr, lpOutPtr)

	// 1000 Hz - 2000 Hz的能量
	length >>= 1 // dataLength / 4 <=> 带宽 = 1000 Hz
	logOfEnergy(hp60[:], length, kOffsetVector[3], &totalEnergy, &features[3])

	// 对于下频带（0 Hz - 1000 Hz），在500 Hz分割并降采样
	frequencyBand = 3
	inPtr = lp60[:]     // [0 - 1000] Hz
	hpOutPtr = hp120[:] // [500 - 1000] Hz
	lpOutPtr = lp120[:] // [0 - 500] Hz
	splitFilter(inPtr, length, &self.upperState[frequencyBand],
		&self.lowerState[frequencyBand], hpOutPtr, lpOutPtr)

	// 500 Hz - 1000 Hz的能量
	length >>= 1 // dataLength / 8 <=> 带宽 = 500 Hz
	logOfEnergy(hp120[:], length, kOffsetVector[2], &totalEnergy, &features[2])

	// 对于下频带（0 Hz - 500 Hz），在250 Hz分割并降采样
	frequencyBand = 4
	inPtr = lp120[:]   // [0 - 500] Hz
	hpOutPtr = hp60[:] // [250 - 500] Hz
	lpOutPtr = lp60[:] // [0 - 250] Hz
	splitFilter(inPtr, length, &self.upperState[frequencyBand],
		&self.lowerState[frequencyBand], hpOutPtr, lpOutPtr)

	// 250 Hz - 500 Hz的能量
	length >>= 1 // dataLength / 16 <=> 带宽 = 250 Hz
	logOfEnergy(hp60[:], length, kOffsetVector[1], &totalEnergy, &features[1])

	// 通过高通滤波下频带来移除0 Hz - 80 Hz
	highPassFilter(lp60[:], length, self.hpFilterState[:], hp120[:])

	// 80 Hz - 250 Hz的能量
	logOfEnergy(hp120[:], length, kOffsetVector[0], &totalEnergy, &features[0])

	return totalEnergy
}
