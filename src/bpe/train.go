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
