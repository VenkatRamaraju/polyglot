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

	// execute the instruction
	if *psFunction == "t" {
		// train mode
		if err := bpe.Train(); err != nil {
			fmt.Println("Error during training:", err)
		}
	} else {
		// api mode
		if err := bpe.EncodeDecode("artifacts/merges.json"); err != nil {
			fmt.Println("Error during training:", err)
		}
	}
}
