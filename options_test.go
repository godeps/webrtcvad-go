package webrtcvad

import (
	"testing"
)

// TestNewWithOptions 测试选项模式创建VAD
func TestNewWithOptions(t *testing.T) {
	// 测试默认配置
	vad, err := NewWithOptions()
	if err != nil {
		t.Fatalf("创建默认VAD失败: %v", err)
	}
	if vad == nil {
		t.Fatal("VAD实例为nil")
	}

	// 测试带选项的配置
	vad, err = NewWithOptions(WithMode(2))
	if err != nil {
		t.Fatalf("创建VAD失败: %v", err)
	}
	if vad == nil {
		t.Fatal("VAD实例为nil")
	}

	// 测试无效模式
	_, err = NewWithOptions(WithMode(5))
	if err == nil {
		t.Error("应该拒绝无效模式")
	}
}

// TestNewStreamVADWithOptions 测试选项模式创建StreamVAD
func TestNewStreamVADWithOptions(t *testing.T) {
	// 测试默认配置
	svad, err := NewStreamVADWithOptions()
	if err != nil {
		t.Fatalf("创建默认StreamVAD失败: %v", err)
	}
	if svad == nil {
		t.Fatal("StreamVAD实例为nil")
	}

	// 测试完整配置
	svad, err = NewStreamVADWithOptions(
		WithStreamMode(2),
		WithSampleRate(16000),
		WithFrameDuration(20),
	)
	if err != nil {
		t.Fatalf("创建StreamVAD失败: %v", err)
	}
	if svad == nil {
		t.Fatal("StreamVAD实例为nil")
	}

	// 验证配置
	if svad.sampleRate != 16000 {
		t.Errorf("采样率错误: 期望16000, 得到%d", svad.sampleRate)
	}
	if svad.frameMs != 20 {
		t.Errorf("帧长度错误: 期望20, 得到%d", svad.frameMs)
	}

	// 测试无效采样率
	_, err = NewStreamVADWithOptions(WithSampleRate(11025))
	if err == nil {
		t.Error("应该拒绝无效采样率")
	}

	// 测试无效帧长度
	_, err = NewStreamVADWithOptions(WithFrameDuration(15))
	if err == nil {
		t.Error("应该拒绝无效帧长度")
	}
}

// TestPresetConfigurations 测试预定义配置
func TestPresetConfigurations(t *testing.T) {
	tests := []struct {
		name    string
		factory func() (*VAD, error)
	}{
		{"DefaultVAD", DefaultVAD},
		{"AggressiveVAD", AggressiveVAD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vad, err := tt.factory()
			if err != nil {
				t.Fatalf("创建%s失败: %v", tt.name, err)
			}
			if vad == nil {
				t.Fatalf("%s实例为nil", tt.name)
			}

			// 测试基本功能
			sampleRate := 16000
			frameSize := sampleRate * 10 / 1000 * 2
			frame := make([]byte, frameSize)
			_, err = vad.IsSpeech(frame, sampleRate)
			if err != nil {
				t.Fatalf("%s检测失败: %v", tt.name, err)
			}
		})
	}
}

// TestPresetStreamVADConfigurations 测试预定义StreamVAD配置
func TestPresetStreamVADConfigurations(t *testing.T) {
	tests := []struct {
		name    string
		factory func() (*StreamVAD, error)
	}{
		{"DefaultStreamVAD", DefaultStreamVAD},
		{"RealtimeStreamVAD", RealtimeStreamVAD},
		{"HighQualityStreamVAD", HighQualityStreamVAD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svad, err := tt.factory()
			if err != nil {
				t.Fatalf("创建%s失败: %v", tt.name, err)
			}
			if svad == nil {
				t.Fatalf("%s实例为nil", tt.name)
			}

			// 测试基本功能
			frameSize := svad.sampleRate * svad.frameMs / 1000 * 2
			frame := make([]byte, frameSize)
			_, err = svad.Write(frame)
			if err != nil {
				t.Fatalf("%s写入失败: %v", tt.name, err)
			}
		})
	}
}

// TestOptionsChaining 测试选项链式调用
func TestOptionsChaining(t *testing.T) {
	// 测试多个选项的组合
	vad, err := NewWithOptions(
		WithMode(2),
	)
	if err != nil {
		t.Fatalf("创建VAD失败: %v", err)
	}

	// 验证功能正常
	sampleRate := 16000
	frameSize := sampleRate * 10 / 1000 * 2
	frame := make([]byte, frameSize)
	_, err = vad.IsSpeech(frame, sampleRate)
	if err != nil {
		t.Fatalf("检测失败: %v", err)
	}
}

// TestStreamOptionsChaining 测试StreamVAD选项链式调用
func TestStreamOptionsChaining(t *testing.T) {
	svad, err := NewStreamVADWithOptions(
		WithStreamMode(1),
		WithSampleRate(8000),
		WithFrameDuration(10),
	)
	if err != nil {
		t.Fatalf("创建StreamVAD失败: %v", err)
	}

	// 验证配置正确
	if svad.sampleRate != 8000 {
		t.Errorf("采样率配置错误")
	}
	if svad.frameMs != 10 {
		t.Errorf("帧长度配置错误")
	}
}

// BenchmarkNewWithOptions Benchmark选项模式创建
func BenchmarkNewWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vad, _ := NewWithOptions(WithMode(2))
		_ = vad
	}
}

// BenchmarkNewDirect Benchmark直接创建（对比）
func BenchmarkNewDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vad, _ := New(2)
		_ = vad
	}
}
