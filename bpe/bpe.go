package bpe

import (
	"fmt"
	"sync"
)

// merge implements the byte pair encoding algorithm and returns an error if the merge process fails.
func merge(dataDataset *dataDataset) error {
	// Initialize max token value
	lMintToken := getMaxToken(dataDataset) + 1

	// Before vocab size
	lOldSequenceLength := getTotalSequenceLength(dataDataset)

	for {
		// initialize a map
		dataStatistics := &dataStatistics{
			mapPairFrequency: make(map[[2]int64]int),
			pdMutex:          &sync.Mutex{},
		}

		// Populate merge pairs based on current sentences
		// Get the most frequently occurring pair in the map for minting
		err := generateMergePairs(dataStatistics, dataDataset)
		if err != nil {
			return fmt.Errorf("failed to generate merge pairs: %w", err)
		}

		// replace max pair with the minted token
		replace(*dataStatistics.palMaxPair, lMintToken, dataDataset)

		// After vocab size
		newSequence := getTotalSequenceLength(dataDataset)

		// calculate compression ratio
		fCompressionRatio := float64(lOldSequenceLength) / float64(newSequence)
		fmt.Println(fCompressionRatio, string(dataStatistics.palMaxPair[0]), string(dataStatistics.palMaxPair[1]))

		// Break after a certain ratio
		if fCompressionRatio > 10 {
			break
		}
		lMintToken += 1
	}
	return nil
}

// generateMergePairs analyzes the dataset's sentences to create and track pairs of adjacent unicode points.
func generateMergePairs(dataStatistics *dataStatistics, dataDataset *dataDataset) error {
	iMaxCount := 0
	alMaxPair := [2]int64{-1, -1}
	dataStatistics.piMaxCount = &iMaxCount
	dataStatistics.palMaxPair = &alMaxPair

	for _, alUnicode := range dataDataset.aalSentences {
		for iIndex := range alUnicode {
			if iIndex+1 >= len(alUnicode) {
				continue
			}
			insertMerge(dataStatistics, [2]int64{alUnicode[iIndex], alUnicode[iIndex+1]})
		}
	}
	return nil
}
