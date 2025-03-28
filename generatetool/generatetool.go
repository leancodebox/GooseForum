package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(os.RemoveAll("app/assert/frontend/dist"))
}
