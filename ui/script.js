document.addEventListener('DOMContentLoaded', () => {
    // Elements for encoding
    const inputText = document.getElementById('input-text');
    const encodeBtn = document.getElementById('encode-btn');
    const tokenCount = document.getElementById('token-count');
    const charCount = document.getElementById('char-count');
    const tokenIds = document.getElementById('token-ids');
    const tokenDisplay = document.getElementById('token-display');
    const copyToDecodeBtn = document.getElementById('copy-to-decode');

    // Elements for token display toggle
    const displayToggle = document.getElementById('display-toggle');
    const toggleText = document.getElementById('toggle-text');
    const toggleIds = document.getElementById('toggle-ids');

    // Elements for decoding
    const inputTokens = document.getElementById('input-tokens');
    const decodeBtn = document.getElementById('decode-btn');
    const decodedText = document.getElementById('decoded-text');

    // Theme toggle
    const themeSwitch = document.getElementById('theme-switch');

    // Backend API URL - make sure this matches your Go server
    const API_BASE_URL = 'http://localhost:8080';

    // Check if backend is available (silent check)
    checkBackendConnection();

    // Initialize theme based on user preference or local storage
    initializeTheme();

    // Store token data for copying
    let currentTokens = [];

    // Handle display toggle
    displayToggle.addEventListener('change', () => {
        if (displayToggle.checked) {
            // Show Token IDs
            tokenDisplay.classList.remove('active-display');
            tokenDisplay.classList.add('hidden-display');
            tokenIds.classList.remove('hidden-display');
            tokenIds.classList.add('active-display');
            toggleText.classList.remove('toggle-active');
            toggleIds.classList.add('toggle-active');
        } else {
            // Show Token Texts
            tokenIds.classList.remove('active-display');
            tokenIds.classList.add('hidden-display');
            tokenDisplay.classList.remove('hidden-display');
            tokenDisplay.classList.add('active-display');
            toggleIds.classList.remove('toggle-active');
            toggleText.classList.add('toggle-active');
        }
    });

    // Click handlers for toggle labels
    toggleText.addEventListener('click', () => {
        displayToggle.checked = false;
        displayToggle.dispatchEvent(new Event('change'));
    });

    toggleIds.addEventListener('click', () => {
        displayToggle.checked = true;
        displayToggle.dispatchEvent(new Event('change'));
    });

    // Theme toggle handler
    themeSwitch.addEventListener('change', () => {
        if (themeSwitch.checked) {
            document.documentElement.classList.remove('light-theme');
            document.documentElement.classList.add('dark-theme');
            localStorage.setItem('theme', 'dark');
        } else {
            document.documentElement.classList.remove('dark-theme');
            document.documentElement.classList.add('light-theme');
            localStorage.setItem('theme', 'light');
        }
    });

    // Initialize theme based on user preference
    function initializeTheme() {
        const savedTheme = localStorage.getItem('theme');
        if (savedTheme === 'dark') {
            document.documentElement.classList.add('dark-theme');
            document.documentElement.classList.remove('light-theme');
            themeSwitch.checked = true;
        } else {
            document.documentElement.classList.add('light-theme');
            document.documentElement.classList.remove('dark-theme');
            themeSwitch.checked = false;
        }
    }

    // Update character count when typing in the input text
    inputText.addEventListener('input', () => {
        charCount.textContent = inputText.value.length;
    });

    // Add example text on focus if empty
    inputText.addEventListener('focus', () => {
        if (inputText.value.trim() === '') {
            inputText.placeholder = 'Try typing something like "Hello, world!"';
        }
    });

    inputText.addEventListener('blur', () => {
        inputText.placeholder = 'Enter text to tokenize...';
    });

    // Add example tokens on focus if empty
    inputTokens.addEventListener('focus', () => {
        if (inputTokens.value.trim() === '') {
            inputTokens.placeholder = 'Try entering numbers like: 15496, 11, 995';
        }
    });

    inputTokens.addEventListener('blur', () => {
        inputTokens.placeholder = 'Enter token IDs (comma-separated integers)...';
    });

    // Function to check if backend is available (silent check, no UI warnings)
    async function checkBackendConnection() {
        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 2000);
            
            const response = await fetch(`${API_BASE_URL}/encode`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify("test"),
                signal: controller.signal
            });
            
            clearTimeout(timeoutId);
            
            if (response.ok) {
                console.log("Backend connection successful");
            }
        } catch (error) {
            console.warn("Backend connection check failed:", error);
            // Silent failure - no UI warning
        }
    }

    // Function to format token IDs as visual elements
    function formatTokenIds(tokens) {
        // Save token list for copy functionality
        currentTokens = tokens;
        
        // Update the "Copy to Decode" button state
        copyToDecodeBtn.disabled = tokens.length === 0;
        
        // Clear existing content
        tokenIds.innerHTML = '';
        
        // Check if there are tokens to display
        if (tokens.length === 0) {
            const emptyState = document.createElement('div');
            emptyState.className = 'empty-state';
            emptyState.textContent = 'No tokens to display';
            tokenIds.appendChild(emptyState);
            return;
        }
        
        // Create elements with staggered animation
        tokens.forEach((token, index) => {
            const tokenEl = document.createElement('span');
            tokenEl.className = 'token-id';
            tokenEl.textContent = token;
            tokenEl.style.animationDelay = `${index * 20}ms`;
            tokenIds.appendChild(tokenEl);
        });
    }

    // Handle encoding
    encodeBtn.addEventListener('click', async () => {
        const text = inputText.value.trim();
        
        if (!text) {
            showError(inputText, 'Please enter some text to encode');
            return;
        }
        
        try {
            encodeBtn.disabled = true;
            encodeBtn.textContent = 'Encoding...';
            tokenDisplay.innerHTML = '<div class="loading">Processing...</div>';
            tokenIds.innerHTML = '<div class="loading">Loading...</div>';
            copyToDecodeBtn.disabled = true;
            
            const response = await fetch(`${API_BASE_URL}/encode`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(text)
            });
            
            if (!response.ok) {
                throw new Error(`Server error: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            
            // Update token count
            tokenCount.textContent = data.tokens.length;
            
            // Calculate and display compression ratio
            const compressionRatio = (inputText.value.length / data.tokens.length).toFixed(2);
            document.getElementById('compression-ratio').textContent = compressionRatio;
            
            // Display token IDs in a visually appealing format
            formatTokenIds(data.tokens);
            
            // Display colored tokens
            tokenDisplay.innerHTML = '';
            
            // Check if there are tokens to display
            if (data.token_texts.length === 0) {
                const emptyState = document.createElement('div');
                emptyState.className = 'empty-state';
                emptyState.textContent = 'No tokens to display';
                tokenDisplay.appendChild(emptyState);
            } else {
                // Create a single flowing text with highlighted tokens
                const tokenContainer = document.createElement('div');
                tokenContainer.className = 'token-flow';
                
                // Process each token and add it to the flow
                data.token_texts.forEach((token, index) => {
                    const tokenElement = document.createElement('span');
                    tokenElement.className = `token token-${index % 10}`;
                    tokenElement.textContent = token;
                    tokenElement.dataset.tokenId = data.tokens[index];
                    tokenContainer.appendChild(tokenElement);
                });
                
                tokenDisplay.appendChild(tokenContainer);
            }
            
            // Ensure the display matches the toggle state
            displayToggle.dispatchEvent(new Event('change'));
        } catch (error) {
            console.error('Error during encoding:', error);
            tokenIds.innerHTML = '';
            tokenDisplay.innerHTML = '';
            currentTokens = [];
            copyToDecodeBtn.disabled = true;
            
            if (error.message.includes('Failed to fetch') || error.name === 'AbortError') {
                showError(inputText, 'Cannot connect to backend server. Please make sure the Go server is running on port 8080.');
            } else {
                showError(inputText, `Encoding failed: ${error.message}`);
            }
        } finally {
            encodeBtn.disabled = false;
            encodeBtn.textContent = 'Encode';
        }
    });

    // Copy to Decode button functionality
    copyToDecodeBtn.addEventListener('click', () => {
        if (currentTokens.length === 0) return;
        
        // Convert token array to comma-separated string
        const tokenString = currentTokens.join(', ');
        
        // Set the value in the decode input
        inputTokens.value = tokenString;
        
        // Visual feedback
        copyToDecodeBtn.classList.add('copy-success');
        copyToDecodeBtn.innerHTML = '<i class="bi bi-check-circle"></i> Copied!';
        
        // With side-by-side layout, just focus on decode button
        decodeBtn.focus();
        
        // If on mobile, need to scroll to the decode section
        if (window.innerWidth <= 768) {
            const decodeSection = document.querySelector('.tokenizer-section:last-child');
            if (decodeSection) {
                decodeSection.scrollIntoView({ behavior: 'smooth' });
            }
        }
        
        // Reset button after animation
        setTimeout(() => {
            copyToDecodeBtn.classList.remove('copy-success');
            copyToDecodeBtn.innerHTML = '<i class="bi bi-arrow-right-circle"></i> Copy to Decode';
        }, 2000);
    });

    // Handle decoding
    decodeBtn.addEventListener('click', async () => {
        const tokenText = inputTokens.value.trim();
        
        if (!tokenText) {
            showError(inputTokens, 'Please enter token IDs to decode');
            return;
        }
        
        try {
            // Parse the input as a comma-separated list of integers
            let tokens;
            try {
                tokens = tokenText.split(',').map(t => parseInt(t.trim(), 10));
                
                // Validate that all values are valid numbers
                if (tokens.some(isNaN)) {
                    throw new Error('Invalid token format');
                }
            } catch (e) {
                showError(inputTokens, 'Please enter valid comma-separated integers');
                return;
            }
            
            decodeBtn.disabled = true;
            decodeBtn.textContent = 'Decoding...';
            decodedText.innerHTML = '<div class="loading">Processing...</div>';
            
            const response = await fetch(`${API_BASE_URL}/decode`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ tokens })
            });
            
            if (!response.ok) {
                throw new Error(`Server error: ${response.status} ${response.statusText}`);
            }
            
            const text = await response.json();
            
            // Instead of replacing the element, update its content directly
            if (text && text.trim().length > 0) {
                // Use textContent instead of innerText to properly handle all Unicode characters
                decodedText.textContent = text;
                
                // Apply special highlighting if the text is short (< 50 chars)
                if (text.length < 50) {
                    decodedText.style.fontSize = '18px';
                } else {
                    decodedText.style.fontSize = '16px'; // Reset font size if needed
                }
                decodedText.style.fontStyle = 'normal';
                decodedText.style.opacity = '1';
            } else {
                decodedText.textContent = '(No output)';
                decodedText.style.fontStyle = 'italic';
                decodedText.style.opacity = '0.7';
            }
            
            // Add highlight animation by removing and re-adding the class
            decodedText.classList.remove('highlight-animation');
            void decodedText.offsetWidth; // Force reflow to restart animation
            decodedText.classList.add('highlight-animation');
            
            // Flash effect on the decode button to indicate success
            decodeBtn.classList.add('copy-success');
            setTimeout(() => {
                decodeBtn.classList.remove('copy-success');
            }, 1000);
            
        } catch (error) {
            console.error('Error during decoding:', error);
            decodedText.textContent = '';
            
            if (error.message.includes('Failed to fetch') || error.name === 'AbortError') {
                showError(inputTokens, 'Cannot connect to backend server. Please make sure the Go server is running on port 8080.');
            } else {
                showError(inputTokens, `Decoding failed: ${error.message}`);
            }
        } finally {
            decodeBtn.disabled = false;
            decodeBtn.textContent = 'Decode';
        }
    });

    // Function to show error message
    function showError(element, message) {
        const originalBorder = element.style.border;
        
        element.style.border = '1px solid #e74c3c';
        alert(message);
        
        setTimeout(() => {
            element.style.border = originalBorder;
        }, 2000);
    }

    // Add keyboard shortcuts: Ctrl+Enter to submit
    inputText.addEventListener('keydown', (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            encodeBtn.click();
        }
    });

    inputTokens.addEventListener('keydown', (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            decodeBtn.click();
        }
    });

    // Add click-to-copy functionality for token IDs
    tokenIds.addEventListener('click', (e) => {
        if (e.target.classList.contains('token-id')) {
            const tokenValue = e.target.textContent;
            navigator.clipboard.writeText(tokenValue).then(() => {
                // Visual feedback for copy
                const originalBg = e.target.style.backgroundColor;
                e.target.style.backgroundColor = 'var(--button-bg)';
                e.target.style.color = '#fff';
                
                setTimeout(() => {
                    e.target.style.backgroundColor = originalBg;
                    e.target.style.color = '';
                }, 300);
            });
        }
    });

    // Add click-to-copy functionality for tokens in the text display
    tokenDisplay.addEventListener('click', (e) => {
        if (e.target.classList.contains('token')) {
            const tokenId = e.target.dataset.tokenId;
            navigator.clipboard.writeText(tokenId).then(() => {
                // Visual feedback for copy
                const originalFilter = e.target.style.filter;
                e.target.style.filter = 'brightness(0.7)';
                
                // Show a temporary tooltip
                const tooltip = document.createElement('div');
                tooltip.className = 'token-tooltip';
                tooltip.textContent = `Copied ID: ${tokenId}`;
                
                // Position the tooltip near the token
                const rect = e.target.getBoundingClientRect();
                const containerRect = tokenDisplay.getBoundingClientRect();
                tooltip.style.left = `${rect.left - containerRect.left + tokenDisplay.scrollLeft}px`;
                tooltip.style.top = `${rect.top - containerRect.top - 25 + tokenDisplay.scrollTop}px`;
                
                tokenDisplay.appendChild(tooltip);
                
                setTimeout(() => {
                    e.target.style.filter = originalFilter;
                    tokenDisplay.removeChild(tooltip);
                }, 1000);
            });
        }
    });
}); 