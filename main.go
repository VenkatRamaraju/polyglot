package main

import (
	"bpe"
	"flag"
	"fmt"
)

// main function initializes the application and starts the training process.
func main() {
	// get a function
	psFunction := flag.String("func", "", "Configuration File")
	flag.Parse()

	if *psFunction == "t" {
		// train
		if err := bpe.Train(); err != nil {
			fmt.Println("Error during training:", err)
		}
	} else {
		// Encode-decode demonstration
		if err := bpe.EncodeDecode("artifacts/merges.json"); err != nil {
			fmt.Println("Error during training:", err)
		}
	}
}
