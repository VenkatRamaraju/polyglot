#!/usr/bin/env python3
import requests
import json
import sys

def test_tokenizer_api():
    """
    Test tokenizer API by encoding a string, then decoding it back.
    
    This test:
    1. Sends a string to the /encode endpoint
    2. Gets back tokens (int list) and token_texts (string list)
    3. Sends the tokens to the /decode endpoint
    4. Verifies the decoded string matches the original
    """
    # Server URL
    base_url = "http://localhost:8080"
    
    # Test string to encode
    test_string = "Hello, this is a test of the tokenizer API!"
    print(f"Original string: {test_string}")
    
    # Step 1: Encode the string
    print("\n--- Testing /encode endpoint ---")
    try:
        encode_response = requests.post(
            f"{base_url}/encode",
            data=json.dumps(test_string),
            headers={"Content-Type": "application/json"}
        )
        encode_response.raise_for_status()  # Raise exception for error status codes
        
        result = encode_response.json()
        tokens = result["tokens"]
        token_texts = result["token_texts"]
        
        print(f"Encoded tokens: {tokens}")
        print(f"Token texts: {token_texts}")
        
    except requests.exceptions.RequestException as e:
        print(f"Error during encode request: {e}")
        sys.exit(1)
    
    # Step 2: Decode the tokens
    print("\n--- Testing /decode endpoint ---")
    try:
        decode_response = requests.post(
            f"{base_url}/decode",
            json={"tokens": tokens},
            headers={"Content-Type": "application/json"}
        )
        decode_response.raise_for_status()
        
        decoded_string = decode_response.json()
        print(f"Decoded string: {decoded_string}")
        
        # Verify the result
        if decoded_string == test_string:
            print("\n✅ SUCCESS: The decoded string matches the original string!")
        else:
            print("\n❌ FAILURE: The decoded string does NOT match the original string!")
            print(f"Original: {test_string}")
            print(f"Decoded: {decoded_string}")
            
    except requests.exceptions.RequestException as e:
        print(f"Error during decode request: {e}")
        sys.exit(1)

if __name__ == "__main__":
    print("BPE Tokenizer API Test")
    print("=====================\n")
    
    # Check if server is running
    try:
        requests.get("http://localhost:8080")
        print("Server is accessible")
    except requests.exceptions.ConnectionError:
        print("ERROR: Cannot connect to server at http://localhost:8080")
        print("Make sure the server is running before executing this test.")
        sys.exit(1)
        
    test_tokenizer_api() 