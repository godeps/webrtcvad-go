# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **FFT功能**
  - `ComplexFFT` - 复数FFT（支持2-1024点，低/高精度模式）
  - `ComplexIFFT` - 复数逆FFT（带自动缩放）
  - `CreateRealFFT` - 实数FFT对象创建
  - `RealForwardFFT` - 实数前向FFT（CCS格式输出）
  - `RealInverseFFT` - 实数逆FFT
  - 预计算的正弦查找表（1024点）

- **互相关功能**
  - `CrossCorrelation` - 通用互相关计算（支持右移和步长）
  - `AutoCorrelation` - 自相关计算
  - `CrossCorrelationNorm` - 能量归一化互相关
  - `CrossCorrelationWithLag` - 带指定延迟的互相关
  - `FindPeakCorrelation` - 查找互相关峰值
  - `NormalizedCrossCorrelation` - Pearson相关系数计算

- **AR滤波器和LPC功能**
  - `NewARFilter` - AR滤波器创建
  - `LevinsonDurbin` - Levinson-Durbin算法（自相关→AR系数）
  - `LPCAnalysis` - 线性预测编码分析
  - `LPCSynthesis` - LPC合成滤波器
  - `ARFilterInt16` - 高性能定点AR滤波（Q15格式）
  - `ComputeParcorCoefficients` - PARCOR系数（反射系数）计算
  - `PredictionError` - 预测误差计算（MSE）

### Performance (扩展功能)
- `ComplexFFT` - ~3.4μs/op (256点)
- `RealFFT` - ~1.9μs/op (256点)
- `CrossCorrelation` - ~11.8μs/op (256×128点)
- `LevinsonDurbin` - ~476ns/op (16阶)
- `LPCAnalysis` - ~3.8μs/op (256样本，12阶)
- `ARFilterInt16` - ~5.4μs/op (512样本，10阶)

## [1.0.0] - 2025-10-29

### Added
- 纯Go实现的WebRTC VAD，无cgo依赖
- 支持8kHz, 16kHz, 32kHz, 48kHz采样率
- 支持10ms, 20ms, 30ms帧长度
- 4种激进度模式（0-3）
- 完整的单元测试套件
- 性能基准测试
- 示例程序
- 完整的中文文档

### Features
- ✅ 100%纯Go实现
- ✅ 零外部依赖
- ✅ 算法与原始WebRTC C实现完全一致
- ✅ 测试结果与Python版本(py-webrtcvad)完全匹配
- ✅ 高性能（~1-2μs/帧 @ 16kHz）
- ✅ 易于交叉编译
- ✅ 类型安全
- ✅ 完整的GoDoc注释

### Performance
- 8kHz:  ~1.0 μs/op
- 16kHz: ~1.5 μs/op
- 48kHz: ~1.8 μs/op

(测试环境: Apple M1 Pro)

### Test Results
所有测试用例通过：
- TestConstructor: PASS
- TestSetMode: PASS
- TestValidRateAndFrameLength: PASS
- TestProcessZeroes: PASS
- TestProcessFile: PASS (所有4种模式结果与Python实现完全一致)

### Documentation
- README.md - 完整的使用文档
- 内联GoDoc注释
- 示例程序
- 性能基准测试

