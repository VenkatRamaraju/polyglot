package bpe

import (
	"encoding/json"
	"fmt"
	"math"
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

	// Marshal with indentation
	abData, err := json.MarshalIndent(mapJSON, "", "")
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	// Write to file
	if err := os.WriteFile(sFilePath, abData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// get the pair corresponding to the smallest minted token
func getMinimumMintedTokenPair(dataStatistics *dataStatistics, mapMerges map[string]interface{}) ([2]int64, bool) {
	tfFound := false
	var minToken int64
	minToken = math.MaxInt64
	var minPair [2]int64

	for alPair := range dataStatistics.mapPairFrequency {
		// look for this token in the map merges

		// convert to key
		lMintedToken, tfOK := mapMerges[pairToKey(alPair)]

		if !tfOK {
			continue
		}

		tfFound = true

		// track the lowest minted token
		if int64(lMintedToken.(float64)) < minToken {
			minToken = int64(lMintedToken.(float64))
			minPair[0] = int64(alPair[0])
			minPair[1] = int64(alPair[1])
		}

	}

	// fmt.Println("min pair", minPair)

	return minPair, tfFound
}

// convert a string to a token list (integers)
func encode(mapMerges map[string]interface{}, sInput string) ([]int64, error) {
	// run encoding loop

	// Convert to unicode integers
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

	for {
		dataMergeStatistics = &dataStatistics{
			mapPairFrequency: make(map[[2]int64]int),
			pdMutex:          &sync.Mutex{},
		}

		// get merge pairs
		err := generateMergePairs(dataMergeStatistics, dataset)
		if err != nil {
			return nil, fmt.Errorf("failed to generate merge pairs: %w", err)
		}

		fmt.Println(dataMergeStatistics.mapPairFrequency)
		fmt.Println("=====================")

		// get the smallest minted token from stats
		alPair, tfOK := getMinimumMintedTokenPair(dataMergeStatistics, mapMerges)
		if !tfOK {
			break
		}

		// replace all instances of alPair in the dataset and reassign
		replace(alPair, int64(mapMerges[pairToKey(alPair)].(float64)), dataset)

	}

	return dataset.aalSentences[0], nil
}

// convert a token list (integers) to a string
func decode(mapMerges map[string]interface{}, tokens []int64) string {
	// initialize token map
	// mapToken := make(map[int64]string)
	aalInsertOrder := mapMerges["ordering"].([]interface{})
	lMaxToken := aalInsertOrder[len(aalInsertOrder)-1]

	// convert to key
	alTokens := lMaxToken.([]interface{})
	lFirstToken := alTokens[0].(float64)
	lSecondToken := alTokens[1].(float64)
	sKey := strconv.FormatInt(int64(lFirstToken), 10) + "," + strconv.FormatInt(int64(lSecondToken), 10)

	// initial population
	iMaxToken := int64(mapMerges["merges"].(map[string]interface{})[sKey].(float64))
	iIndex := int64(0)
	mapTokens := make(map[int64]string)
	for iIndex < iMaxToken {
		mapTokens[iIndex] = string(iIndex)
		iIndex += 1
	}

	// override some with the minted tokens
	for _, lMintedKey := range aalInsertOrder {
		alTokens = lMintedKey.([]interface{})
		lFirstToken = alTokens[0].(float64)
		lSecondToken = alTokens[1].(float64)

		// get token
		sKey = strconv.FormatInt(int64(lFirstToken), 10) + "," + strconv.FormatInt(int64(lSecondToken), 10)
		iMintedToken := int64(mapMerges["merges"].(map[string]interface{})[sKey].(float64))

		mapTokens[iMintedToken] = string(int64(lFirstToken)) + string(int64(lSecondToken))
	}

	// decode
	var sResult string
	for _, lToken := range tokens {
		sResult += string(mapTokens[lToken])
	}

	return sResult
}

// EncodeDecode converts a string to an integer list and back to a string to demonstrate our algorithms
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
	sEncode := "there is a lot of work to do here"
	alEncoded, err := encode(mapMerges["merges"].(map[string]interface{}), sEncode)
	if err != nil {
		return fmt.Errorf("unable to encode: %w", err)
	}
	fmt.Println(sEncode)
	fmt.Println(alEncoded)
	fmt.Println(decode(mapMerges, alEncoded))

	return nil
}
