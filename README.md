# Polyglot

Polyglot is a high-performance multilingual tokenizer, built entirely from scratch in Go, that efficiently compresses text from 10 diverse languages using the Byte-Pair Encoding (BPE) algorithm. The system supports English, Hebrew, Bengali, Vietnamese, Korean, Arabic, Russian, Thai, Chinese, and Japanese.

## Performance Metrics

- Compression Ratio: *TBD*
- Vocabulary Size: *TBD*
- Total Training Corpus: 432,584,912 tokens (10M sentences)

## Training Methodology

- **Dataset**: The tokenizer was trained on 10M sentences from the [opus-100 dataset](https://huggingface.co/datasets/Helsinki-NLP/opus-100), with 1M sentences per language. The language set was carefully selected to incorporate a sufficiently diverse range of scripts in our training dataset.
- **Training Process**: The current version has a compression ratio of 3.0. As subsequent version will be published with a 5.0 compression ratio.
- **Implementation**: Data aggregation and formatting were implemented in Python. The core BPE algorithm and server were written in Go. Data was chunked and streamed from S3 for efficient processing on machines of various sizes.

## Deployment Instructions

Deploy Polyglot locally using Docker with the following commands:

```bash
# Build the Docker image
docker build -t polyglot-app .

# Run the container
docker run -p 8080:8080 -p 3000:3000 polyglot-app
```

Navigate to [localhost:3000](http://localhost:3000/) to interface with the tool.

## Frontend Interface

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
