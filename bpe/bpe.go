package bpe

import (
	"fmt"
)

// merge implements the byte pair encoding algorithm.
// It returns an error if the merge process fails.
func merge(dataset *Dataset) error {
	// Initialize max token value
	mintToken := getMaxToken(dataset) + 1

	// Before vocab size
	oldSequence := getTotalSequenceLength(dataset)

	for {
		// initialize a map
		statistics := &Statistics{
			pairFrequency: make(map[[2]int64]int),
		}

		// Populate merge pairs based on current sentences
		// Get the most frequently occurring pair in the map for minting
		err := generateMergePairs(statistics, dataset)
		if err != nil {
			return fmt.Errorf("failed to generate merge pairs: %w", err)
		}

		// replace max pair with the minted token
		replace(*statistics.maxPair, mintToken, dataset)

		// After vocab size
		newSequence := getTotalSequenceLength(dataset)

		// calculate compression ratio
		compression := float64(oldSequence) / float64(newSequence)
		fmt.Println(compression, string(statistics.maxPair[0]), string(statistics.maxPair[1]))

		// break after a certain ratio
		if compression > 10 {
			break
		}

		// new mint token and move on
		mintToken += 1
	}
	return nil
}

// generateMergePairs analyzes the dataset's sentences to create and track pairs of adjacent unicode points.
// It updates the statistics by inserting these pairs and keeps track of the most frequently occurring pair,
// which is then returned. An error is returned if the insert operation fails.
func generateMergePairs(statistics *Statistics, dataset *Dataset) error {
	// track max pair and count
	maxCount := 0
	maxPair := [2]int64{-1, -1}

	// initialize
	statistics.maxCount = &maxCount
	statistics.maxPair = &maxPair

	for _, unicodePoints := range dataset.sentences {
		// Create merge pairs
		for index := range unicodePoints {
			// Out of range check
			if index+1 >= len(unicodePoints) {
				break
			}

			// Insert pair
			insertMerge(statistics, [2]int64{unicodePoints[index], unicodePoints[index+1]})
		}
	}
	return nil
}
