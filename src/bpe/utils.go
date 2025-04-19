package bpe

import (
	"encoding/json"
	"errors"
	"fmt"
	"normalize"
	"os"
	"runtime"
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
	totalSequences := len(dataset.aalSentences)

	// Pre-allocate the result slice to avoid resizing
	aalNewList := make([][]int64, totalSequences)

	// Use a worker pool pattern instead of spawning a goroutine per sequence
	numWorkers := runtime.NumCPU()
	jobs := make(chan int, numWorkers*4) // Buffer channel for better throughput

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Create worker pool
	for w := 0; w < numWorkers; w++ {
		go func() {
			defer wg.Done()

			for seqIdx := range jobs {
				sequence := dataset.aalSentences[seqIdx]
				sequenceLen := len(sequence)

				// Pre-allocate with estimated capacity to reduce reallocations
				// Most replacements will result in a sequence that's shorter or equal length
				estimatedNewLen := sequenceLen
				alNewSequence := make([]int64, 0, estimatedNewLen)

				// Process the sequence
				index := 0
				for index < sequenceLen {
					if index < sequenceLen-1 &&
						sequence[index] == alPair[0] &&
						sequence[index+1] == alPair[1] {
						alNewSequence = append(alNewSequence, lMintToken)
						index += 2
					} else {
						alNewSequence = append(alNewSequence, sequence[index])
						index++
					}
				}

				// No mutex needed - each worker writes to a unique index
				aalNewList[seqIdx] = alNewSequence
			}
		}()
	}

	// Distribute work
	for i := 0; i < totalSequences; i++ {
		jobs <- i
	}
	close(jobs)

	// Wait for completion
	wg.Wait()

	// Reassign
	dataset.aalSentences = aalNewList
}

// get the vocab size
func getVocabSize(dataset *dataDataset, mapTokenizer map[string]interface{}) (int, error) {
	// vocabulary size variable
	mapUniqueTokens := make(map[int64]bool)

	// get all unique tokens from all sentences
	for _, alSentence := range dataset.aalSentences {
		for _, lToken := range alSentence {
			mapUniqueTokens[lToken] = true
		}
	}

	// get all unique minted tokens
	dataMerges, tfOK := mapTokenizer["merges"]
	if !tfOK {
		return -1, errors.New("map merges does not exist")
	}
	mapMerges, tfOK := dataMerges.(map[string]interface{})
	if !tfOK {
		return -1, errors.New("Merges map type is incorrect")
	}

	return len(mapUniqueTokens) + len(mapMerges), nil
}

// loadMergesMap loads the merges map from the JSON file
func LoadMergesMap() (map[string]interface{}, error) {
	// Read merges map from JSON file
	pdFile, err := os.Open("artifacts/merges.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open merges file: %w", err)
	}
	defer pdFile.Close()

	// Create a map to store the JSON data
	var artifactsMap map[string]interface{}

	// Decode the JSON data into the map
	decoder := json.NewDecoder(pdFile)
	if err = decoder.Decode(&artifactsMap); err != nil {
		return nil, fmt.Errorf("failed to decode merges map: %w", err)
	}

	return artifactsMap, nil
}

// get sequence length
func getTotalSequenceLength(dataset *dataDataset) int64 {
	var lCount int64
	for _, alSequence := range dataset.aalSentences {
		lCount += int64(len(alSequence))
	}
	return lCount
}
