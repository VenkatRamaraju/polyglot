package bpe

import (
	"errors"
	"fmt"
	"sync"
)

// getData retrieves all sentences from S3
func getData() (*dataDataset, error) {
	dataDataset := &dataDataset{pdMutex: &sync.Mutex{}}

	// Get files
	asJSONFiles, err := listS3Keys("tknzr", "us-east-1")
	if err != nil {
		return nil, fmt.Errorf("unable to pull s3 files: %w", err)
	}

	// Get training dataset - we can populate map in parallel
	var dWg sync.WaitGroup
	ch := make(chan error)
	for _, sJSONFile := range asJSONFiles {
		dWg.Add(1)
		go func(qsFileName string) {
			// Defer completion of routine
			defer dWg.Done()

			// Get the file contents
			mapLanguageToSentence, err := fetchJSONFromS3("tknzr", qsFileName)
			if err != nil {
				ch <- err
				return
			}

			// Iterate over all languages
			for sLanguage := range mapLanguageToSentence {
				// Grab sentence list
				dLanguages, valid := mapLanguageToSentence[sLanguage].([]interface{})
				if !valid {
					ch <- errors.New("Unable to parse JSON for " + sLanguage + " into a string array")
				}

				// Add to list
				dataDataset.AddList(dLanguages)
			}
		}(sJSONFile)
	}
	go func() {
		dWg.Wait()
		close(ch)
	}()

	// Collect errors
	for err := range ch {
		if err != nil {
			return nil, err
		}
	}

	return dataDataset, nil
}
