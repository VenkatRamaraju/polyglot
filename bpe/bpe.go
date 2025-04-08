package bpe

import (
	"errors"
	"fmt"
	"normalize"
	"os"
	"sync"
)

// global variables
var Merges = make(map[[2]int64]int)
var mutex sync.Mutex
var mapCount = make(map[string]int64)
var mutex2 sync.Mutex

// update max token
func updateMax(maxToken *int64, currentToken int64) {
	if currentToken > *maxToken {
		*maxToken = currentToken
	}
}

// insert into map
func insert(pair [2]int64) {
	mutex.Lock()
	defer mutex.Unlock()
	Merges[pair]++
}

// insert into map
func counts(language string, length int) {
	mutex.Lock()
	defer mutex.Unlock()
	// if _, tfOK := mapCount[language]; !tfOK {
	// 	mapCount[language] = 0
	// }
	mapCount[language] += int64(length)
}

// populate the merges map in parallel
func populateMerges() (int64, error) {
	// get files
	jsonFiles, err := listS3Keys("tknzr", "us-east-1")
	if err != nil {
		return -1, fmt.Errorf("unable to pull s3 files: %w", err)
	}

	// Initialize the map to store training merges
	var maxToken int64

	// get training dataset - we can populate map in parallel
	var wg sync.WaitGroup
	ch := make(chan error)
	for _, jsonFile := range jsonFiles {
		wg.Add(1)
		go func() {
			// decrement counter
			defer wg.Done()

			// get the file contents
			languageToSentence, err := fetchJSONFromS3("tknzr", jsonFile)
			if err != nil {
				ch <- fmt.Errorf("failed fetching JSON from S3: %w", err)
			}

			// iterate over all languages
			for language := range languageToSentence {
				// grab sentence list
				sentences, valid := languageToSentence[language].([]interface{})
				if !valid {
					ch <- errors.New("Unable to parse JSON for " + language + " into a string array")
				}

				counts(language, len(sentences))

				// // create merges
				// err = generateStatistics(sentences, &maxToken)
				// if err != nil {
				// 	ch <- fmt.Errorf("failed running the merging algorithm: %w", err)
				// }
			}
		}()
	}

	// wait for all routines to finish
	wg.Wait()

	// close the channel
	close(ch)

	// check if errors were encountered
	for err := range ch {
		return -1, fmt.Errorf("error building statistics: %w", err)
	}

	fmt.Println(mapCount)
	os.Exit(1)

	return maxToken, nil
}

// convert byte pairs to a map representation
func generateStatistics(sentences []interface{}, maxToken *int64) error {
	for _, sentence := range sentences {
		sentence, valid := sentence.(string)
		if !valid {
			return errors.New("cannnot convert to string")
		}

		// normalize sentence
		sentence = normalize.Normalize(sentence)

		// convert to unicode integers
		var unicodePoints []int64
		for _, r := range sentence {
			unicodePoints = append(unicodePoints, int64(r))
		}

		// check first pair of max
		if len(unicodePoints) > 0 {
			updateMax(maxToken, unicodePoints[0])
		}

		// create merge pairs
		for index := range unicodePoints {
			// out of range
			if index+1 >= len(unicodePoints) {
				break
			}

			// track maxes
			updateMax(maxToken, unicodePoints[index+1])

			// insert pair
			insert([2]int64{unicodePoints[index], unicodePoints[index+1]})
		}
	}
	return nil
}
