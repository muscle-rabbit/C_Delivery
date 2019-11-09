package main

import (
	"fmt"
	"strings"
)

func main() {
	text := "まかない弁当(ハンバーグ)x1"
	i := strings.IndexRune(text, 'x')
	fmt.Println(string(text[i+1]))
}
