package webrtcvad

import (
	"testing"
)

// TestResample48khzTo8khzFull 测试完整的重采样滤波器
func TestResample48khzTo8khzFull(t *testing.T) {
	// 创建测试输入（480样本 @ 48kHz = 10ms）
	input := make([]int16, 480)
	for i := range input {
		// 1kHz正弦波
		// f = 1000Hz, fs = 48000Hz
		// omega = 2*pi*f/fs = 2*pi*1000/48000
		input[i] = int16(10000.0 * fastSin(float64(i)*2*3.14159*1000/48000))
	}

	// 创建输出缓冲区（80样本 @ 8kHz = 10ms）
	output := make([]int16, 80)

	// 创建状态和临时内存
	var state state48khzTo8khzFull
	tmpMem := make([]int32, 512)

	// 执行重采样
	resample48khzTo8khzFull(input, output, &state, tmpMem)

	// 验证输出不全为零
	nonZeroCount := 0
	for _, v := range output {
		if v != 0 {
			nonZeroCount++
		}
	}

	if nonZeroCount == 0 {
		t.Error("输出全为零，重采样失败")
	}

	// 输出应该是80个样本
	if len(output) != 80 {
		t.Errorf("输出长度错误: 期望80, 得到%d", len(output))
	}

	t.Logf("重采样成功: %d个非零样本/%d", nonZeroCount, len(output))
}

// fastSin 快速正弦近似
func fastSin(x float64) float64 {
	// 简单的正弦近似
	for x > 3.14159 {
		x -= 2 * 3.14159
	}
	for x < -3.14159 {
		x += 2 * 3.14159
	}
	if x < 0 {
		return -x * (1.27323954 + 0.405284735*x)
	}
	return x * (1.27323954 - 0.405284735*x)
}

// TestDownBy2ShortToInt 测试2倍降采样
func TestDownBy2ShortToInt(t *testing.T) {
	// 输入: 16个样本
	input := []int16{100, 200, 300, 400, 500, 600, 700, 800,
		900, 1000, 1100, 1200, 1300, 1400, 1500, 1600}

	// 输出: 8个样本
	output := make([]int32, 8)
	state := make([]int32, 8)

	// 执行降采样
	downBy2ShortToInt(input, 16, output, state)

	// 验证输出
	if len(output) != 8 {
		t.Errorf("输出长度错误")
	}

	// 输出不应该全为零
	nonZero := false
	for _, v := range output {
		if v != 0 {
			nonZero = true
			break
		}
	}

	if !nonZero {
		t.Error("降采样输出全为零")
	}

	t.Logf("降采样成功: 输出样本 %v", output[:4])
}

// TestResample48khzTo32khz 测试分数重采样
func TestResample48khzTo32khz(t *testing.T) {
	// 输入: 240个样本 (80块 * 3)
	input := make([]int32, 240)
	for i := range input {
		input[i] = int32(i * 1000)
	}

	// 输出: 160个样本 (80块 * 2)
	output := make([]int32, 160)

	// 执行重采样
	resample48khzTo32khz(input, output, 80)

	// 验证输出长度
	if len(output) != 160 {
		t.Errorf("输出长度错误: 期望160, 得到%d", len(output))
	}

	// 输出不应该全为零
	nonZero := false
	for _, v := range output {
		if v != 0 {
			nonZero = true
			break
		}
	}

	if !nonZero {
		t.Error("重采样输出全为零")
	}

	t.Logf("分数重采样成功: 240->160样本")
}

// TestResample48khzTo8khzConsistency 测试重采样一致性
func TestResample48khzTo8khzConsistency(t *testing.T) {
	// 创建相同的输入
	input := make([]int16, 480)
	for i := range input {
		input[i] = int16((i % 100) * 100)
	}

	// 第一次重采样
	output1 := make([]int16, 80)
	var state1 state48khzTo8khzFull
	tmpMem1 := make([]int32, 512)
	resample48khzTo8khzFull(input, output1, &state1, tmpMem1)

	// 第二次重采样（新状态）
	output2 := make([]int16, 80)
	var state2 state48khzTo8khzFull
	tmpMem2 := make([]int32, 512)
	resample48khzTo8khzFull(input, output2, &state2, tmpMem2)

	// 输出应该相同
	for i := range output1 {
		if output1[i] != output2[i] {
			t.Errorf("重采样不一致: index %d, %d != %d", i, output1[i], output2[i])
		}
	}

	t.Log("重采样一致性测试通过")
}

// BenchmarkResample48khzTo8khzFull Benchmark完整重采样
func BenchmarkResample48khzTo8khzFull(b *testing.B) {
	input := make([]int16, 480)
	output := make([]int16, 80)
	var state state48khzTo8khzFull
	tmpMem := make([]int32, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resample48khzTo8khzFull(input, output, &state, tmpMem)
	}
}

// BenchmarkDownBy2ShortToInt Benchmark 2倍降采样
func BenchmarkDownBy2ShortToInt(b *testing.B) {
	input := make([]int16, 480)
	output := make([]int32, 240)
	state := make([]int32, 8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		downBy2ShortToInt(input, 480, output, state)
	}
}
