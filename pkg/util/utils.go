package util

// Max : returns the larger of x or y
func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// Min : returns the smaller of x and y
func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

// CeilDivide : returns the ceil of x / y
func CeilDivide(x, y int) int {
	if x == 0 {
		if y == 0 {
			return 1
		} else {
			return 0
		}
	}
	return int(float64(x)/float64(y) + 0.5)
}

// BoolValue : returns +1 or -1 if argument is true or false, respectively
func BoolValue(b bool) int {
	if b {
		return 1
	}
	return -1
}

// Xor : exclusive OR operator
func Xor(x bool, y bool) bool {
	return (x || y) && !(x && y)
}

// BelowMultiple : largest multiple of factor less or equal to value
func BelowMultiple(value int, factor int) (int, bool) {
	if value <= 0 || factor < 1 {
		return 0, false
	}
	return (value / factor) * factor, true
}

// AboveMultiple : smallest multiple of factor greater or equal to value
func AboveMultiple(value int, factor int) (int, bool) {
	if value <= 0 || factor < 1 {
		return 0, false
	}
	return ((value + factor - 1) / factor) * factor, true
}
