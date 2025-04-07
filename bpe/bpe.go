package bpe

import (
	"errors"
	"normalize"
)

// convert byte pairs to a map representation
func generateStatistics(sentences []interface{}, merges map[[2]int]int) error {
	for _, sentence := range sentences {
		sentence, valid := sentence.(string)
		if !valid {
			return errors.New("cannnot convert to string")
		}

		// normalize sentence
		sentence = normalize.Normalize(sentence)

		// convert to unicode integers
		var unicodePoints []int
		for _, r := range sentence {
			unicodePoints = append(unicodePoints, int(r))
		}

		// create merge pairs
		for index := range unicodePoints {
			if index+1 >= len(unicodePoints) {
				break
			}

			// create pair if necessary
			pair := [2]int{unicodePoints[index], unicodePoints[index+1]}
			if _, present := merges[pair]; !present {
				merges[pair] = 0
			}

			// add pair
			merges[pair] += 1
		}
	}
	return nil
}
