package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"bpe"
)

// Global variable to store the merges and decoder maps
var mapMerges map[string]interface{}
var mapDecoder map[int64]string
var pdSync sync.Once

// enableCORS sets the necessary headers for Cross-Origin Resource Sharing
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Launch the server
func Launch() {
	// Pre-load the merges map
	pdSync.Do(func() {
		var err error
		mapMerges, mapDecoder, err = bpe.LoadMaps()
		if err != nil {
			log.Fatalf("Failed to load merges map: %s", err)
		}
	})

	// Set up handlers with CORS middleware
	http.HandleFunc("/encode", encodeHandler)
	http.HandleFunc("/decode", decodeHandler)

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}

// Response structure for the encode endpoint
type EncodeResponse struct {
	Tokens            []int64  `json:"tokens"`
	TokenTexts        []string `json:"token_texts"`
	ComputationTimeMs string   `json:"computation_time_ms"`
}

// encodeHandler handles the /encode endpoint
func encodeHandler(dataWriter http.ResponseWriter, pdRequest *http.Request) {
	// Enable CORS for all requests
	enableCORS(dataWriter)

	// Handle preflight OPTIONS request
	if pdRequest.Method == http.MethodOptions {
		dataWriter.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if pdRequest.Method != http.MethodPost {
		http.Error(dataWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the input string from the HTTP request
	var sInput string
	if err := json.NewDecoder(pdRequest.Body).Decode(&sInput); err != nil {
		http.Error(dataWriter, "Invalid input, expected a JSON string", http.StatusBadRequest)
		return
	}

	// get a part of it
	convertedMap, tfOK := mapMerges["merges"].(map[string]interface{})
	if !tfOK {
		http.Error(dataWriter, "Failed to convert map: merges key not found or not a map", http.StatusInternalServerError)
		return
	}

	// Call bpe.Encode() with the input string
	startTime := time.Now()
	alEncodedTokens, err := bpe.Encode(convertedMap, sInput)
	if err != nil {
		http.Error(dataWriter, fmt.Sprintf("Encoding error: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert tokens to text representations
	asTokenTexts, err := bpe.ListToTokens(alEncodedTokens, convertedMap)
	if err != nil {
		http.Error(dataWriter, fmt.Sprintf("Token text conversion error: %v", err), http.StatusInternalServerError)
		return
	}
	totalComputationTime := time.Since(startTime)

	dataResponse := EncodeResponse{
		Tokens:            alEncodedTokens,
		TokenTexts:        asTokenTexts,
		ComputationTimeMs: bpe.FormatDuration(totalComputationTime),
	}

	// Return the encoded result as the HTTP response
	dataWriter.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(dataWriter).Encode(dataResponse); err != nil {
		http.Error(dataWriter, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// Request structure for the decode endpoint
type DecodeRequest struct {
	Tokens []int64 `json:"tokens"`
}

// decodeHandler handles the /decode endpoint
func decodeHandler(dataWriter http.ResponseWriter, pdRequest *http.Request) {
	// Enable CORS for all requests
	enableCORS(dataWriter)

	// Handle preflight OPTIONS request
	if pdRequest.Method == http.MethodOptions {
		dataWriter.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if pdRequest.Method != http.MethodPost {
		http.Error(dataWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the input tokens from the HTTP request
	var request DecodeRequest
	if err := json.NewDecoder(pdRequest.Body).Decode(&request); err != nil {
		http.Error(dataWriter, "Invalid input, expected a JSON object with 'tokens' field", http.StatusBadRequest)
		return
	}

	// Call bpe.Decode() with the input tokens
	startTime := time.Now()
	sDecodedString, err := bpe.Decode(mapDecoder, request.Tokens)
	if err != nil {
		http.Error(dataWriter, fmt.Sprintf("Decoding error: %v", err), http.StatusInternalServerError)
		return
	}
	totalComputationTime := time.Since(startTime)

	// Return the decoded string as the HTTP response
	dataWriter.Header().Set("Content-Type", "application/json")
	// Create a response that includes both the decoded text and computation time
	response := struct {
		Text              string `json:"text"`
		ComputationTimeMs string `json:"computation_time_ms"`
	}{
		Text:              sDecodedString,
		ComputationTimeMs: bpe.FormatDuration(totalComputationTime),
	}

	if err := json.NewEncoder(dataWriter).Encode(response); err != nil {
		http.Error(dataWriter, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
