package webrtcvad

// resample.go 实现真正的多级重采样滤波器
// 基于WebRTC的高质量重采样实现

// 全通滤波器系数（Allpass filter coefficients）
var kResampleAllpass = [2][3]int16{
	{821, 6110, 12382},
	{3050, 9368, 15063},
}

// 48kHz到32kHz插值系数 (2/3重采样)
var kCoefficients48To32 = [2][8]int16{
	{778, -2050, 1087, 23285, 12903, -3783, 441, 222},
	{222, 441, -3783, 12903, 23285, 1087, -2050, 778},
}

// downBy2ShortToInt 降采样：int16 -> int32（2倍降采样）
// 使用全通滤波器对（allpass filter pair）
//
// 参数:
//   - in: 输入样本（int16）
//   - length: 输入长度
//   - out: 输出样本（int32，移位15位+偏移16384）
//   - state: 滤波器状态（长度8）
func downBy2ShortToInt(in []int16, length int, out []int32, state []int32) {
	length >>= 1

	// 下侧全通滤波器（处理偶数索引样本）
	for i := 0; i < length; i++ {
		tmp0 := (int32(in[i<<1]) << 15) + (1 << 14)
		diff := tmp0 - state[1]
		// 缩放和舍入
		diff = (diff + (1 << 13)) >> 14
		tmp1 := state[0] + diff*int32(kResampleAllpass[1][0])
		state[0] = tmp0
		diff = tmp1 - state[2]
		// 缩放和截断
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = state[1] + diff*int32(kResampleAllpass[1][1])
		state[1] = tmp1
		diff = tmp0 - state[3]
		// 缩放和截断
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		state[3] = state[2] + diff*int32(kResampleAllpass[1][2])
		state[2] = tmp0

		// 除以2并暂存
		out[i] = state[3] >> 1
	}

	// 上侧全通滤波器（处理奇数索引样本）
	inOdd := 1
	for i := 0; i < length; i++ {
		tmp0 := (int32(in[(i<<1)+inOdd]) << 15) + (1 << 14)
		diff := tmp0 - state[5]
		// 缩放和舍入
		diff = (diff + (1 << 13)) >> 14
		tmp1 := state[4] + diff*int32(kResampleAllpass[0][0])
		state[4] = tmp0
		diff = tmp1 - state[6]
		// 缩放和舍入
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = state[5] + diff*int32(kResampleAllpass[0][1])
		state[5] = tmp1
		diff = tmp0 - state[7]
		// 缩放和截断
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		state[7] = state[6] + diff*int32(kResampleAllpass[0][2])
		state[6] = tmp0

		// 除以2并累加
		out[i] += state[7] >> 1
	}
}

// downBy2IntToShort 降采样：int32 -> int16（2倍降采样）
//
// 参数:
//   - in: 输入样本（int32，移位15位+偏移16384）【会被覆盖】
//   - length: 输入长度
//   - out: 输出样本（int16，饱和）
//   - state: 滤波器状态（长度8）
func downBy2IntToShort(in []int32, length int, out []int16, state []int32) {
	length >>= 1

	// 下侧全通滤波器
	for i := 0; i < length; i++ {
		tmp0 := in[i<<1]
		diff := tmp0 - state[1]
		// 缩放和舍入
		diff = (diff + (1 << 13)) >> 14
		tmp1 := state[0] + diff*int32(kResampleAllpass[1][0])
		state[0] = tmp0
		diff = tmp1 - state[2]
		// 缩放和截断
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = state[1] + diff*int32(kResampleAllpass[1][1])
		state[1] = tmp1
		diff = tmp0 - state[3]
		// 缩放和截断
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		state[3] = state[2] + diff*int32(kResampleAllpass[1][2])
		state[2] = tmp0

		// 除以2并暂存
		in[i<<1] = state[3] >> 1
	}

	// 上侧全通滤波器
	inOdd := 1
	for i := 0; i < length; i++ {
		tmp0 := in[(i<<1)+inOdd]
		diff := tmp0 - state[5]
		// 缩放和舍入
		diff = (diff + (1 << 13)) >> 14
		tmp1 := state[4] + diff*int32(kResampleAllpass[0][0])
		state[4] = tmp0
		diff = tmp1 - state[6]
		// 缩放和舍入
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = state[5] + diff*int32(kResampleAllpass[0][1])
		state[5] = tmp1
		diff = tmp0 - state[7]
		// 缩放和截断
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		state[7] = state[6] + diff*int32(kResampleAllpass[0][2])
		state[6] = tmp0

		// 除以2并暂存
		in[(i<<1)+inOdd] = state[7] >> 1
	}

	// 合并全通滤波器输出
	for i := 0; i < length; i += 2 {
		// 除以2，相加并舍入
		tmp0 := (in[i<<1] + in[(i<<1)+1]) >> 15
		tmp1 := (in[(i<<1)+2] + in[(i<<1)+3]) >> 15
		// 饱和到int16
		if tmp0 > 0x00007FFF {
			tmp0 = 0x00007FFF
		}
		if tmp0 < -0x00008000 {
			tmp0 = -0x00008000
		}
		out[i] = int16(tmp0)
		if tmp1 > 0x00007FFF {
			tmp1 = 0x00007FFF
		}
		if tmp1 < -0x00008000 {
			tmp1 = -0x00008000
		}
		out[i+1] = int16(tmp1)
	}
}

// lpBy2IntToInt 低通滤波（2倍降采样，int32->int32）
//
// 参数:
//   - in: 输入样本
//   - length: 输入长度
//   - out: 输出样本（长度与输入相同，但只填充length/2）
//   - state: 滤波器状态（长度16）
func lpBy2IntToInt(in []int32, length int, out []int32, state []int32) {
	halfLength := length >> 1

	// 下侧全通滤波器
	for i := 0; i < halfLength; i++ {
		tmp0 := in[i<<1]
		diff := tmp0 - state[1]
		diff = (diff + (1 << 13)) >> 14
		tmp1 := state[0] + diff*int32(kResampleAllpass[1][0])
		state[0] = tmp0
		diff = tmp1 - state[2]
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = state[1] + diff*int32(kResampleAllpass[1][1])
		state[1] = tmp1
		diff = tmp0 - state[3]
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		state[3] = state[2] + diff*int32(kResampleAllpass[1][2])
		state[2] = tmp0
		out[i] = state[3]
	}

	// 上侧全通滤波器
	for i := 0; i < halfLength; i++ {
		tmp0 := in[(i<<1)+1]
		diff := tmp0 - state[9]
		diff = (diff + (1 << 13)) >> 14
		tmp1 := state[8] + diff*int32(kResampleAllpass[0][0])
		state[8] = tmp0
		diff = tmp1 - state[10]
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		tmp0 = state[9] + diff*int32(kResampleAllpass[0][1])
		state[9] = tmp1
		diff = tmp0 - state[11]
		diff = diff >> 14
		if diff < 0 {
			diff += 1
		}
		state[11] = state[10] + diff*int32(kResampleAllpass[0][2])
		state[10] = tmp0
		out[i] = (out[i] + state[11]) >> 1
	}
}

// resample48khzTo32khz 分数重采样 2/3 (48kHz -> 32kHz)
//
// 参数:
//   - in: 输入样本（int32，长度至少 3*K+5，因为需要额外的重叠样本）
//   - out: 输出样本（int32，长度 2*K）
//   - K: 块数
func resample48khzTo32khz(in []int32, out []int32, K int) {
	inIdx := 0
	outIdx := 0

	for m := 0; m < K; m++ {
		// 第一个输出样本（需要8个输入样本：in[0]..in[7]）
		tmp := int32(1 << 14)
		for j := 0; j < 8; j++ {
			if inIdx+j < len(in) {
				tmp += int32(kCoefficients48To32[0][j]) * in[inIdx+j]
			}
		}
		out[outIdx] = tmp

		// 第二个输出样本（需要8个输入样本：in[1]..in[8]）
		tmp = int32(1 << 14)
		for j := 0; j < 8; j++ {
			if inIdx+j+1 < len(in) {
				tmp += int32(kCoefficients48To32[1][j]) * in[inIdx+j+1]
			}
		}
		out[outIdx+1] = tmp

		// 每个块消耗3个输入样本，产生2个输出样本
		inIdx += 3
		outIdx += 2
	}
}

// state48khzTo8khzFull 完整的48kHz到8kHz重采样状态
type state48khzTo8khzFull struct {
	S_48_24 [8]int32  // 48->24状态
	S_24_24 [16]int32 // 24->24(LP)状态
	S_24_16 [8]int32  // 24->16状态
	S_16_8  [8]int32  // 16->8状态
}

// resetResample48khzTo8khzFull 重置完整重采样状态
func resetResample48khzTo8khzFull(state *state48khzTo8khzFull) {
	clear(state.S_48_24[:])
	clear(state.S_24_24[:])
	clear(state.S_24_16[:])
	clear(state.S_16_8[:])
}

// resample48khzTo8khzFull 完整的48kHz到8kHz重采样
//
// 参数:
//   - in: 输入样本（480样本 @ 48kHz，10ms）
//   - out: 输出样本（80样本 @ 8kHz，10ms）
//   - state: 重采样状态
//   - tmpMem: 临时内存（至少512个int32）
func resample48khzTo8khzFull(in []int16, out []int16, state *state48khzTo8khzFull, tmpMem []int32) {
	// 阶段1: 48kHz -> 24kHz (2倍降采样)
	// 输入: 480个int16样本
	// 输出: 240个int32样本
	downBy2ShortToInt(in, 480, tmpMem[256:], state.S_48_24[:])

	// 阶段2: 24kHz -> 24kHz (低通滤波)
	// 输入: 240个int32样本
	// 输出: 240个int32样本
	lpBy2IntToInt(tmpMem[256:256+240], 240, tmpMem[16:], state.S_24_24[:])

	// 阶段3: 24kHz -> 16kHz (分数重采样 2/3)
	// 输入: 240个int32样本
	// 输出: 160个int32样本
	// 复制状态
	copy(tmpMem[8:16], state.S_24_16[:8])
	copy(state.S_24_16[:8], tmpMem[248:256])
	resample48khzTo32khz(tmpMem[8:], tmpMem[:], 80)

	// 阶段4: 16kHz -> 8kHz (2倍降采样)
	// 输入: 160个int32样本
	// 输出: 80个int16样本
	downBy2IntToShort(tmpMem[:160], 160, out, state.S_16_8[:])
}
