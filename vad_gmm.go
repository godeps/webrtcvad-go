package webrtcvad

// GMM相关常量
const (
	kCompVar = 22005 // 用于高斯概率计算的比较值
	kLog2Exp = 5909  // log2(exp(1))，Q12定点数
)

// gaussianProbability 计算正态分布的概率
//
// 对于正态分布，计算并返回input的概率（Q20格式）
// 正态分布概率的公式为：
//
// 1 / s * exp(-(x - m)^2 / (2 * s^2))
//
// 其中参数在以下Q域中给出：
// m = |mean|（Q7）
// s = |std|（Q7）
// x = |input|（Q4）
//
// 除了概率，我们还输出delta（Q11格式），用于更新噪声/语音模型
func gaussianProbability(input, mean, std int16, delta *int16) int32 {
	var (
		tmp16    int16
		invStd   int16
		invStd2  int16
		expValue int16 = 0
		tmp32    int32
	)

	// 计算invStd = 1 / s，Q10格式
	// 131072 = 1以Q17表示，(std >> 1)用于舍入而不是截断
	// Q域：Q17 / Q7 = Q10
	tmp32 = 131072 + int32(std>>1)
	invStd = int16(divW32W16(tmp32, std))

	// 计算invStd2 = 1 / s^2，Q14格式
	tmp16 = invStd >> 2 // Q10 -> Q8
	// Q域：(Q8 * Q8) >> 2 = Q14
	invStd2 = int16((int32(tmp16) * int32(tmp16)) >> 2)

	tmp16 = input << 3   // Q4 -> Q7
	tmp16 = tmp16 - mean // Q7 - Q7 = Q7

	// 稍后使用，用于更新噪声/语音模型
	// delta = (x - m) / s^2，Q11格式
	// Q域：(Q14 * Q7) >> 10 = Q11
	*delta = int16((int32(invStd2) * int32(tmp16)) >> 10)

	// 计算指数 tmp32 = (x - m)^2 / (2 * s^2)，Q10格式
	// 用一次移位替换除以2
	// Q域：(Q11 * Q7) >> 8 = Q10
	tmp32 = (int32(*delta) * int32(tmp16)) >> 9

	// 如果指数足够小以给出非零概率，我们计算
	// expValue ~= exp(-(x - m)^2 / (2 * s^2))
	//          ~= exp2(-log2(exp(1)) * tmp32)
	if tmp32 < kCompVar {
		// 计算 tmp16 = log2(exp(1)) * tmp32，Q10格式
		// Q域：(Q12 * Q10) >> 12 = Q10
		tmp16 = int16((kLog2Exp * tmp32) >> 12)
		tmp16 = -tmp16
		expValue = 0x0400 | (tmp16 & 0x03FF)
		tmp16 ^= int16(-1) // 使用int16(-1)替代0xFFFF
		tmp16 >>= 10
		tmp16 += 1
		// 获取 expValue = exp(-tmp32)，Q10格式
		expValue >>= uint(tmp16)
	}

	// 计算并返回 (1 / s) * exp(-(x - m)^2 / (2 * s^2))，Q20格式
	// Q域：Q10 * Q10 = Q20
	return int32(invStd) * int32(expValue)
}
