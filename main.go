package main

import (
	"bpe"
	"fmt"
)

func main() {
	// kick off the training process
	err := bpe.Train()
	if err != nil {
		fmt.Println(err)
	}
}
