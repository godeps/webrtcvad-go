package webrtcvad

import "errors"

// errors.go 定义常用错误类型

var (
	// ErrInvalidMode 无效的VAD模式
	ErrInvalidMode = errors.New("mode must be 0-3")

	// ErrInvalidSampleRate 无效的采样率
	ErrInvalidSampleRate = errors.New("sample rate must be 8000, 16000, 32000, or 48000 Hz")

	// ErrInvalidFrameLength 无效的帧长度
	ErrInvalidFrameLength = errors.New("frame length must correspond to 10, 20, or 30 ms")

	// ErrNotInitialized VAD未初始化
	ErrNotInitialized = errors.New("VAD not initialized")

	// ErrBufferTooSmall 缓冲区太小
	ErrBufferTooSmall = errors.New("buffer too small")
)
