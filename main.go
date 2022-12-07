package main

import (
	"fmt"

	"github.com/sethvargo/go-githubactions"
)

func main() {
	val := githubactions.GetInput("val")
	fmt.Println("What is the `val`?")
	fmt.Scanln(&val)
	if val == "" {
		githubactions.Fatalf("missing 'val'")
	}
	if val != "" {
		fmt.Printf("%s", val)
	}
}
