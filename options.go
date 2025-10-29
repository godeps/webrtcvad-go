package webrtcvad

// options.go 提供基于选项模式的VAD配置
// 使API更灵活、可扩展，同时保持向后兼容性

// Option VAD配置选项函数类型
type Option func(*VAD) error

// WithMode 设置VAD激进度模式
//
// 参数:
//   - mode: 激进度模式（0-3）
//   - 0: 质量模式（最不激进）
//   - 1: 低比特率模式
//   - 2: 激进模式
//   - 3: 非常激进模式
func WithMode(mode int) Option {
	return func(v *VAD) error {
		return v.SetMode(mode)
	}
}

// NewWithOptions 使用选项模式创建VAD实例
//
// 示例:
//
//	vad, err := webrtcvad.NewWithOptions(
//	    webrtcvad.WithMode(2),
//	)
//
// 参数:
//   - opts: 可变数量的配置选项
//
// 返回:
//   - *VAD: VAD实例
//   - error: 错误信息
func NewWithOptions(opts ...Option) (*VAD, error) {
	// 创建默认VAD实例（mode=0）
	vad, err := New(kDefaultMode)
	if err != nil {
		return nil, err
	}

	// 应用所有选项
	for _, opt := range opts {
		if err := opt(vad); err != nil {
			return nil, err
		}
	}

	return vad, nil
}

// StreamVADOption StreamVAD配置选项函数类型
type StreamVADOption func(*streamVADConfig) error

// streamVADConfig StreamVAD内部配置
type streamVADConfig struct {
	mode       int
	sampleRate int
	frameMs    int
}

// WithStreamMode 设置StreamVAD的激进度模式
func WithStreamMode(mode int) StreamVADOption {
	return func(cfg *streamVADConfig) error {
		if mode < 0 || mode > 3 {
			return ErrInvalidMode
		}
		cfg.mode = mode
		return nil
	}
}

// WithSampleRate 设置StreamVAD的采样率
func WithSampleRate(rate int) StreamVADOption {
	return func(cfg *streamVADConfig) error {
		if !isValidSampleRate(rate) {
			return ErrInvalidSampleRate
		}
		cfg.sampleRate = rate
		return nil
	}
}

// WithFrameDuration 设置StreamVAD的帧长度（毫秒）
func WithFrameDuration(ms int) StreamVADOption {
	return func(cfg *streamVADConfig) error {
		if ms != 10 && ms != 20 && ms != 30 {
			return ErrInvalidFrameLength
		}
		cfg.frameMs = ms
		return nil
	}
}

// NewStreamVADWithOptions 使用选项模式创建StreamVAD
//
// 示例:
//
//	svad, err := webrtcvad.NewStreamVADWithOptions(
//	    webrtcvad.WithStreamMode(2),
//	    webrtcvad.WithSampleRate(16000),
//	    webrtcvad.WithFrameDuration(20),
//	)
//
// 参数:
//   - opts: 可变数量的配置选项
//
// 返回:
//   - *StreamVAD: StreamVAD实例
//   - error: 错误信息
func NewStreamVADWithOptions(opts ...StreamVADOption) (*StreamVAD, error) {
	// 默认配置
	cfg := &streamVADConfig{
		mode:       1,     // 默认模式1
		sampleRate: 16000, // 默认16kHz
		frameMs:    20,    // 默认20ms
	}

	// 应用所有选项
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	// 创建StreamVAD实例
	return NewStreamVAD(cfg.mode, cfg.sampleRate, cfg.frameMs)
}

// 预定义的常用配置

// DefaultVAD 创建默认配置的VAD（mode=0，质量模式）
func DefaultVAD() (*VAD, error) {
	return New(0)
}

// AggressiveVAD 创建激进模式的VAD（mode=3）
func AggressiveVAD() (*VAD, error) {
	return New(3)
}

// DefaultStreamVAD 创建默认配置的StreamVAD
// 默认: mode=1, 16kHz, 20ms
func DefaultStreamVAD() (*StreamVAD, error) {
	return NewStreamVAD(1, 16000, 20)
}

// RealtimeStreamVAD 创建适合实时处理的StreamVAD
// 配置: mode=2, 16kHz, 10ms（低延迟）
func RealtimeStreamVAD() (*StreamVAD, error) {
	return NewStreamVAD(2, 16000, 10)
}

// HighQualityStreamVAD 创建高质量StreamVAD
// 配置: mode=0, 48kHz, 30ms（高质量，低激进度）
func HighQualityStreamVAD() (*StreamVAD, error) {
	return NewStreamVAD(0, 48000, 30)
}
