package webrtcvad

import (
	"errors"
)

const (
	// kNumChannels 频带数量
	kNumChannels = 6
	// kNumGaussians 每个频带的高斯分布数量
	kNumGaussians = 2
	// kTableSize 查找表大小
	kTableSize = kNumChannels * kNumGaussians
	// kMinEnergy 触发音频信号的最小能量
	kMinEnergy = 10
	// kInitCheck 初始化检查标志
	kInitCheck = 42
	// kDefaultMode 默认激进度模式
	kDefaultMode = 0
	// kMaxSpeechFrames 最大连续语音帧数
	kMaxSpeechFrames = 6
	// kMinStd 最小标准差
	kMinStd = 384
)

// 频谱权重
var kSpectrumWeight = [kNumChannels]int16{6, 8, 10, 12, 14, 16}

// 噪声和语音更新常量（Q15定点数）
const (
	kNoiseUpdateConst  = 655  // Q15
	kSpeechUpdateConst = 6554 // Q15
	kBackEta           = 154  // Q8
)

// 两个模型之间的最小差异（Q5定点数）
var kMinimumDifference = [kNumChannels]int16{544, 544, 576, 576, 576, 576}

// 语音模型均值的上限（Q7定点数）
var kMaximumSpeech = [kNumChannels]int16{11392, 11392, 11520, 11520, 11520, 11520}

// 均值的最小值
var kMinimumMean = [kNumGaussians]int16{640, 768}

// 噪声模型均值的上限（Q7定点数）
var kMaximumNoise = [kNumChannels]int16{9216, 9088, 8960, 8832, 8704, 8576}

// 高斯模型的起始值（Q7定点数）

// 噪声的两个高斯权重
var kNoiseDataWeights = [kTableSize]int16{
	34, 62, 72, 66, 53, 25, 94, 66, 56, 62, 75, 103,
}

// 语音的两个高斯权重
var kSpeechDataWeights = [kTableSize]int16{
	48, 82, 45, 87, 50, 47, 80, 46, 83, 41, 78, 81,
}

// 噪声的两个高斯均值
var kNoiseDataMeans = [kTableSize]int16{
	6738, 4892, 7065, 6715, 6771, 3369, 7646, 3863, 7820, 7266, 5020, 4362,
}

// 语音的两个高斯均值
var kSpeechDataMeans = [kTableSize]int16{
	8306, 10085, 10078, 11823, 11843, 6309, 9473, 9571, 10879, 7581, 8180, 7483,
}

// 噪声的两个高斯标准差
var kNoiseDataStds = [kTableSize]int16{
	378, 1064, 493, 582, 688, 593, 474, 697, 475, 688, 421, 455,
}

// 语音的两个高斯标准差
var kSpeechDataStds = [kTableSize]int16{
	555, 505, 567, 524, 585, 1231, 509, 828, 492, 1540, 1079, 850,
}

// 不同帧长度的阈值（10ms, 20ms, 30ms）

// 模式0：质量模式
var (
	kOverHangMax1Q    = [3]int16{8, 4, 3}
	kOverHangMax2Q    = [3]int16{14, 7, 5}
	kLocalThresholdQ  = [3]int16{24, 21, 24}
	kGlobalThresholdQ = [3]int16{57, 48, 57}
)

// 模式1：低比特率模式
var (
	kOverHangMax1LBR    = [3]int16{8, 4, 3}
	kOverHangMax2LBR    = [3]int16{14, 7, 5}
	kLocalThresholdLBR  = [3]int16{37, 32, 37}
	kGlobalThresholdLBR = [3]int16{100, 80, 100}
)

// 模式2：激进模式
var (
	kOverHangMax1AGG    = [3]int16{6, 3, 2}
	kOverHangMax2AGG    = [3]int16{9, 5, 3}
	kLocalThresholdAGG  = [3]int16{82, 78, 82}
	kGlobalThresholdAGG = [3]int16{285, 260, 285}
)

// 模式3：非常激进模式
var (
	kOverHangMax1VAG    = [3]int16{6, 3, 2}
	kOverHangMax2VAG    = [3]int16{9, 5, 3}
	kLocalThresholdVAG  = [3]int16{94, 94, 94}
	kGlobalThresholdVAG = [3]int16{1100, 1050, 1100}
)

// vadInst VAD实例结构
type vadInst struct {
	vad                      int
	downsamplingFilterStates [4]int32
	state48To8               state48khzTo8khz
	noiseMeans               [kTableSize]int16
	speechMeans              [kTableSize]int16
	noiseStds                [kTableSize]int16
	speechStds               [kTableSize]int16
	frameCounter             int32
	overHang                 int16
	numOfSpeech              int16
	indexVector              [16 * kNumChannels]int16
	lowValueVector           [16 * kNumChannels]int16
	meanValue                [kNumChannels]int16
	upperState               [5]int16
	lowerState               [5]int16
	hpFilterState            [4]int16
	overHangMax1             [3]int16
	overHangMax2             [3]int16
	individual               [3]int16
	total                    [3]int16
	initFlag                 int
}

// state48khzTo8khz定义在spl.go中
// 使用完整的多级重采样滤波器实现

// createVadInst 创建VAD实例
func createVadInst() *vadInst {
	inst := &vadInst{}
	inst.initFlag = 0
	return inst
}

// initCore 初始化VAD核心组件
func initCore(self *vadInst) error {
	if self == nil {
		return errors.New("VAD instance is nil")
	}

	// 初始化通用结构变量
	self.vad = 1 // 默认语音激活
	self.frameCounter = 0
	self.overHang = 0
	self.numOfSpeech = 0

	// 初始化降采样滤波器状态
	for i := range self.downsamplingFilterStates {
		self.downsamplingFilterStates[i] = 0
	}

	// 初始化48kHz到8kHz降采样
	resetResample48khzTo8khz(&self.state48To8)

	// 读取初始PDF参数
	for i := 0; i < kTableSize; i++ {
		self.noiseMeans[i] = kNoiseDataMeans[i]
		self.speechMeans[i] = kSpeechDataMeans[i]
		self.noiseStds[i] = kNoiseDataStds[i]
		self.speechStds[i] = kSpeechDataStds[i]
	}

	// 初始化索引和最小值向量
	for i := 0; i < 16*kNumChannels; i++ {
		self.lowValueVector[i] = 10000
		self.indexVector[i] = 0
	}

	// 初始化分割滤波器状态
	for i := range self.upperState {
		self.upperState[i] = 0
	}
	for i := range self.lowerState {
		self.lowerState[i] = 0
	}

	// 初始化高通滤波器状态
	for i := range self.hpFilterState {
		self.hpFilterState[i] = 0
	}

	// 初始化均值内存（用于FindMinimum）
	for i := 0; i < kNumChannels; i++ {
		self.meanValue[i] = 1600
	}

	// 设置激进度模式为默认值
	if err := setModeCore(self, kDefaultMode); err != nil {
		return err
	}

	self.initFlag = kInitCheck

	return nil
}

// setModeCore 设置激进度模式
func setModeCore(self *vadInst, mode int) error {
	switch mode {
	case 0: // 质量模式
		copy(self.overHangMax1[:], kOverHangMax1Q[:])
		copy(self.overHangMax2[:], kOverHangMax2Q[:])
		copy(self.individual[:], kLocalThresholdQ[:])
		copy(self.total[:], kGlobalThresholdQ[:])
	case 1: // 低比特率模式
		copy(self.overHangMax1[:], kOverHangMax1LBR[:])
		copy(self.overHangMax2[:], kOverHangMax2LBR[:])
		copy(self.individual[:], kLocalThresholdLBR[:])
		copy(self.total[:], kGlobalThresholdLBR[:])
	case 2: // 激进模式
		copy(self.overHangMax1[:], kOverHangMax1AGG[:])
		copy(self.overHangMax2[:], kOverHangMax2AGG[:])
		copy(self.individual[:], kLocalThresholdAGG[:])
		copy(self.total[:], kGlobalThresholdAGG[:])
	case 3: // 非常激进模式
		copy(self.overHangMax1[:], kOverHangMax1VAG[:])
		copy(self.overHangMax2[:], kOverHangMax2VAG[:])
		copy(self.individual[:], kLocalThresholdVAG[:])
		copy(self.total[:], kGlobalThresholdVAG[:])
	default:
		return errors.New("invalid mode")
	}

	return nil
}

// process 处理音频帧并返回VAD决策
func process(inst *vadInst, fs int, audioFrame []int16) (int, error) {
	if inst == nil {
		return -1, errors.New("VAD instance is nil")
	}

	if inst.initFlag != kInitCheck {
		return -1, errors.New("VAD not initialized")
	}

	if len(audioFrame) == 0 {
		return -1, errors.New("audio frame is nil or empty")
	}

	frameLength := len(audioFrame)
	if !ValidRateAndFrameLength(fs, frameLength) {
		return -1, errors.New("invalid rate and frame length")
	}

	var vad int
	var err error

	switch fs {
	case 48000:
		vad, err = calcVad48khz(inst, audioFrame, frameLength)
	case 32000:
		vad, err = calcVad32khz(inst, audioFrame, frameLength)
	case 16000:
		vad, err = calcVad16khz(inst, audioFrame, frameLength)
	case 8000:
		vad, err = calcVad8khz(inst, audioFrame, frameLength)
	default:
		return -1, errors.New("unsupported sample rate")
	}

	if err != nil {
		return -1, err
	}

	// 将VAD值归一化为0或1
	if vad > 0 {
		vad = 1
	}

	return vad, nil
}
