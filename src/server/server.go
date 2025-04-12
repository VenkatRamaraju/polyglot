package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"bpe"
)

// Global variable to store the merges map
var mapMerges map[string]interface{}
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
		mapMerges, err = loadMergesMap()
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

// loadMergesMap loads the merges map from the JSON file
func loadMergesMap() (map[string]interface{}, error) {
	// Read merges map from JSON file
	pdFile, err := os.Open("artifacts/merges.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open merges file: %w", err)
	}
	defer pdFile.Close()

	// Create a map to store the JSON data
	var artifactsMap map[string]interface{}

	// Decode the JSON data into the map
	decoder := json.NewDecoder(pdFile)
	if err = decoder.Decode(&artifactsMap); err != nil {
		return nil, fmt.Errorf("failed to decode merges map: %w", err)
	}

	return artifactsMap, nil
}

// Response structure for the encode endpoint
type EncodeResponse struct {
	Tokens     []int64  `json:"tokens"`
	TokenTexts []string `json:"token_texts"`
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

	// Create response
	dataResponse := EncodeResponse{
		Tokens:     alEncodedTokens,
		TokenTexts: asTokenTexts,
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
	sDecodedString, err := bpe.Decode(mapMerges, request.Tokens)
	if err != nil {
		http.Error(dataWriter, fmt.Sprintf("Decoding error: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the decoded string as the HTTP response
	dataWriter.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(dataWriter).Encode(sDecodedString); err != nil {
		http.Error(dataWriter, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
