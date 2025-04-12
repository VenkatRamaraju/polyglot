package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"bpe" // Updated import path for bpe package
)

// Launch the server
func Launch() {
	http.HandleFunc("/encode", encodeHanlder)
	http.HandleFunc("/decode", decodeHandler)

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}

func initialize() (map[string]interface{}, error) {
	// read merges map into json
	pdFile, err := os.Open("artifacts/merges.json")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map: %w", err)
	}
	defer pdFile.Close()

	// Create a map to store the JSON data
	var mapMerges map[string]interface{}

	// Decode the JSON data into the map
	pdDecoder := json.NewDecoder(pdFile)
	err = pdDecoder.Decode(&mapMerges)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map: %w", err)
	}

	return mapMerges, nil
}

// encode a string to its integer list
func encodeHanlder(w http.ResponseWriter, r *http.Request) {
	// Retrieve the input string from the HTTP request
	var input string
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Call bpe.Encode() with the input string
	encodedResult, err := bpe.Encode(nil, "")
	if err != nil {
		http.Error(w, "Encoding error", http.StatusInternalServerError)
		return
	}

	// Return the encoded result as the HTTP response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(encodedResult)
}

// encode a string to its integer list
func decodeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
