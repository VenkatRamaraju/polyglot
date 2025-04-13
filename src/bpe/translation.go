package bpe

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"normalize"
	"os"
	"strconv"
	"strings"
	"sync"
)

// helper: convert string key "1,2" back to [2]int64
func stringToKey(s string) ([2]int64, error) {
	asComponents := strings.Split(s, ",")
	if len(asComponents) != 2 {
		return [2]int64{}, fmt.Errorf("invalid key format: %s", s)
	}
	lFirst, err := strconv.ParseInt(asComponents[0], 10, 64)
	if err != nil {
		return [2]int64{}, fmt.Errorf("invalid int64 in key: %s", asComponents[0])
	}
	lSecond, err := strconv.ParseInt(asComponents[1], 10, 64)
	if err != nil {
		return [2]int64{}, fmt.Errorf("invalid int64 in key: %s", asComponents[1])
	}
	return [2]int64{lFirst, lSecond}, nil
}

// ReadMapFromJSONFile reads a JSON file with string keys and returns map[[2]int64]int
func ReadMapFromJSONFile(sFileName string) (map[[2]int64]int, error) {
	abData, err := os.ReadFile(sFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// First unmarshal into map[string]int
	var mapJSON map[string]int
	if err := json.Unmarshal(abData, &mapJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Convert keys back to [2]int64
	mapPairs := make(map[[2]int64]int, len(mapJSON))
	for sKey, iValue := range mapJSON {
		sFormattedKey, err := stringToKey(sKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key '%s': %w", sKey, err)
		}
		mapPairs[sFormattedKey] = iValue
	}

	return mapPairs, nil
}

// helper: convert [2]int64 to a string key
func keyToString(alKey [2]int64) string {
	return strconv.FormatInt(alKey[0], 10) + "," + strconv.FormatInt(alKey[1], 10)
}

// WriteMergesMapToJSONFile converts a map[[2]int64]int to a string-keyed map and writes it as JSON
func WriteMergesMapToJSONFile(mapMerges *Merges, sFilePath string) error {
	// Convert merges to JSON-friendly format
	mapMergesJSON := make(map[string]int64, len(mapMerges.mapMerges))
	for alKeys, iValue := range mapMerges.mapMerges {
		mapMergesJSON[keyToString(alKeys)] = iValue
	}

	// Create overall JSON Map
	mapJSON := make(map[string]interface{})
	mapJSON["merges"] = mapMergesJSON
	mapJSON["ordering"] = mapMerges.alKeys

	// Marshal without indentation (compressed format)
	abData, err := json.Marshal(mapJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	// Write to file
	if err := os.WriteFile(sFilePath, abData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// getSmallestMintedTokenPair: For a dataset, get the pair corresponding to the smallest minted token
func getSmallestMintedTokenPair(dataStatistics *dataStatistics, mapMerges map[string]interface{}) ([2]int64, bool) {
	// state variables
	tfFound := false
	var minToken int64
	minToken = math.MaxInt64
	var minPair [2]int64

	// iterate over sentence statistics
	for alPair := range dataStatistics.mapPairFrequency {
		// Check if a minted token exists
		lMintedToken, tfOK := mapMerges[keyToString(alPair)]
		if !tfOK {
			continue
		}

		// Found something (prevents break condition in the caller)
		tfFound = true

		// track the lowest minted token
		if int64(lMintedToken.(float64)) < minToken {
			minToken = int64(lMintedToken.(float64))
			minPair[0] = int64(alPair[0])
			minPair[1] = int64(alPair[1])
		}
	}
	return minPair, tfFound
}

// recursively get the characters that make up this set, return as string
func getCharacterComposition(token int64, mapMerges map[string]interface{}) ([]string, error) {
	// iterate over map
	var asComponents []string
	for sPair, interfaceToken := range mapMerges {
		// convert to long
		fToken, tfOK := interfaceToken.(float64)
		if !tfOK {
			return nil, errors.New("failed to convert token to float")
		}
		lToken := int64(fToken)

		// check if found
		if token == lToken {
			// this token is composed of many tokens
			alPair, err := stringToKey(sPair)
			if err != nil {
				return nil, fmt.Errorf("failed to convert string to pair: %w", err)
			}

			// get sub-components
			asSubComponents1, err := getCharacterComposition(alPair[0], mapMerges)
			if err != nil {
				return nil, fmt.Errorf("failed to first pair to its subcomponents: %w", err)
			}
			asSubComponents2, err := getCharacterComposition(alPair[1], mapMerges)
			if err != nil {
				return nil, fmt.Errorf("failed to convert second pair to its subcomponents: %w", err)
			}

			// add all components
			asComponents = append(asComponents, asSubComponents1...)
			asComponents = append(asComponents, asSubComponents2...)

			// we have broken down this token, complete
			break
		}
	}

	if len(asComponents) > 0 {
		return asComponents, nil
	} else {
		return []string{string(token)}, nil
	}
}

// ListToTokens: list tokens to character sets
func ListToTokens(tokenList []int64, mapMerges map[string]interface{}) ([]string, error) {
	// convert every token to its character
	var asTokens []string
	for iIndex := range tokenList {
		// get list of characters for this one token
		asComponents, err := getCharacterComposition(tokenList[iIndex], mapMerges)
		if err != nil {
			return nil, fmt.Errorf("unable to get components of a token: %w", err)
		}

		// add to token list
		asTokens = append(asTokens, strings.Join(asComponents, ""))

	}

	return asTokens, nil
}

// Encode: convert a string to a token list (integers)
func Encode(mapMerges map[string]interface{}, sInput string) ([]int64, error) {
	// normalize string
	sInput = normalize.Normalize(sInput)

	// Convert input to unicode integers
	var unicodePoints []int64
	for _, r := range sInput {
		unicodePoints = append(unicodePoints, int64(r))
	}

	// create dataset
	dataset := &dataDataset{
		aalSentences: [][]int64{unicodePoints},
		pdMutex:      &sync.Mutex{},
	}

	// get stats
	var dataMergeStatistics *dataStatistics

	// encoding loop
	for {
		// initialize statistics
		dataMergeStatistics = &dataStatistics{
			mapPairFrequency: make(map[[2]int64]int),
			pdMutex:          &sync.Mutex{},
		}

		// populate statistics
		err := countStatistics(dataMergeStatistics, dataset)
		if err != nil {
			return nil, fmt.Errorf("failed to generate merge pairs: %w", err)
		}

		// Get the smallest minted token from statistics
		alPair, tfOK := getSmallestMintedTokenPair(dataMergeStatistics, mapMerges)
		if !tfOK {
			break
		}

		// Replace all instances of alPair in the dataset and reassign the dataset for subsequent iteration
		replace(alPair, int64(mapMerges[keyToString(alPair)].(float64)), dataset)
	}
	return dataset.aalSentences[0], nil
}

// Decode: convert a token list (integers) to a string
func Decode(mapTokenizer map[string]interface{}, tokens []int64) (string, error) {
	// get the highest token
	mapOrdering, tfOK := mapTokenizer["ordering"]
	if !tfOK {
		return "", errors.New(`"ordering" not found in merges map`)
	}
	aalInsertOrder, tfOK := mapOrdering.([]interface{})
	if !tfOK {
		return "", errors.New(`"ordering" map has unexpected structure`)
	}
	if len(aalInsertOrder) == 0 {
		return "", errors.New(`"ordering" map has no merges`)
	}
	lMaxToken := aalInsertOrder[len(aalInsertOrder)-1]

	// convert to key
	alTokens, tfOK := lMaxToken.([]interface{})
	if !tfOK {
		return "", errors.New("highest token is not inferrable")
	}
	fFirstToken, tfOK := alTokens[0].(float64)
	if !tfOK {
		return "", errors.New("malformed first token in merges map")
	}
	fSecondToken, tfOK := alTokens[1].(float64)
	if !tfOK {
		return "", errors.New("malformed second token in merges map")
	}

	// initial population
	dataMerges, tfOK := mapTokenizer["merges"]
	if !tfOK {
		return "", errors.New("map merges does not exist")
	}
	mapMerges, tfOK := dataMerges.(map[string]interface{})
	if !tfOK {
		return "", errors.New("Merges map type is incorrect")
	}

	// get the max token
	dataMaxToken, tfOK := mapMerges[keyToString([2]int64{int64(fFirstToken), int64(fSecondToken)})]
	if !tfOK {
		return "", errors.New("no max token exists")
	}
	fMaxToken, tfOK := dataMaxToken.(float64)
	if !tfOK {
		return "", errors.New("token type is not int/float")
	}
	iMaxToken := int64(fMaxToken)

	// populate the map with the basic mapping before overrides
	iIndex := int64(0)
	mapTokens := make(map[int64]string)
	for iIndex < iMaxToken {
		mapTokens[iIndex] = string(iIndex)
		iIndex += 1
	}

	// override some with the minted tokens
	for _, lMintedKey := range aalInsertOrder {
		alTokens, tfOK = lMintedKey.([]interface{})
		if !tfOK {
			return "", errors.New("tokens type cannot be inferred")
		}
		if len(alTokens) != 2 {
			return "", errors.New("cncorrect token count")
		}
		fFirstToken, tfOK = alTokens[0].(float64)
		if !tfOK {
			return "", errors.New("cannot convert first token to float")
		}
		fSecondToken, tfOK := alTokens[1].(float64)
		if !tfOK {
			return "", errors.New("cannot convert second token to float")
		}

		// get token
		dataMerges, tfOK := mapTokenizer["merges"]
		if !tfOK {
			return "", errors.New("map merges does not exist")
		}
		mapMerges, tfOK := dataMerges.(map[string]interface{})
		if !tfOK {
			return "", errors.New("merges map type is incorrect")
		}
		sKey := keyToString([2]int64{int64(fFirstToken), int64(fSecondToken)})
		dataMintedToken, tfOK := mapMerges[sKey]
		if !tfOK {
			return "", errors.New("minted token does not exist")
		}
		fMintedToken, tfOK := dataMintedToken.(float64)
		if !tfOK {
			return "", errors.New("minted token is not float")
		}

		// override
		mapTokens[int64(fMintedToken)] = string(int64(fFirstToken)) + string(int64(fSecondToken))
	}

	// decode overall string using new mapping
	var sResult string
	for _, lToken := range tokens {
		sResult += string(mapTokens[lToken])
	}

	return sResult, nil
}

// [Test Function] EncodeDecode converts a string to an integer list and back to a string to demonstrate the validity of BPE
func EncodeDecode(sFilePath string) error {
	// read merges map into json
	pdFile, err := os.Open(sFilePath)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}
	defer pdFile.Close()

	// Create a map to store the JSON data
	var mapMerges map[string]interface{}

	// Decode the JSON data into the map
	pdDecoder := json.NewDecoder(pdFile)
	err = pdDecoder.Decode(&mapMerges)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	// encode a string
	sInput := "there is a lot of work to do"
	alEncoded, err := Encode(mapMerges["merges"].(map[string]interface{}), sInput)
	if err != nil {
		return fmt.Errorf("unable to encode: %w", err)
	}

	// convert to character set
	asTokens, err := ListToTokens(alEncoded, mapMerges["merges"].(map[string]interface{}))
	if err != nil {
		return fmt.Errorf("unable to convert to token list: %w", err)
	}

	// decode it
	sDecoded, err := Decode(mapMerges, alEncoded)
	if err != nil {
		return fmt.Errorf("failed to decode list: %w", err)
	}

	fmt.Println("Original String:", sInput)
	fmt.Println("Encoded List:", alEncoded)
	fmt.Println("Token list:")
	var asFormatted []string
	for _, sToken := range asTokens {
		asFormatted = append(asFormatted, "\""+sToken+"\"")
	}
	fmt.Println(asFormatted)
	fmt.Println("Encoded characters:", alEncoded)
	fmt.Println("Decoded String:", sDecoded)
	fmt.Println("Encode equals Decode:", sInput == sDecoded)

	return nil
}
