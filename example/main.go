package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/bytectlgo/webrtcvad-go"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "用法: %s <激进度> <原始音频文件路径>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "激进度: 0-3 (0=质量模式, 3=非常激进)\n")
		fmt.Fprintf(os.Stderr, "音频格式: 16位单声道PCM, 8/16/32/48 kHz采样率\n")
		os.Exit(1)
	}

	// 解析参数
	var aggressiveness int
	_, err := fmt.Sscanf(os.Args[1], "%d", &aggressiveness)
	if err != nil || aggressiveness < 0 || aggressiveness > 3 {
		log.Fatalf("无效的激进度: %s (必须是0-3)", os.Args[1])
	}

	audioFile := os.Args[2]

	// 读取音频文件
	data, err := os.ReadFile(audioFile)
	if err != nil {
		log.Fatalf("无法读取音频文件: %v", err)
	}

	// 假设音频参数（可以根据实际情况调整）
	sampleRate := 16000   // 16kHz
	frameDurationMs := 30 // 30ms帧

	// 计算帧大小
	frameSize := sampleRate * frameDurationMs / 1000 // 样本数
	frameBytes := frameSize * 2                      // 字节数（16位 = 2字节）

	fmt.Printf("音频文件: %s\n", audioFile)
	fmt.Printf("总大小: %d 字节\n", len(data))
	fmt.Printf("采样率: %d Hz\n", sampleRate)
	fmt.Printf("帧时长: %d ms\n", frameDurationMs)
	fmt.Printf("每帧字节数: %d\n", frameBytes)
	fmt.Printf("激进度模式: %d\n\n", aggressiveness)

	// 创建VAD实例
	vad, err := webrtcvad.New(aggressiveness)
	if err != nil {
		log.Fatalf("创建VAD失败: %v", err)
	}

	// 处理音频帧
	var (
		totalFrames   int
		speechFrames  int
		voiceSegments []VoiceSegment
		inSegment     bool
		segmentStart  int
	)

	fmt.Println("检测结果:")
	fmt.Println("0=静音, 1=语音")
	fmt.Println()

	for offset := 0; offset+frameBytes <= len(data); offset += frameBytes {
		chunk := data[offset : offset+frameBytes]

		isSpeech, err := vad.IsSpeech(chunk, sampleRate)
		if err != nil {
			log.Printf("处理帧失败: %v", err)
			continue
		}

		totalFrames++

		// 打印结果
		if isSpeech {
			fmt.Print("1")
			speechFrames++

			if !inSegment {
				// 开始新的语音段
				segmentStart = offset
				inSegment = true
			}
		} else {
			fmt.Print("0")

			if inSegment {
				// 结束当前语音段
				segment := VoiceSegment{
					Start: segmentStart,
					End:   offset,
				}
				voiceSegments = append(voiceSegments, segment)
				inSegment = false
			}
		}
	}

	// 处理最后一个语音段
	if inSegment {
		segment := VoiceSegment{
			Start: segmentStart,
			End:   len(data),
		}
		voiceSegments = append(voiceSegments, segment)
	}

	fmt.Println()
	fmt.Println()
	fmt.Printf("总帧数: %d\n", totalFrames)
	fmt.Printf("语音帧数: %d (%.1f%%)\n", speechFrames, float64(speechFrames)*100/float64(totalFrames))
	fmt.Printf("检测到 %d 个语音段:\n", len(voiceSegments))

	for i, seg := range voiceSegments {
		startSec := float64(seg.Start) / float64(sampleRate*2)
		endSec := float64(seg.End) / float64(sampleRate*2)
		duration := endSec - startSec
		fmt.Printf("  段 %d: %.2fs - %.2fs (时长: %.2fs)\n", i+1, startSec, endSec, duration)
	}
}

// VoiceSegment 表示一个语音段
type VoiceSegment struct {
	Start int // 字节偏移
	End   int // 字节偏移
}

// 辅助函数：将字节转换为int16样本（用于可选的音频处理）
func bytesToSamples(buf []byte) []int16 {
	samples := make([]int16, len(buf)/2)
	for i := range samples {
		samples[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
	}
	return samples
}
