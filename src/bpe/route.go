package bpe

import (
	"fmt"
)

// Train executes the training process and returns an error if any step in the process fails.
func Train() error {
	// Get data from the source
	pdDataset, err := getData()
	if err != nil {
		return fmt.Errorf("error getting data: %w", err)
	}

	// Notify
	fmt.Println("Done getting data")

	// Perform merges on the statistics
	dataMerges, err := merge(pdDataset)
	if err != nil {
		return fmt.Errorf("error running the BPE algorithm: %w", err)
	}

	// write data to file
	err = WriteMergesMapToJSONFile(dataMerges, "artifacts/merges.json")
	if err != nil {
		return fmt.Errorf("error writing merges map to file: %w", err)
	}

	return nil
}

func GetVocabularySize() error {
	// Get data from the source
	pdDataset, err := getData()
	if err != nil {
		return fmt.Errorf("error getting data: %w", err)
	}

	// Notify
	fmt.Println("Done getting data")

	// Load the merges map
	mapMerges, err := LoadMergesMap()
	if err != nil {
		return fmt.Errorf("Failed to load merges map: %s", err)
	}

	// get the vocab size
	vocabularySize, err := getVocabSize(pdDataset, mapMerges)
	if err != nil {
		return fmt.Errorf("Error getting vocabulary size: %s", err)
	}

	fmt.Println("The vocabulary size is", vocabularySize)

	return nil
}
