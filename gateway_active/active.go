package main

import(
	"fmt"
)

func minh123(s interface{}){
	fmt.Printf("type of s is %T", s)
//	fmt.Printf("type of s is %d \n", s(5,6))
}

func add(a int, b int) int {
	return a + b
}

var minhstr = "afdjfdlglfj"
func main() {
	fmt.Println(minhstr)
//	minh123("223345")
	minh123(add)
}