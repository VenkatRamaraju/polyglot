// byte pair encoding package
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
	if err := merge(dataDataset); err != nil {
		return fmt.Errorf("error running the BPE algorithm: %w", err)
	}
	return nil
}
