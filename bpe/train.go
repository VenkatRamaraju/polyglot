// byte pair encoding package
package bpe

import (
	"fmt"
)

// Train orchestrates the training process by populating data and performing merges.
// It returns an error if any step in the process fails.
func Train() error {
	// Get data from the source
	err := populateData()
	if err != nil {
		return fmt.Errorf("error populating the merges: %w", err)
	}

	// Perform merges on the statistics
	err = merge()
	if err != nil {
		return fmt.Errorf("error running the BPE algorithm: %w", err)
	}

	return nil
}
