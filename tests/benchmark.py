#!/usr/bin/env python3
import json, random, logging
from pathlib import Path
import requests
from datasets import load_dataset
from tqdm import tqdm
import pprint
import tiktoken
from transformers import AutoTokenizer
from util import count_words_batch

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
transformers_tokenizer = AutoTokenizer.from_pretrained("gpt2")
sentencepiece_tokenizer = AutoTokenizer.from_pretrained("t5-base")
mbert_tokenizer = AutoTokenizer.from_pretrained("bert-base-multilingual-cased")
xlm_tokenizer = AutoTokenizer.from_pretrained("xlm-roberta-base")

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
    headers = {'Content-Type': 'application/json'}
    response = requests.post(
        "http://localhost:8080/encode", 
        data=json.dumps(text),
        headers=headers,
        timeout=5
    )
    return response.json() if response.status_code == 200 else response

def tiktoken_encode(tokens):
    return tiktoken_encoder.encode(tokens)

def transformers_encode(text):
    return transformers_tokenizer.encode(text)

def sentencepiece_encode(text):
    return sentencepiece_tokenizer.encode(text)

def bert_encode(text):
    return mbert_tokenizer.encode(text)

def xlm_encode(text):
    return xlm_tokenizer.encode(text)

# --------------------------------------------------------------------------- #
def main():
    results = {}
    tokenizers = ["polyglot", "tiktoken", "transformers", "sentencepiece", "bert", "xlm"]
    for code, name in LANGUAGES.items():
        print(f"Processing {name} ({code})...")
        sentences = sample_sentences(code)

        # Total words
        total_words = sum(count_words_batch(sentences, code, False))

        # Per tokenizer
        for tokenizer in tokenizers:
            # Process metrics for this language
            total_tokens = 0
            total_characters = 0

            # Start processing for one language
            for sentence in tqdm(sentences):
                # Encode
                if tokenizer == "polyglot":
                    encoded_response = polyglot_encode(sentence)
                    total_tokens += len(encoded_response["tokens"])
                elif tokenizer == "tiktoken":
                    encoded_response = tiktoken_encode(sentence)
                    total_tokens += len(encoded_response)
                elif tokenizer == "transformers":
                    encoded_response = transformers_encode(sentence)
                    total_tokens += len(encoded_response)
                elif tokenizer == "sentencepiece":
                    encoded_response = sentencepiece_encode(sentence)
                    total_tokens += len(encoded_response)
                elif tokenizer == "bert":
                    encoded_response = bert_encode(sentence)
                    total_tokens += len(encoded_response)
                elif tokenizer == "xlm":
                    encoded_response = xlm_encode(sentence)
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
            results[code][tokenizer]["token_fertility"] = round(total_tokens / total_words, 2)
    
    pprint.pprint(results)

if __name__ == "__main__":
    main()
