package web

func IsPowerOf2(n int32) bool {
	return 0 == (n & (n - 1))
}
