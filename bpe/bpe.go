package bpe

import (
	"fmt"
)

// merge implements the byte pair encoding algorithm.
// It returns an error if the merge process fails.
func merge() error {
	// Initialize max token value
	mintToken := getMaxToken() + 1

	// Before vocab size
	oldSequence := getTotalSequenceLength()

	for {
		// Clear the map of previous statistics
		clearMerges()

		// Populate merge pairs based on current sentences
		// Get the most frequently occurring pair in the map for minting
		maxPair, err := generateMergePairs()
		if err != nil {
			return fmt.Errorf("failed to generate merge pairs: %w", err)
		}

		// replace max pair with the minted token
		replace(maxPair, mintToken)

		// After vocab size
		newSequence := getTotalSequenceLength()

		// calculate compression ratio
		compression := float64(oldSequence) / float64(newSequence)
		fmt.Println(compression, string(maxPair[0]), string(maxPair[1]))

		// break after a certain ratio
		if compression > 10 {
			break
		}

		// new mint token
		mintToken += 1
	}
	return nil
}

// generateMergePairs converts byte pairs from SentenceList into a map representation.
// It returns an error if the insert operation fails.
func generateMergePairs() ([2]int64, error) {
	// track max pair and count
	maxCount := 0
	maxPair := [2]int64{-1, -1}

	for _, unicodePoints := range SentenceList {
		// Create merge pairs
		for index := range unicodePoints {
			// Out of range check
			if index+1 >= len(unicodePoints) {
				break
			}

			// Insert pair
			insertMerge([2]int64{unicodePoints[index], unicodePoints[index+1]}, &maxPair, &maxCount)
		}
	}
	return maxPair, nil
}
