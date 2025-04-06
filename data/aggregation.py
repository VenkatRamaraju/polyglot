#!/usr/bin/env python

"""
Aggregation

Pull data from various HuggingFace Datasets
"""

# Imports
from datasets import load_dataset
import boto3
import os
import json
import time

# Global variables
datasets_info = {
    "English": ("opus100", "en-fr"),
    "Spanish": ("opus100", "en-es"),
    "Turkish": ("opus100", "en-tr"),
    "Vietnamese": ("opus100", "en-vi"),
    "Telugu": ("opus100", "en-te"),
    "Arabic": ("opus100", "ar-en"),
    "Russian": ("opus100", "en-ru"),
    "Hindi": ("opus100", "en-hi"),
    "Chinese": ("opus100", "en-zh"),
    "Japanese": ("opus100", "en-ja")
}
language_code_map = {
    "English": "en",
    "Spanish": "es",
    "Turkish": "tr",
    "Vietnamese": "vi",
    "Telugu": "te",
    "Arabic": "ar",
    "Russian": "ru",
    "Hindi": "hi",
    "Chinese": "zh",
    "Japanese": "ja"
}
MAX_EXAMPLES = 50


def upload_to_s3(language_to_sentences: dict):
    """
    Upload to S3

    Args:
        language_to_sentences: Dictionary of language to sentence set mapping
    """

    # Upload
    s3 = boto3.client('s3', aws_access_key_id=os.environ["AWS_ACCESS_KEY_ID"],
                    aws_secret_access_key=os.environ["AWS_SECRET_ACCESS_KEY"])
    json_data = json.dumps(language_to_sentences)
    s3.put_object(Bucket="tknzr", Key="raw.json", Body=json_data)
    print("Upload complete...")



def get_data() -> dict:
    """
    Get the sentence data from the opus dataset

    Returns:
        dict: Language to sentences mapping
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


def main():
    """
    Orchestrate the data collection flow
    """
    # Grab data
    start = time.time()
    sentences = get_data()
    print(time.time() - start)

    # Upload to S3
    upload_to_s3(sentences)



if __name__ == "__main__":
    main()
