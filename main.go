package main

import (
	"bpe"
	"fmt"
)

// main function initializes the application and starts the training process.
func main() {
	if err := bpe.Train(); err != nil {
		fmt.Println("Error during training:", err)
	}
}
