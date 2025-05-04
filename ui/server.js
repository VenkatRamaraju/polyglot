const http = require('http');
const fs = require('fs');
const path = require('path');
const { createProxyServer } = require('http-proxy');

const PORT = 3000;
const API_PORT = 8080;

// Create a proxy server for API requests
const apiProxy = createProxyServer();

const MIME_TYPES = {
    '.html': 'text/html',
    '.css': 'text/css',
    '.js': 'text/javascript',
    '.json': 'application/json',
    '.png': 'image/png',
    '.jpg': 'image/jpeg',
    '.gif': 'image/gif',
    '.svg': 'image/svg+xml',
    '.ico': 'image/x-icon'
};

const server = http.createServer((req, res) => {
    // Set CORS headers to allow requests to the API
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
    
    // Handle preflight OPTIONS request
    if (req.method === 'OPTIONS') {
        res.writeHead(204);
        res.end();
        return;
    }
    
    // Proxy API requests to the Go backend
    if (req.url.startsWith('/api')) {
        console.log(`Proxying API request: ${req.url}`);
        // Rewrite the URL to remove the /api prefix
        req.url = req.url.replace(/^\/api/, '');
        // Proxy to the backend
        apiProxy.web(req, res, { 
            target: `http://localhost:${API_PORT}`,
            ignorePath: false
        }, (err) => {
            console.error('Proxy error:', err);
            res.writeHead(500);
            res.end('Proxy Error: Cannot connect to backend server. Please make sure the Go server is running on port 8080.');
        });
        return;
    }
    
    // Get the file path
    let filePath = path.join(__dirname, req.url === '/' ? 'index.html' : req.url);
    
    // Get the file extension
    const extname = path.extname(filePath);
    
    // Set the content type
    const contentType = MIME_TYPES[extname] || 'text/plain';
    
    // Read the file
    fs.readFile(filePath, (err, content) => {
        if (err) {
            if (err.code === 'ENOENT') {
                // Page not found
                fs.readFile(path.join(__dirname, '404.html'), (err, content) => {
                    res.writeHead(404, { 'Content-Type': 'text/html' });
                    res.end(content, 'utf8');
                });
            } else {
                // Server error
                res.writeHead(500);
                res.end(`Server Error: ${err.code}`);
            }
        } else {
            // Success
            res.writeHead(200, { 'Content-Type': contentType });
            res.end(content, 'utf8');
        }
    });
});

server.listen(PORT, () => {
    console.log(`Tokenizer UI running at http://localhost:${PORT}/ (API: http://localhost:${API_PORT})`);
});