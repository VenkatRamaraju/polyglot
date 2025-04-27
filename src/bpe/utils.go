package bpe

import (
	"encoding/json"
	"errors"
	"fmt"
	"normalize"
	"os"
	"sync"
	"time"
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
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// do it in parallel
	for _, sequence := range dataset.aalSentences {
		wg.Add(1)
		go func() {
			defer wg.Done()
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
			mutex.Lock()
			aalNewList = append(aalNewList, alNewSequence)
			mutex.Unlock()
		}()
	}

	// wait till all threads are complete
	wg.Wait()

	// reassign
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

// LoadMaps loads the merges map from the JSON file
func LoadMaps() (map[string]interface{}, map[int64]string, error) {
	// Read merges map from JSON file
	pdFile, err := os.Open("artifacts/merges.json")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open merges file: %w", err)
	}
	defer pdFile.Close()

	// Create a map to store the JSON data
	var artifactsMap map[string]interface{}

	// Decode the JSON data into the map
	decoder := json.NewDecoder(pdFile)
	if err = decoder.Decode(&artifactsMap); err != nil {
		return nil, nil, fmt.Errorf("failed to decode merges map: %w", err)
	}

	// generate Decoded map
	decodedMap, err := GenerateDecodingMap(artifactsMap)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode merges map: %w", err)
	}

	return artifactsMap, decodedMap, nil
}

// get sequence length
func getTotalSequenceLength(dataset *dataDataset) int64 {
	var lCount int64
	for _, alSequence := range dataset.aalSentences {
		lCount += int64(len(alSequence))
	}
	return lCount
}

// formatting a duration based on unit
func FormatDuration(d time.Duration) string {
	ns := d.Nanoseconds()

	switch {
	case ns < 1_000: // less than 1 microsecond
		return fmt.Sprintf("%.2fns", float64(ns))
	case ns < 1_000_000: // less than 1 millisecond
		us := float64(ns) / 1_000
		return fmt.Sprintf("%.2fµs", us)
	case ns < 1_000_000_000: // less than 1 second
		ms := float64(ns) / 1_000_000
		return fmt.Sprintf("%.2fms", ms)
	default:
		s := float64(ns) / 1_000_000_000
		return fmt.Sprintf("%.2fs", s)
	}
}
