#!/usr/bin/env python

"""
Aggregation

Pull data from various HuggingFace Datasets
"""

# Imports
import os
import json
import time
from datasets import load_dataset
from math import ceil
from typing import List, Dict
import boto3

# Global variables
datasets_info = {
    "English": ("opus100", "en-fr"),
    "Hebrew": ("opus100", "en-he"),
    "Bengali": ("opus100", "bn-en"),
    "Vietnamese": ("opus100", "en-vi"),
    "Korean": ("opus100", "en-ko"),
    "Arabic": ("opus100", "ar-en"),
    "Russian": ("opus100", "en-ru"),
    "Thai": ("opus100", "en-th"),
    "Chinese": ("opus100", "en-zh"),
    "Japanese": ("opus100", "en-ja")
}
language_code_map = {
    "English": "en",
    "Hebrew": "he",
    "Bengali": "bn",
    "Vietnamese": "vi",
    "Korean": "ko",
    "Arabic": "ar",
    "Russian": "ru",
    "Thai": "th",
    "Chinese": "zh",
    "Japanese": "ja"
}
MAX_EXAMPLES = 1000000
BATCH_SIZE = 10000

def upload_to_s3(language_to_sentences: List[Dict[str, List[str]]]):
    """
    Uploads a batch of sentences to S3.

    Args:
        language_to_sentences: A list of dictionaries mapping languages to their respective sentences.
    """
    i = 0
    for batch in language_to_sentences:
        # Initialize S3 client
        s3 = boto3.client('s3', aws_access_key_id=os.environ["AWS_ACCESS_KEY_ID"],
                          aws_secret_access_key=os.environ["AWS_SECRET_ACCESS_KEY"])
        json_data = json.dumps(batch)
        # Upload the JSON data to S3
        s3.put_object(Bucket="tknzr", Key="raw_" + str(i) + ".json", Body=json_data)
        i += 1
    print("Upload complete.")


def get_data() -> dict:
    """
    Retrieves sentence data from various datasets.

    Returns:
        dict: A mapping of languages to their respective sentences.
    """
    # Initialize mapping
    language_to_sentence = {}

    # Load each dataset
    for language, (dataset_name, config) in datasets_info.items():
        try:
            # Load dataset with configuration if provided.
            if config is None:
                dataset = load_dataset(dataset_name)
            else:
                dataset = load_dataset(dataset_name, config)
        except Exception as e:
            print(f"Failed to load {dataset_name} with config {config}: {e}")
            continue

        # Get a split, train is default
        split_name = "train" if "train" in dataset.keys() else list(dataset.keys())[0]
        examples = dataset[split_name][:MAX_EXAMPLES]

        # Add to map
        language_to_sentence[language] = [translation[language_code_map[language]] for translation in examples["translation"]]

    return language_to_sentence


def batch_up(data: Dict[str, List[str]]) -> List[Dict[str, List[str]]]:
    """
    Batch up

    Args:
       data: Dict[str, List[str]]: Unbatche dictionary

    Returns:
        Batched values, each one with multiple language of a certain batch size
    """
    max_len = max((len(v) for v in data.values()), default=0)
    if max_len == 0:
        return []

    num_batches = ceil(max_len / BATCH_SIZE)
    batched = []

    for i in range(num_batches):
        batch = {}
        any_data = False
        for key, values in data.items():
            start = i * BATCH_SIZE
            end = start + BATCH_SIZE
            sliced = values[start:end]
            batch[key] = sliced
            if sliced:
                any_data = True
        if any_data:
            batched.append(batch)

    return batched


def main():
    """
    Orchestrate the data collection flow
    """
    # Grab data
    start = time.time()
    sentences = get_data()
    print(time.time() - start)

    # batch up the JSONs
    json_batches = batch_up(sentences)

    # Upload to S3
    upload_to_s3(json_batches)



if __name__ == "__main__":
    main()
