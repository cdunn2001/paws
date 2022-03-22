package web

import (
	"fmt"
)

func ExampleIsPowerOf2() {
	fmt.Printf("1 -> %t\n", IsPowerOf2(1))
	fmt.Printf("2 -> %t\n", IsPowerOf2(2))
	fmt.Printf("3 -> %t\n", IsPowerOf2(3))
	fmt.Printf("4 -> %t\n", IsPowerOf2(4))
	// Output:
	// 1 -> true
	// 2 -> true
	// 3 -> false
	// 4 -> true
}
