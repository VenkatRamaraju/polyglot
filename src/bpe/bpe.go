package bpe

import (
	"fmt"
	"sync"
)

// merge implements the byte pair encoding algorithm and returns an error if the merge process fails.
func merge(dataDataset *dataDataset) (*Merges, error) {
	// Initialize max token value
	lMintToken := getMaxToken(dataDataset) + 1

	// Before vocab size
	lOldSequenceLength := getTotalSequenceLength(dataDataset)

	// store merges
	dataMerges := &Merges{
		mapMerges: make(map[[2]int64]int64),
		alKeys:    [][2]int64{},
	}

	for {
		// initialize a map
		pdMergeStatistics := &dataStatistics{
			mapPairFrequency: make(map[[2]int64]int),
			pdMutex:          &sync.Mutex{},
		}

		// Populate merge pairs based on current sentences
		// Store the most frequently occurring pair
		err := countStatistics(pdMergeStatistics, dataDataset)
		if err != nil {
			return nil, fmt.Errorf("failed to generate merge pairs: %w", err)
		}

		// store merges
		dataMerges.insertMerge(*pdMergeStatistics.palMaxPair, lMintToken)

		// replace max pair with the minted token
		replace(*pdMergeStatistics.palMaxPair, lMintToken, dataDataset)

		// After vocab size
		newSequence := getTotalSequenceLength(dataDataset)

		// calculate compression ratio
		fCompressionRatio := float64(lOldSequenceLength) / float64(newSequence)
		fmt.Println(fCompressionRatio, string(pdMergeStatistics.palMaxPair[0]), string(pdMergeStatistics.palMaxPair[1]))

		// Break after a certain ratio
		if fCompressionRatio > 5 {
			break
		}

		// next minted token
		lMintToken += 1
	}
	return dataMerges, nil
}

// countStatistics analyzes the dataset's sentences to create and track pairs of adjacent unicode points.
func countStatistics(dataStatistics *dataStatistics, dataDataset *dataDataset) error {
	// variables to track
	iMaxCount := 0
	alMaxPair := [2]int64{-1, -1}
	dataStatistics.piMaxCount = &iMaxCount
	dataStatistics.palMaxPair = &alMaxPair

	// Count each occurence
	for _, alUnicode := range dataDataset.aalSentences {
		for iIndex := range alUnicode {
			if iIndex+1 >= len(alUnicode) {
				continue
			}
			insertPair(dataStatistics, [2]int64{alUnicode[iIndex], alUnicode[iIndex+1]})
		}
	}
	return nil
}
