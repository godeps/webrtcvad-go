package webrtcvad

import (
	"testing"
)

// TestIsSpeechBatch 测试批量检测
func TestIsSpeechBatch(t *testing.T) {
	vad, err := New(1)
	if err != nil {
		t.Fatalf("创建VAD失败: %v", err)
	}

	// 创建测试帧（16kHz, 10ms = 160样本 = 320字节）
	sampleRate := 16000
	frameSize := sampleRate * 10 / 1000 * 2

	// 创建5帧
	frames := make([][]byte, 5)
	for i := range frames {
		frames[i] = make([]byte, frameSize)
		// 填充一些测试数据
		for j := range frames[i] {
			frames[i][j] = byte(j % 256)
		}
	}

	// 批量检测
	results, err := vad.IsSpeechBatch(frames, sampleRate)
	if err != nil {
		t.Fatalf("批量检测失败: %v", err)
	}

	// 验证结果数量
	if len(results) != len(frames) {
		t.Errorf("结果数量错误: 期望%d, 得到%d", len(frames), len(results))
	}

	// 每个结果都应该是bool类型
	for i, result := range results {
		// 验证类型（编译时已检查，这里只是逻辑验证）
		t.Logf("Frame %d: isSpeech=%v", i, result)
	}
}

// TestIsSpeechBatchTo 测试零分配批量检测
func TestIsSpeechBatchTo(t *testing.T) {
	vad, err := New(2)
	if err != nil {
		t.Fatalf("创建VAD失败: %v", err)
	}

	// 创建测试帧
	sampleRate := 8000
	frameSize := sampleRate * 20 / 1000 * 2
	numFrames := 10

	frames := make([][]byte, numFrames)
	for i := range frames {
		frames[i] = make([]byte, frameSize)
	}

	// 预分配结果数组
	results := make([]bool, numFrames)

	// 零分配批量检测
	err = vad.IsSpeechBatchTo(frames, sampleRate, results)
	if err != nil {
		t.Fatalf("批量检测失败: %v", err)
	}

	// 验证所有结果都被填充
	for i := range results {
		t.Logf("Frame %d: isSpeech=%v", i, results[i])
	}
}

// TestIsSpeechBatchToSmallBuffer 测试缓冲区太小的情况
func TestIsSpeechBatchToSmallBuffer(t *testing.T) {
	vad, err := New(1)
	if err != nil {
		t.Fatalf("创建VAD失败: %v", err)
	}

	sampleRate := 16000
	frameSize := sampleRate * 10 / 1000 * 2

	frames := make([][]byte, 5)
	for i := range frames {
		frames[i] = make([]byte, frameSize)
	}

	// 结果数组太小
	results := make([]bool, 3)

	err = vad.IsSpeechBatchTo(frames, sampleRate, results)
	if err == nil {
		t.Error("应该返回错误：结果数组太小")
	}
}

// TestIsSpeechBatchInvalidFrame 测试批量检测中的无效帧
func TestIsSpeechBatchInvalidFrame(t *testing.T) {
	vad, err := New(1)
	if err != nil {
		t.Fatalf("创建VAD失败: %v", err)
	}

	sampleRate := 16000
	frameSize := sampleRate * 10 / 1000 * 2

	frames := make([][]byte, 3)
	frames[0] = make([]byte, frameSize) // 有效
	frames[1] = make([]byte, 100)       // 无效长度
	frames[2] = make([]byte, frameSize) // 有效

	results, err := vad.IsSpeechBatch(frames, sampleRate)
	if err == nil {
		t.Error("应该返回错误：帧1长度无效")
	}

	// 错误应该指示是哪一帧
	t.Logf("错误信息: %v", err)
	
	// 结果应该只包含处理过的帧
	if len(results) != 3 {
		t.Logf("部分结果: %v", results)
	}
}

// BenchmarkIsSpeechBatch Benchmark批量检测
func BenchmarkIsSpeechBatch(b *testing.B) {
	vad, _ := New(1)
	sampleRate := 16000
	frameSize := sampleRate * 10 / 1000 * 2
	numFrames := 10

	frames := make([][]byte, numFrames)
	for i := range frames {
		frames[i] = make([]byte, frameSize)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vad.IsSpeechBatch(frames, sampleRate)
	}
}

// BenchmarkIsSpeechBatchTo Benchmark零分配批量检测
func BenchmarkIsSpeechBatchTo(b *testing.B) {
	vad, _ := New(1)
	sampleRate := 16000
	frameSize := sampleRate * 10 / 1000 * 2
	numFrames := 10

	frames := make([][]byte, numFrames)
	results := make([]bool, numFrames)
	for i := range frames {
		frames[i] = make([]byte, frameSize)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vad.IsSpeechBatchTo(frames, sampleRate, results)
	}
}

// BenchmarkIsSpeechSingle Benchmark单帧检测（对比）
func BenchmarkIsSpeechSingle(b *testing.B) {
	vad, _ := New(1)
	sampleRate := 16000
	frameSize := sampleRate * 10 / 1000 * 2
	frame := make([]byte, frameSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vad.IsSpeech(frame, sampleRate)
	}
}

