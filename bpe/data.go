package bpe

import (
	"errors"
	"fmt"
	"sync"
)

// getData retrieves all sentences from S3
func getData() (*Dataset, error) {
	dataset := &Dataset{}

	// Get files
	jsonFiles, err := listS3Keys("tknzr", "us-east-1")
	if err != nil {
		return nil, fmt.Errorf("unable to pull s3 files: %w", err)
	}

	// Get training dataset - we can populate map in parallel
	var wg sync.WaitGroup
	ch := make(chan error)
	for _, jsonFile := range jsonFiles {
		wg.Add(1)
		go func(fileName string) {
			// Defer completion of routine
			defer wg.Done()

			// Get the file contents
			languageToSentence, err := fetchJSONFromS3("tknzr", fileName)
			if err != nil {
				ch <- err
			}

			// Iterate over all languages
			for language := range languageToSentence {
				// Grab sentence list
				sentences, valid := languageToSentence[language].([]interface{})
				if !valid {
					ch <- errors.New("Unable to parse JSON for " + language + " into a string array")
				}

				// Add to list
				dataset.AddList(sentences)
			}
		}(jsonFile)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Collect errors
	for err := range ch {
		if err != nil {
			return nil, err
		}
	}

	return dataset, nil
}
