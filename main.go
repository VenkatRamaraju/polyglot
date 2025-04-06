package main

import (
	"bpe"
	"normalize"
)

func main() {
	// normalize the training text
	normalizedText := normalize.Normalize("నేను ఒక టోకనైజర్ రాయబోతున్నాను. నేను ఒక టోకనైజర్ రాయబోతున్నాను. నేను ఒక టోకనైజర్ రాయబోతున్నాను. నేను ఒక టోకనైజర్ రాయబోతున్నాను.")

	// encode the normalized text
	bpe.Encode(normalizedText)
}
