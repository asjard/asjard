package main

import "fmt"

func a(b string) {
	b = "xx"
}

func main() {
	data := "yy"
	a(data)
	fmt.Println(data)
}
