package webrtcvad

// calcVad8khz 计算8kHz音频的VAD
func calcVad8khz(inst *vadInst, speechFrame []int16, frameLength int) (int, error) {
	featureVector := make([]int16, kNumChannels)

	// 获取频带能量
	totalPower := calculateFeatures(inst, speechFrame, frameLength, featureVector)

	// 执行VAD判决
	inst.vad = int(gmmProbability(inst, featureVector, totalPower, frameLength))

	return inst.vad, nil
}

// calcVad16khz 计算16kHz音频的VAD
func calcVad16khz(inst *vadInst, speechFrame []int16, frameLength int) (int, error) {
	speechNB := make([]int16, 240) // 降采样后的语音帧：480样本（30ms宽带）

	// 宽带：在执行VAD前降采样
	downsampling(speechFrame, speechNB, inst.downsamplingFilterStates[:], frameLength)

	length := frameLength / 2
	vad, err := calcVad8khz(inst, speechNB, length)

	return vad, err
}

// calcVad32khz 计算32kHz音频的VAD
func calcVad32khz(inst *vadInst, speechFrame []int16, frameLength int) (int, error) {
	speechWB := make([]int16, 480) // 降采样后的语音帧：960样本（30ms超宽带）
	speechNB := make([]int16, 240) // 降采样后的语音帧：480样本（30ms宽带）

	// 降采样信号 32->16->8 然后执行VAD
	downsampling(speechFrame, speechWB, inst.downsamplingFilterStates[2:], frameLength)
	length := frameLength / 2

	downsampling(speechWB, speechNB, inst.downsamplingFilterStates[:], length)
	length /= 2

	// 在8kHz信号上执行VAD
	vad, err := calcVad8khz(inst, speechNB, length)

	return vad, err
}

// calcVad48khz 计算48kHz音频的VAD
func calcVad48khz(inst *vadInst, speechFrame []int16, frameLength int) (int, error) {
	const (
		kFrameLen10ms48khz = 480
		kFrameLen10ms8khz  = 80
	)

	speechNB := make([]int16, 240) // 30ms的8kHz数据
	tmpMem := make([]int32, 480+256)

	num10msFrames := frameLength / kFrameLen10ms48khz

	for i := 0; i < num10msFrames; i++ {
		startIdx := i * kFrameLen10ms48khz
		endIdx := startIdx + kFrameLen10ms48khz
		outStartIdx := i * kFrameLen10ms8khz

		resample48khzTo8khz(
			speechFrame[startIdx:endIdx],
			speechNB[outStartIdx:outStartIdx+kFrameLen10ms8khz],
			&inst.state48To8,
			tmpMem,
		)
	}

	// 在8kHz信号上执行VAD
	vad, err := calcVad8khz(inst, speechNB, frameLength/6)

	return vad, err
}

// weightedAverage 计算加权平均值
//
// data被加上offset后再进行平均
func weightedAverage(data []int16, offset int16, weights []int16) int32 {
	var weightedAverage int32 = 0

	for k := 0; k < kNumGaussians; k++ {
		idx := k * kNumChannels
		data[idx] += offset
		weightedAverage += int32(data[idx]) * int32(weights[idx])
	}

	return weightedAverage
}

// overflowingMulS16ByS32ToS32 允许溢出的乘法（保持与C代码行为一致）
func overflowingMulS16ByS32ToS32(a int16, b int32) int32 {
	return int32(a) * b
}

// gmmProbability 使用高斯混合模型计算语音和背景噪声的概率
//
// 执行假设检验来决定哪种类型的信号更可能
//
// 返回VAD决策（0 - 噪声，1 - 语音）
func gmmProbability(self *vadInst, features []int16, totalPower int16, frameLength int) int16 {
	var (
		channel, k            int
		featureMinimum        int16
		h0, h1                int16
		logLikelihoodRatio    int16
		vadflag               int16 = 0
		shiftsH0, shiftsH1    int16
		tmpS16, tmp1S16       int16
		tmp2S16               int16
		diff                  int16
		gaussian              int
		nmk, nmk2, nmk3       int16
		smk, smk2             int16
		nsk, ssk              int16
		delt, ndelt           int16
		maxspe, maxmu         int16
		deltaN                [kTableSize]int16
		deltaS                [kTableSize]int16
		ngprvec               [kTableSize]int16 // 条件概率 = 0
		sgprvec               [kTableSize]int16 // 条件概率 = 0
		h0Test, h1Test        int32
		tmp1S32, tmp2S32      int32
		sumLogLikelihoodRatio int32 = 0
		noiseGlobalMean       int32
		speechGlobalMean      int32
		noiseProbability      [kNumGaussians]int32
		speechProbability     [kNumGaussians]int32
		overhead1, overhead2  int16
		individualTest        int16
		totalTest             int16
	)

	// 根据帧长度设置各种阈值（80, 160或240样本）
	if frameLength == 80 {
		overhead1 = self.overHangMax1[0]
		overhead2 = self.overHangMax2[0]
		individualTest = self.individual[0]
		totalTest = self.total[0]
	} else if frameLength == 160 {
		overhead1 = self.overHangMax1[1]
		overhead2 = self.overHangMax2[1]
		individualTest = self.individual[1]
		totalTest = self.total[1]
	} else {
		overhead1 = self.overHangMax1[2]
		overhead2 = self.overHangMax2[2]
		individualTest = self.individual[2]
		totalTest = self.total[2]
	}

	if totalPower > kMinEnergy {
		// 当前帧的信号功率足够大，可以处理
		// 处理包含两部分：
		// 1) 计算语音的似然度，从而做出VAD决策
		// 2) 根据决策更新底层模型

		// 检测方案是一个LRT（似然比检验），假设为：
		// H0: 噪声
		// H1: 语音

		// 我们将全局LRT与局部测试结合，对每个频率子带（称为channel）

		for channel = 0; channel < kNumChannels; channel++ {
			// 对每个频道，我们用包含kNumGaussians的GMM建模概率
			// 根据H0或H1有不同的均值和标准差
			h0Test = 0
			h1Test = 0

			for k = 0; k < kNumGaussians; k++ {
				gaussian = channel + k*kNumChannels

				// H0下的概率，即帧为噪声的概率
				// 值以Q27给出 = Q7 * Q20
				tmp1S32 = gaussianProbability(
					features[channel],
					self.noiseMeans[gaussian],
					self.noiseStds[gaussian],
					&deltaN[gaussian],
				)
				noiseProbability[k] = int32(kNoiseDataWeights[gaussian]) * tmp1S32
				h0Test += noiseProbability[k] // Q27

				// H1下的概率，即帧为语音的概率
				// 值以Q27给出 = Q7 * Q20
				tmp1S32 = gaussianProbability(
					features[channel],
					self.speechMeans[gaussian],
					self.speechStds[gaussian],
					&deltaS[gaussian],
				)
				speechProbability[k] = int32(kSpeechDataWeights[gaussian]) * tmp1S32
				h1Test += speechProbability[k] // Q27
			}

			// 计算对数似然比：log2(Pr{X|H1} / Pr{X|H0})
			// 近似：
			// log2(Pr{X|H1} / Pr{X|H0}) ≈ shifts_h0 - shifts_h1
			shiftsH0 = normW32(h0Test)
			shiftsH1 = normW32(h1Test)
			if h0Test == 0 {
				shiftsH0 = 31
			}
			if h1Test == 0 {
				shiftsH1 = 31
			}
			logLikelihoodRatio = shiftsH0 - shiftsH1

			// 用频谱权重更新sum_log_likelihood_ratios
			// 这用于全局VAD决策
			sumLogLikelihoodRatio += int32(logLikelihoodRatio) * int32(kSpectrumWeight[channel])

			// 局部VAD决策
			if (logLikelihoodRatio * 4) > individualTest {
				vadflag = 1
			}

			// 计算局部噪声概率（稍后更新GMM时使用）
			h0 = int16(h0Test >> 12) // Q15
			if h0 > 0 {
				// 高噪声概率。为GMM中的每个高斯分配条件概率
				tmp1S32 = int32(uint32(noiseProbability[0])&0xFFFFF000) << 2 // Q29
				ngprvec[channel] = int16(divW32W16(tmp1S32, h0))             // Q14
				ngprvec[channel+kNumChannels] = 16384 - ngprvec[channel]
			} else {
				// 低噪声概率。为第一个高斯分配条件概率1，其余为0
				ngprvec[channel] = 16384
			}

			// 计算局部语音概率（稍后更新GMM时使用）
			h1 = int16(h1Test >> 12) // Q15
			if h1 > 0 {
				// 高语音概率。为GMM中的每个高斯分配条件概率
				tmp1S32 = int32(uint32(speechProbability[0])&0xFFFFF000) << 2 // Q29
				sgprvec[channel] = int16(divW32W16(tmp1S32, h1))              // Q14
				sgprvec[channel+kNumChannels] = 16384 - sgprvec[channel]
			}
		}

		// 做出全局VAD决策
		if sumLogLikelihoodRatio >= int32(totalTest) {
			vadflag = 1
		}

		// 更新模型参数
		maxspe = 12800
		for channel = 0; channel < kNumChannels; channel++ {
			// 获取过去的最小值，用于长期修正，Q4格式
			featureMinimum = findMinimum(self, features[channel], channel)

			// 计算"全局"均值，即两个均值的加权和
			noiseGlobalMean = weightedAverage(
				self.noiseMeans[channel:],
				0,
				kNoiseDataWeights[channel:],
			)
			tmp1S16 = int16(noiseGlobalMean >> 6) // Q8

			for k = 0; k < kNumGaussians; k++ {
				gaussian = channel + k*kNumChannels

				nmk = self.noiseMeans[gaussian]
				smk = self.speechMeans[gaussian]
				nsk = self.noiseStds[gaussian]
				ssk = self.speechStds[gaussian]

				// 如果帧只包含噪声，更新噪声均值向量
				nmk2 = nmk
				if vadflag == 0 {
					// deltaN = (x-mu)/sigma^2
					// ngprvec[k] = |noise_probability[k]| /
					//   (|noise_probability[0]| + |noise_probability[1]|)

					// (Q14 * Q11 >> 11) = Q14
					delt = int16((int32(ngprvec[gaussian]) * int32(deltaN[gaussian])) >> 11)
					// Q7 + (Q14 * Q15 >> 22) = Q7
					nmk2 = nmk + int16((int32(delt)*kNoiseUpdateConst)>>22)
				}

				// 噪声均值的长期修正
				// Q8 - Q8 = Q8
				ndelt = (featureMinimum << 4) - tmp1S16
				// Q7 + (Q8 * Q8) >> 9 = Q7
				nmk3 = nmk2 + int16((int32(ndelt)*kBackEta)>>9)

				// 控制噪声均值不要漂移太多
				tmpS16 = int16((k + 5) << 7)
				if nmk3 < tmpS16 {
					nmk3 = tmpS16
				}
				tmpS16 = int16((72 + k - channel) << 7)
				if nmk3 > tmpS16 {
					nmk3 = tmpS16
				}
				self.noiseMeans[gaussian] = nmk3

				if vadflag != 0 {
					// 更新语音均值向量：
					// |deltaS| = (x-mu)/sigma^2
					// sgprvec[k] = |speech_probability[k]| /
					//   (|speech_probability[0]| + |speech_probability[1]|)

					// (Q14 * Q11) >> 11 = Q14
					delt = int16((int32(sgprvec[gaussian]) * int32(deltaS[gaussian])) >> 11)
					// Q14 * Q15 >> 21 = Q8
					tmpS16 = int16((int32(delt) * kSpeechUpdateConst) >> 21)
					// Q7 + (Q8 >> 1) = Q7。带舍入
					smk2 = smk + ((tmpS16 + 1) >> 1)

					// 控制语音均值不要漂移太多
					maxmu = maxspe + 640
					if smk2 < kMinimumMean[k] {
						smk2 = kMinimumMean[k]
					}
					if smk2 > maxmu {
						smk2 = maxmu
					}
					self.speechMeans[gaussian] = smk2 // Q7

					// (Q7 >> 3) = Q4。带舍入
					tmpS16 = (smk + 4) >> 3
					tmpS16 = features[channel] - tmpS16 // Q4
					// (Q11 * Q4 >> 3) = Q12
					tmp1S32 = (int32(deltaS[gaussian]) * int32(tmpS16)) >> 3
					tmp2S32 = tmp1S32 - 4096
					tmpS16 = sgprvec[gaussian] >> 2
					// (Q14 >> 2) * Q12 = Q24
					tmp1S32 = int32(tmpS16) * tmp2S32

					tmp2S32 = tmp1S32 >> 4 // Q20

					// 0.1 * Q20 / Q7 = Q13
					if tmp2S32 > 0 {
						tmpS16 = int16(divW32W16(tmp2S32, ssk*10))
					} else {
						tmpS16 = int16(divW32W16(-tmp2S32, ssk*10))
						tmpS16 = -tmpS16
					}
					// 除以4，更新因子为0.025 (= 0.1 / 4)
					// 除以4等于右移2位，因此
					// (Q13 >> 8) = (Q13 >> 6) / 4 = Q7
					tmpS16 += 128 // 舍入
					ssk += tmpS16 >> 8
					if ssk < kMinStd {
						ssk = kMinStd
					}
					self.speechStds[gaussian] = ssk
				} else {
					// 更新GMM方差向量
					// deltaN * (features[channel] - nmk) - 1
					// Q4 - (Q7 >> 3) = Q4
					tmpS16 = features[channel] - (nmk >> 3)
					// (Q11 * Q4 >> 3) = Q12
					tmp1S32 = (int32(deltaN[gaussian]) * int32(tmpS16)) >> 3
					tmp1S32 -= 4096

					// (Q14 >> 2) * Q12 = Q24
					tmpS16 = (ngprvec[gaussian] + 2) >> 2
					tmp2S32 = overflowingMulS16ByS32ToS32(tmpS16, tmp1S32)
					// Q20 * 约0.001 (2^-10=0.0009766)，因此
					// (Q24 >> 14) = (Q24 >> 4) / 2^10 = Q20
					tmp1S32 = tmp2S32 >> 14

					// Q20 / Q7 = Q13
					if tmp1S32 > 0 {
						tmpS16 = int16(divW32W16(tmp1S32, nsk))
					} else {
						tmpS16 = int16(divW32W16(-tmp1S32, nsk))
						tmpS16 = -tmpS16
					}
					tmpS16 += 32       // 舍入
					nsk += tmpS16 >> 6 // Q13 >> 6 = Q7
					if nsk < kMinStd {
						nsk = kMinStd
					}
					self.noiseStds[gaussian] = nsk
				}
			}

			// 如果模型太接近，分离它们
			// noiseGlobalMean以Q14表示 (= Q7 * Q7)
			noiseGlobalMean = weightedAverage(
				self.noiseMeans[channel:],
				0,
				kNoiseDataWeights[channel:],
			)

			// speechGlobalMean以Q14表示 (= Q7 * Q7)
			speechGlobalMean = weightedAverage(
				self.speechMeans[channel:],
				0,
				kSpeechDataWeights[channel:],
			)

			// diff = "全局"语音均值 - "全局"噪声均值
			// (Q14 >> 9) - (Q14 >> 9) = Q5
			diff = int16(speechGlobalMean>>9) - int16(noiseGlobalMean>>9)

			if diff < kMinimumDifference[channel] {
				tmpS16 = kMinimumDifference[channel] - diff

				// tmp1S16 = ~0.8 * (kMinimumDifference - diff)，Q7
				// tmp2S16 = ~0.2 * (kMinimumDifference - diff)，Q7
				tmp1S16 = int16((13 * int32(tmpS16)) >> 2)
				tmp2S16 = int16((3 * int32(tmpS16)) >> 2)

				// 为语音模型移动高斯均值tmp1S16，并更新speechGlobalMean
				speechGlobalMean = weightedAverage(
					self.speechMeans[channel:],
					tmp1S16,
					kSpeechDataWeights[channel:],
				)

				// 为噪声模型移动高斯均值-tmp2S16，并更新noiseGlobalMean
				noiseGlobalMean = weightedAverage(
					self.noiseMeans[channel:],
					-tmp2S16,
					kNoiseDataWeights[channel:],
				)
			}

			// 控制语音和噪声均值不要漂移太多
			maxspe = kMaximumSpeech[channel]
			tmp2S16 = int16(speechGlobalMean >> 7)
			if tmp2S16 > maxspe {
				// 语音模型的上限
				tmp2S16 -= maxspe
				for k = 0; k < kNumGaussians; k++ {
					self.speechMeans[channel+k*kNumChannels] -= tmp2S16
				}
			}

			tmp2S16 = int16(noiseGlobalMean >> 7)
			if tmp2S16 > kMaximumNoise[channel] {
				tmp2S16 -= kMaximumNoise[channel]
				for k = 0; k < kNumGaussians; k++ {
					self.noiseMeans[channel+k*kNumChannels] -= tmp2S16
				}
			}
		}
		self.frameCounter++
	}

	// 关于转换迟滞的平滑
	if vadflag == 0 {
		if self.overHang > 0 {
			vadflag = 2 + self.overHang
			self.overHang--
		}
		self.numOfSpeech = 0
	} else {
		self.numOfSpeech++
		if self.numOfSpeech > kMaxSpeechFrames {
			self.numOfSpeech = kMaxSpeechFrames
			self.overHang = overhead2
		} else {
			self.overHang = overhead1
		}
	}

	return vadflag
}
