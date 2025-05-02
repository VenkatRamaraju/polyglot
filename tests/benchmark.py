#!/usr/bin/env python3
import json, random, time, logging
from pathlib import Path
import requests
from datasets import load_dataset
from tqdm import tqdm
import pprint
import tiktoken

# Global variabes
LANGUAGES = {
    "en": "English",
    "he": "Hebrew",
    "bn": "Bengali",
    "vi": "Vietnamese",
    "ko": "Korean",
    "ar": "Arabic",
    "ru": "Russian",
    "th": "Thai",
    "zh-Hans": "Chinese",
    "ja": "Japanese",
}
tiktoken_encoder = tiktoken.encoding_for_model("gpt-4o")

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
def polyglot_encode(text: str):
    # Send the text directly as a JSON string, not as {"text": text}
    headers = {'Content-Type': 'application/json'}
    response = requests.post(
        "http://localhost:8080/encode", 
        data=json.dumps(text),  # Convert text to JSON string
        headers=headers,
        timeout=5
    )
    return response.json() if response.status_code == 200 else response

def tiktoken_encode(tokens):
    return tiktoken_encoder.encode(tokens)

# --------------------------------------------------------------------------- #
def main():
    results = {}
    tokenizers = ["polyglot", "tiktoken"]
    for code, name in LANGUAGES.items():
        print(f"Processing {name} ({code})...")
        sentences = sample_sentences(code)

        # Per tokenizer
        for tokenizer in tokenizers:
            # Process metrics for this language
            total_tokens = 0
            total_words = 0
            total_characters = 0
            total_seconds = 0.0

            # Start processing for one language
            for sentence in tqdm(sentences):
                # Encode
                if tokenizer == "polyglot":
                    encoded_response = polyglot_encode(sentence)
                    total_tokens += len(encoded_response["tokens"])
                elif tokenizer == "tiktoken":
                    encoded_response = tiktoken_encode(sentence)
                    total_tokens += len(encoded_response)
                else:
                    raise Exception("Unknown tokenizer:", tokenizer)
                
                # For calculating compression
                total_characters += len(sentence)

            # Log results for this tokenizer and language
            if code not in results.keys():
                results[code] = {}
            if tokenizer not in results[code].keys():
                results[code][tokenizer] = {}
                
            # Create results
            results[code][tokenizer]["compression_ratio"] = round(total_characters / total_tokens, 2)
    
    pprint.pprint(results)
    count = 0
    for language in results.keys():
        if results[language]["polyglot"]["compression_ratio"] > results[language]["tiktoken"]["compression_ratio"]:
            count += 1
    print("polyglot is faster for", count, "out of 10")
    

if __name__ == "__main__":
    main()
