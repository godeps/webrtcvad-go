package webrtcvad

// vad_sp.go 包含VAD核心使用的特定信号处理工具

// 全通滤波器系数，上部和下部，Q13定点数
// Upper: 0.64, Lower: 0.17
var kAllPassCoefsQ13 = [2]int16{5243, 1392}

// 平滑系数
const (
	kSmoothingDown = 6553  // 0.2，Q15定点数
	kSmoothingUp   = 32439 // 0.99，Q15定点数
)

// downsampling 基于分割滤波器和全通函数的降采样滤波器
//
// 通过因子2降采样信号，例如 32->16 或 16->8
//
// 输入：
//   - signalIn：输入信号
//   - inLength：输入信号的长度（样本数）
//
// 输入和输出：
//   - filterState：两个全通滤波器的当前滤波器状态
//     处理完所有样本后filterState会被更新
//
// 输出：
//   - signalOut：降采样后的信号（长度为 inLength / 2）
func downsampling(signalIn, signalOut []int16, filterState []int32, inLength int) {
	var (
		tmp16_1, tmp16_2 int16
		tmp32_1          int32 = filterState[0]
		tmp32_2          int32 = filterState[1]
		n                int
		halfLength       int = inLength >> 1 // 降采样因子为2得到一半长度
	)

	// 滤波器系数为Q13，滤波器状态为Q0
	for n = 0; n < halfLength; n++ {
		// 全通滤波上分支
		tmp16_1 = int16((tmp32_1 >> 1) +
			((int32(kAllPassCoefsQ13[0]) * int32(signalIn[n*2])) >> 14))
		signalOut[n] = tmp16_1
		tmp32_1 = int32(signalIn[n*2]) -
			((int32(kAllPassCoefsQ13[0]) * int32(tmp16_1)) >> 12)

		// 全通滤波下分支
		tmp16_2 = int16((tmp32_2 >> 1) +
			((int32(kAllPassCoefsQ13[1]) * int32(signalIn[n*2+1])) >> 14))
		signalOut[n] += tmp16_2
		tmp32_2 = int32(signalIn[n*2+1]) -
			((int32(kAllPassCoefsQ13[1]) * int32(tmp16_2)) >> 12)
	}

	// 保存滤波器状态
	filterState[0] = tmp32_1
	filterState[1] = tmp32_2
}

// findMinimum 将featureValue插入lowValueVector（如果它是最近100帧中16个最小值之一）
// 然后计算并返回五个最小值的中位数
//
// 输入：
//   - featureValue：要更新的新特征值
//   - channel：频道编号
//
// 输入和输出：
//   - self：VAD的状态信息
//
// 返回：移动窗口的平滑最小值
func findMinimum(self *vadInst, featureValue int16, channel int) int16 {
	var (
		i, j          int
		position      int   = -1
		offset        int   = channel << 4 // 偏移到内存中16个最小值的起始位置
		currentMedian int16 = 1600
		alpha         int16 = 0
		tmp32         int32 = 0
	)

	// 指向channel的16个最小值及每个值年龄的内存指针
	age := self.indexVector[offset : offset+16]
	smallestValues := self.lowValueVector[offset : offset+16]

	// smallestValues中的每个值都老了1个循环。更新age，并移除旧值
	for i = 0; i < 16; i++ {
		if age[i] != 100 {
			age[i]++
		} else {
			// 值太旧，从内存中移除并向下移动较大的值
			for j = i; j < 15; j++ {
				smallestValues[j] = smallestValues[j+1]
				age[j] = age[j+1]
			}
			age[15] = 101
			smallestValues[15] = 10000
		}
	}

	// 检查featureValue是否小于smallestValues中的任何值
	// 如果是，找到要插入新值（featureValue）的位置
	if featureValue < smallestValues[7] {
		if featureValue < smallestValues[3] {
			if featureValue < smallestValues[1] {
				if featureValue < smallestValues[0] {
					position = 0
				} else {
					position = 1
				}
			} else if featureValue < smallestValues[2] {
				position = 2
			} else {
				position = 3
			}
		} else if featureValue < smallestValues[5] {
			if featureValue < smallestValues[4] {
				position = 4
			} else {
				position = 5
			}
		} else if featureValue < smallestValues[6] {
			position = 6
		} else {
			position = 7
		}
	} else if featureValue < smallestValues[15] {
		if featureValue < smallestValues[11] {
			if featureValue < smallestValues[9] {
				if featureValue < smallestValues[8] {
					position = 8
				} else {
					position = 9
				}
			} else if featureValue < smallestValues[10] {
				position = 10
			} else {
				position = 11
			}
		} else if featureValue < smallestValues[13] {
			if featureValue < smallestValues[12] {
				position = 12
			} else {
				position = 13
			}
		} else if featureValue < smallestValues[14] {
			position = 14
		} else {
			position = 15
		}
	}

	// 如果检测到新的小值，将其插入正确位置并向上移动较大的值
	if position > -1 {
		for i = 15; i > position; i-- {
			smallestValues[i] = smallestValues[i-1]
			age[i] = age[i-1]
		}
		smallestValues[position] = featureValue
		age[position] = 1
	}

	// 获取currentMedian
	if self.frameCounter > 2 {
		currentMedian = smallestValues[2]
	} else if self.frameCounter > 0 {
		currentMedian = smallestValues[0]
	}

	// 平滑中位数值
	if self.frameCounter > 0 {
		if currentMedian < self.meanValue[channel] {
			alpha = kSmoothingDown // 0.2，Q15定点数
		} else {
			alpha = kSmoothingUp // 0.99，Q15定点数
		}
	}

	tmp32 = int32(alpha+1) * int32(self.meanValue[channel])
	tmp32 += int32(WEBRTC_SPL_WORD16_MAX-alpha) * int32(currentMedian)
	tmp32 += 16384
	self.meanValue[channel] = int16(tmp32 >> 15)

	return self.meanValue[channel]
}
