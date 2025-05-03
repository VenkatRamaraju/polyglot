# Polyglot

Polyglot is a high-performance multilingual tokenizer, built entirely from scratch in Go, that efficiently compresses text from 10 diverse languages using the Byte-Pair Encoding (BPE) algorithm. The system supports English, Hebrew, Bengali, Vietnamese, Korean, Arabic, Russian, Thai, Chinese, and Japanese.

## Polyglot Metrics

- Compression Ratio: *TBD*
- Vocabulary Size: *TBD*
- Total Training Corpus: 432,584,912 characters (10M sentences)

##  Benchmarking

The tokenizer is evaluated against five state-of-the-art (SOTA) baselines: Tiktoken, Transformers, SentencePiece, mBERT, and XLM. A total of 100,000 sentences—10,000 per language across 10 languages—were sampled from the statmt/cc100 dataset. For each tokenizer and language, the mean compression ratio and token fertility were computed over the corresponding 10,000 sentences.

#### Compression Ratio

#### Token fertility

#### Ranking

## Training

- **Dataset**: The tokenizer was trained on 10M sentences from the [opus-100 dataset](https://huggingface.co/datasets/Helsinki-NLP/opus-100), with 1M sentences per language. The language set was carefully selected to incorporate a sufficiently diverse range of scripts in our training dataset.
- **Training Process**: The current version has a compression ratio of 3.0. Training runs are in progress to push this to 5.0.
- **Implementation**: Data aggregation and formatting were implemented in Python. The core BPE algorithm and server were written in Go. Training data was chunked and streamed from S3 for efficient processing on machines of various sizes.

## Deployment

Deploy Polyglot locally using Docker with the following commands:

```bash
# Build the Docker image
docker build -t polyglot-app .

# Run the container
docker run -p 8080:8080 -p 3000:3000 polyglot-app
```

Navigate to [localhost:3000](http://localhost:3000/) to interface with the tool.

## Frontend

The `ui` directory contains an intuitive user interface that provides the following capabilities:

- Text input for tokenization
- Visualization of tokenized segments and their corresponding integer representations
- Decoding functionality to reconstruct the original text
- Real-time metrics displaying compression ratio, token-to-character counts for performance analysis, and computation times.

## Backend

The backend exposes two RESTful endpoints:

- **`/encode`**: Processes input text and returns the corresponding token sequence with text representations
- **`/decode`**: Accepts a token sequence and reconstructs the original text

## License

This project is licensed under the MIT License.
