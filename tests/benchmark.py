#!/usr/bin/env python3
import json, random, time, logging
from pathlib import Path
import requests
from datasets import load_dataset
from tqdm import tqdm
import pprint

LANGUAGES = {
    "en": "English",
    "he": "Hebrew",
    "bn": "Bengali",
    "vi": "Vietnamese",
    "ko": "Korean",
    "ar": "Arabic",
    "ru": "Russian",
    "th": "Thai",
    # "zh": "Chinese",
    "ja": "Japanese",
}

# --------------------------------------------------------------------------- #
def sample_sentences(lang_code: str, n: int = 100):
    # Use streaming to avoid downloading the entire dataset
    ds = load_dataset("statmt/cc100", lang=lang_code, streaming=True, trust_remote_code=True)
    sentences = []
    # Only take the first n sentences
    for i, row in enumerate(ds["train"]):
        if i >= n:
            break
        sentences.append(row["text"])
    return sentences

# --------------------------------------------------------------------------- #
def encode(text: str):
    # Send the text directly as a JSON string, not as {"text": text}
    headers = {'Content-Type': 'application/json'}
    response = requests.post(
        "http://localhost:8080/encode", 
        data=json.dumps(text),  # Convert text to JSON string
        headers=headers,
        timeout=5
    )
    return response.json() if response.status_code == 200 else response

def decode(tokens):
    response = requests.post(
        "http://localhost:8080/decode", 
        json={"tokens": tokens},
        timeout=5
    )
    return response.json() if response.status_code == 200 else response

# --------------------------------------------------------------------------- #
def main():
    results = {}
    for code, name in LANGUAGES.items():
        print(f"Processing {name} ({code})...")
        sentences = sample_sentences(code)

        # Process metrics for this language
        total_tokens = 0
        total_words = 0
        total_characters = 0
        total_seconds = 0.0

        # Start processing for one language
        start = time.time()

        # Convert to one blob
        testing_data = " ".join(sentences)
        encoded_response = encode(testing_data)

        # For calculating fertility
        total_tokens += len(encoded_response["tokens"])

        # Calculate total time
        total_seconds += encoded_response["computation_seconds"]

        # For calculating compression
        total_characters += len(testing_data)
            
        # Create results
        results[code] = {
            "compression_ratio": total_characters / total_tokens,
            "tokens_per_second": total_tokens / total_seconds
        }

        print(round(total_characters / total_tokens, 2))
        print(round(total_tokens / total_seconds, 2))
        print("Done with", code, "in", time.time()-start)
        print("=" * 100)

    pprint.pprint(results)
    

if __name__ == "__main__":
    main()
