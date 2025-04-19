package main

import (
	"bpe"
	"flag"
	"fmt"
	"polyglot/src/server"
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
	} else if *psFunction == "v" {
		// get vocabulary size
		if err := bpe.GetVocabularySize(); err != nil {
			fmt.Println("Error while calculating vocabulary suze:", err)
		}
	} else {
		// api mode
		server.Launch()
	}
}
