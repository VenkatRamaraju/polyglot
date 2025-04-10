package bpe

import (
	"fmt"
)

// Train executes the training process and returns an error if any step in the process fails.
func Train() error {
	// Get data from the source
	dataDataset, err := getData()
	if err != nil {
		return fmt.Errorf("error getting data: %w", err)
	}

	// Perform merges on the statistics
	mapMerges, err := merge(dataDataset)
	if err != nil {
		return fmt.Errorf("error running the BPE algorithm: %w", err)
	}

	// write data to file
	err = WriteMergesMapToJSONFile(mapMerges, "artifacts/merges.json")
	if err != nil {
		return fmt.Errorf("error writing merges map to file: %w", err)
	}

	return nil
}
