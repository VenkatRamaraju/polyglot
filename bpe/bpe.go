package bpe

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

func Encode(text string) {

	var languageMap = map[string]string{
		"ar": "Arabic",
		"bg": "Bulgarian",
		"de": "German",
		"el": "Greek",
		"en": "English",
		"es": "Spanish",
		"fr": "French",
		"hi": "Hindi",
		"ru": "Russian",
		"tr": "Turkish",
		"ur": "Urdu",
		"vi": "Vietnamese",
		"zh": "Chinese",
	}

	url := "https://huggingface.co/datasets/xnli/resolve/main/xnli.test.tsv"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var buffer []byte
	resp.Body.Read(buffer)
	fmt.Println(string(buffer))

	scanner := bufio.NewScanner(resp.Body)
	found := make(map[string]bool)

	// Skip header
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		lang := parts[0]
		premise := parts[6] // This is the main sentence

		if _, ok := languageMap[lang]; ok && !found[lang] {
			fmt.Printf("%s (%s): %s\n", languageMap[lang], lang, premise)
			found[lang] = true
		}
		if len(found) == len(languageMap) {
			break
		}
	}
}
