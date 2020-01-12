package main

import (
	"fmt"
	"os"
)

// check docker on host and all requirement images
func checkRequirements() error {
	return nil
}

func main() {
	err := checkRequirements()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	g := createGrader()
	fmt.Println("Grader started at :10000")
	err = g.Serve(":10000")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Grader stopped")
}
