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
	mapMerges := &Merges{
		mapMerges: make(map[[2]int64]int64),
		alKeys:    [][2]int64{},
	}

	for {
		// initialize a map
		dataMergeStatistics := &dataStatistics{
			mapPairFrequency: make(map[[2]int64]int),
			pdMutex:          &sync.Mutex{},
		}

		// Populate merge pairs based on current sentences
		// Get the most frequently occurring pair in the map for minting
		err := generateMergePairs(dataMergeStatistics, dataDataset)
		if err != nil {
			return nil, fmt.Errorf("failed to generate merge pairs: %w", err)
		}

		// store merges
		mapMerges.insertMerge(*dataMergeStatistics.palMaxPair, lMintToken)

		// replace max pair with the minted token
		replace(*dataMergeStatistics.palMaxPair, lMintToken, dataDataset)

		// After vocab size
		newSequence := getTotalSequenceLength(dataDataset)

		// calculate compression ratio
		fCompressionRatio := float64(lOldSequenceLength) / float64(newSequence)
		fmt.Println(fCompressionRatio, string(dataMergeStatistics.palMaxPair[0]), string(dataMergeStatistics.palMaxPair[1]))

		// Break after a certain ratio
		if fCompressionRatio > 1.1 {
			break
		}
		lMintToken += 1
	}
	return mapMerges, nil
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
			insertPair(dataStatistics, [2]int64{alUnicode[iIndex], alUnicode[iIndex+1]})
		}
	}
	return nil
}
