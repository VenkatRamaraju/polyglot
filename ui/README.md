# Tokenizer UI

A simple web interface for interacting with the language model tokenizer API.

## Features

- Encode text into tokens and view the tokenization visually
- Decode token IDs back to text
- Simple, clean interface that matches the specification

## Setup

1. Make sure you have Node.js installed on your system
2. Navigate to the `ui` directory
3. Run the UI server:

```bash
node server.js
```

4. The UI will be available at http://localhost:3000

## Usage

### Start the Backend API Server

Before using the UI, make sure the backend server is running:

```bash
# From the project root directory
go run main.go
```

This will start the backend API server on port 8080.

### Using the UI

1. **Encode Text to Tokens**:
   - Enter text in the top text area
   - Click "Encode" button
   - View the tokens, token IDs, and counts

2. **Decode Tokens to Text**:
   - Enter comma-separated token IDs in the bottom text area
   - Click "Decode" button
   - View the resulting decoded text

## Development

The UI consists of three main files:

- `index.html` - The HTML structure
- `styles.css` - CSS styling
- `script.js` - JavaScript functionality

The UI uses a simple Node.js server (`server.js`) to serve the static files and handle CORS issues. 