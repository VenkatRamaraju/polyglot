package bpe

import (
	"normalize"
	"sync"
)

// dataDataset holds the sentences and a mutex for concurrent access.
type dataDataset struct {
	aalSentences [][]int64
	pdMutex      *sync.Mutex
}

// Merges tracks the order of insertions into a map
type Merges struct {
	mapMerges map[[2]int64]int64
	alKeys    [][2]int64
}

// dataStatistics holds the frequency of pairs and a mutex for concurrent access.
type dataStatistics struct {
	mapPairFrequency map[[2]int64]int
	pdMutex          *sync.Mutex
	palMaxPair       *[2]int64
	piMaxCount       *int
}

func (m *Merges) insertMerge(alPair [2]int64, lMintedToken int64) {
	// append to array
	m.alKeys = append(m.alKeys, alPair)

	// add to map
	m.mapMerges[alPair] = lMintedToken
}

// insertPair increments the count for a given pair in StatisticsMap
// also tracks the most frequently occurring pair
func insertPair(dataStatistics *dataStatistics, alPair [2]int64) {
	dataStatistics.pdMutex.Lock()
	defer dataStatistics.pdMutex.Unlock()

	// Increment pair
	dataStatistics.mapPairFrequency[alPair]++

	// update max pair
	if dataStatistics.mapPairFrequency[alPair] > *dataStatistics.piMaxCount {
		*dataStatistics.palMaxPair = alPair
		*dataStatistics.piMaxCount = dataStatistics.mapPairFrequency[alPair]
	}
}

// add a single sentence to a list
func (d *dataDataset) add(alSentence []int64) {
	d.pdMutex.Lock()
	defer d.pdMutex.Unlock()
	d.aalSentences = append(d.aalSentences, alSentence)
}

// AddList add set of sentences to a list
func (d *dataDataset) AddList(adataSentences []interface{}) {
	for index := range adataSentences {
		// Normalize the sentence
		sentence := normalize.Normalize(adataSentences[index].(string))

		// Convert to unicode integers
		var unicodePoints []int64
		for _, r := range sentence {
			unicodePoints = append(unicodePoints, int64(r))
		}

		// Add to list
		d.add(unicodePoints)
	}
}

// getMaxToken scans a list of unicode point sequences and returns the highest token value.
func getMaxToken(dataset *dataDataset) int64 {
	var lMaxToken int64 = -1
	for _, alSentence := range dataset.aalSentences {
		for _, lToken := range alSentence {
			if lToken > lMaxToken {
				lMaxToken = lToken
			}
		}
	}
	return lMaxToken
}

// replaces one token with another
func replace(alPair [2]int64, lMintToken int64, dataset *dataDataset) {
	// new list
	var aalNewList [][]int64

	for _, sequence := range dataset.aalSentences {
		index := 0
		var alNewSequence []int64
		for index < len(sequence) {
			if index < len(sequence)-1 && sequence[index] == alPair[0] && sequence[index+1] == alPair[1] {
				alNewSequence = append(alNewSequence, lMintToken)
				index += 2
			} else {
				alNewSequence = append(alNewSequence, sequence[index])
				index += 1
			}
		}
		aalNewList = append(aalNewList, alNewSequence)
	}

	// reassign
	dataset.aalSentences = aalNewList
}

// get the vocab size
func getVocabSize(dataset *dataDataset) int {
	mapUnique := make(map[int64]bool)
	for _, alSequence := range dataset.aalSentences {
		for index := range alSequence {
			mapUnique[alSequence[index]] = true
		}
	}
	return len(mapUnique)
}

// get sequence length
func getTotalSequenceLength(dataset *dataDataset) int64 {
	var lCount int64
	for _, alSequence := range dataset.aalSentences {
		lCount += int64(len(alSequence))
	}
	return lCount
}
