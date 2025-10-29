package webrtcvad

// generic_utils.go 使用Go泛型优化常用函数
// Go 1.18+ 支持泛型，提供更简洁、类型安全的实现

// Signed 约束所有有符号整数类型
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned 约束所有无符号整数类型
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Integer 约束所有整数类型
type Integer interface {
	Signed | Unsigned
}

// Ordered 约束所有可排序类型
type Ordered interface {
	Integer | ~float32 | ~float64 | ~string
}

// Abs 泛型绝对值函数
//
// 支持所有有符号整数类型：int, int8, int16, int32, int64
//
// 示例:
//
//	a := Abs[int16](-100)  // 返回 100
//	b := Abs[int32](-1000) // 返回 1000
func Abs[T Signed](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// Min 泛型最小值函数
//
// 支持所有可排序类型
//
// 示例:
//
//	min := Min(3, 5)           // 返回 3
//	min := Min[int16](10, 20)  // 返回 10
func Min[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max 泛型最大值函数
//
// 支持所有可排序类型
//
// 示例:
//
//	max := Max(3, 5)           // 返回 5
//	max := Max[int16](10, 20)  // 返回 20
func Max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Clamp 泛型限幅函数
//
// 将值限制在[min, max]范围内
//
// 示例:
//
//	v := Clamp(50, 0, 100)      // 返回 50
//	v := Clamp(-10, 0, 100)     // 返回 0
//	v := Clamp(150, 0, 100)     // 返回 100
func Clamp[T Ordered](value, minVal, maxVal T) T {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// MinSlice 泛型切片最小值
//
// 返回切片中的最小值，空切片返回零值
//
// 示例:
//
//	min := MinSlice([]int{3, 1, 4, 1, 5})  // 返回 1
func MinSlice[T Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}

	minVal := s[0]
	for _, v := range s[1:] {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}

// MaxSlice 泛型切片最大值
//
// 返回切片中的最大值，空切片返回零值
//
// 示例:
//
//	max := MaxSlice([]int{3, 1, 4, 1, 5})  // 返回 5
func MaxSlice[T Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}

	maxVal := s[0]
	for _, v := range s[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

// Sum 泛型切片求和
//
// 返回切片所有元素的和
//
// 示例:
//
//	sum := Sum([]int{1, 2, 3, 4, 5})  // 返回 15
func Sum[T Integer | ~float32 | ~float64](s []T) T {
	var sum T
	for _, v := range s {
		sum += v
	}
	return sum
}

// Average 泛型切片平均值（返回float64）
//
// 示例:
//
//	avg := Average([]int{1, 2, 3, 4, 5})  // 返回 3.0
func Average[T Integer | ~float32 | ~float64](s []T) float64 {
	if len(s) == 0 {
		return 0
	}
	sum := Sum(s)
	return float64(sum) / float64(len(s))
}

// 为了向后兼容，提供类型特化版本

// AbsInt16 int16绝对值（使用泛型实现）
func AbsInt16(x int16) int16 {
	return Abs(x)
}

// AbsInt32 int32绝对值（使用泛型实现）
func AbsInt32(x int32) int32 {
	return Abs(x)
}

// MinInt16 int16最小值（使用泛型实现）
func MinInt16(a, b int16) int16 {
	return Min(a, b)
}

// MaxInt16 int16最大值（使用泛型实现）
func MaxInt16(a, b int16) int16 {
	return Max(a, b)
}

// MinInt32 int32最小值（使用泛型实现）
func MinInt32(a, b int32) int32 {
	return Min(a, b)
}

// MaxInt32 int32最大值（使用泛型实现）
func MaxInt32(a, b int32) int32 {
	return Max(a, b)
}
