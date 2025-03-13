package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(os.RemoveAll("app/assert/frontend/dist"))
	fmt.Println(os.RemoveAll("app/assert/frontend/dist2"))
}
