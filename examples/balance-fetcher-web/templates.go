package main

import (
	"html/template"
	"net/http"
)

// homeTemplate contains the HTML template for the main page
const homeTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Native Balance Fetcher</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: #555;
        }
        select, input, textarea {
            width: 100%;
            padding: 12px;
            border: 2px solid #e1e5e9;
            border-radius: 8px;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        select:focus, input:focus, textarea:focus {
            outline: none;
            border-color: #007bff;
        }
        textarea {
            resize: vertical;
            min-height: 100px;
        }
        button {
            background: #007bff;
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            cursor: pointer;
            transition: background-color 0.3s;
        }
        button:hover {
            background: #0056b3;
        }
        button:disabled {
            background: #6c757d;
            cursor: not-allowed;
        }
        .loading {
            display: none;
            text-align: center;
            margin: 20px 0;
        }
        .results {
            margin-top: 30px;
        }
        .chain-results {
            margin-bottom: 30px;
            border: 1px solid #e1e5e9;
            border-radius: 8px;
            overflow: hidden;
        }
        .chain-header {
            background: #f8f9fa;
            padding: 15px;
            font-weight: 600;
            border-bottom: 1px solid #e1e5e9;
        }
        .balance-item {
            padding: 15px;
            border-bottom: 1px solid #e1e5e9;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .balance-item:last-child {
            border-bottom: none;
        }
        .address {
            font-family: monospace;
            color: #007bff;
            word-break: break-all;
        }
        .balance {
            text-align: right;
        }
        .balance-eth {
            font-size: 18px;
            font-weight: 600;
            color: #28a745;
        }
        .balance-wei {
            font-size: 12px;
            color: #6c757d;
        }
        .error {
            color: #dc3545;
            font-style: italic;
        }
        .error-message {
            background: #f8d7da;
            color: #721c24;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
        }
        .info {
            background: #d1ecf1;
            color: #0c5460;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
        }
        .chain-list {
            margin-bottom: 20px;
        }
        .chain-row {
            display: flex;
            gap: 10px;
            margin-bottom: 10px;
        }
        .chain-row input {
            flex: 1;
        }
        .remove-chain-btn {
            background: #dc3545;
            color: white;
            border: none;
            border-radius: 8px;
            padding: 0 12px;
            font-size: 18px;
            cursor: pointer;
            height: 40px;
            align-self: center;
        }
        .remove-chain-btn:hover {
            background: #a71d2a;
        }
        .add-chain-btn {
            background: #28a745;
            color: white;
            border: none;
            border-radius: 8px;
            padding: 8px 16px;
            font-size: 16px;
            cursor: pointer;
            margin-bottom: 10px;
        }
        .add-chain-btn:hover {
            background: #218838;
        }
        
        .token-section {
            margin-top: 10px;
            padding: 10px;
            background: #f8f9fa;
            border-radius: 8px;
        }
        
        .token-section label {
            font-size: 14px;
            margin-bottom: 8px;
            display: block;
        }
        
        .token-list {
            margin-bottom: 10px;
        }
        
        .token-search-row {
            margin-bottom: 10px;
        }
        
        .token-search-input {
            width: 100%;
            padding: 8px 12px;
            border: 1px solid #e1e5e9;
            border-radius: 6px;
            font-size: 14px;
        }
        
        .token-row {
            display: flex;
            gap: 10px;
            margin-bottom: 8px;
            align-items: center;
        }
        
        .token-row select,
        .token-row input {
            flex: 1;
            padding: 8px;
            border: 1px solid #e1e5e9;
            border-radius: 4px;
            font-size: 12px;
        }
        
        .remove-token-btn {
            background: #dc3545;
            color: white;
            border: none;
            border-radius: 4px;
            padding: 0 8px;
            font-size: 14px;
            cursor: pointer;
            height: 32px;
            align-self: center;
        }
        
        .remove-token-btn:hover {
            background: #a71d2a;
        }
        
        .add-token-btn {
            background: #17a2b8;
            color: white;
            border: none;
            border-radius: 4px;
            padding: 6px 12px;
            font-size: 12px;
            cursor: pointer;
        }
        
        .add-token-btn:hover {
            background: #138496;
        }
        
        .erc20-balances {
            margin-top: 15px;
            padding: 10px;
            background: #f8f9fa;
            border-radius: 8px;
        }
        
        .erc20-balances h4 {
            margin: 0 0 10px 0;
            font-size: 14px;
            color: #495057;
        }
        
        .token-balance {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px 0;
            border-bottom: 1px solid #e9ecef;
        }
        
        .token-balance:last-child {
            border-bottom: none;
        }
        
        .token-info {
            display: flex;
            flex-direction: column;
            gap: 2px;
        }
        
        .token-symbol {
            font-weight: 600;
            color: #007bff;
        }
        
        .token-address {
            font-size: 11px;
            color: #6c757d;
            font-family: monospace;
        }
        
        .token-balance-amount {
            text-align: right;
        }
        
        .token-balance-amount .balance-eth {
            font-size: 14px;
            font-weight: 600;
            color: #28a745;
        }
        
        .token-balance-amount .balance-wei {
            font-size: 10px;
            color: #6c757d;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌐 Balance Fetcher (Native + ERC20)</h1>
        
        <div class="info">
            <strong>Note:</strong> Enter one or more (ChainID, RPC URL) pairs below. You can use any EVM-compatible chain.<br>
            Example RPC URLs: Infura, Alchemy, public endpoints, or your own node.<br>
            <strong>ERC20 Support:</strong> Select tokens from Uniswap's default token list or enter custom contract addresses.
        </div>

        <form id="balanceForm">
            <div class="form-group">
                <label>Chains (ChainID + RPC URL):</label>
                <div id="chainList" class="chain-list"></div>
                <button type="button" class="add-chain-btn" id="addChainBtn">+ Add Chain</button>
            </div>

            <div class="form-group">
                <label for="addresses">Ethereum Addresses (one per line):</label>
                <textarea id="addresses" placeholder="0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6&#10;0x1234567890123456789012345678901234567890&#10;0xabcdefabcdefabcdefabcdefabcdefabcdefabcd" required></textarea>
            </div>

            <div class="form-group">
                <label for="blockNum">Block Number (leave empty for latest):</label>
                <input type="text" id="blockNum" placeholder="18000000">
            </div>

            <button type="submit" id="fetchBtn">Fetch Balances</button>
        </form>

        <div class="loading" id="loading">
            <p>Fetching balances...</p>
        </div>

        <div class="results" id="results"></div>
    </div>

    <script>
        // Dynamic chain list logic
        function createChainRow(chainId = '', rpcUrl = '') {
            const row = document.createElement('div');
            row.className = 'chain-row';
            row.innerHTML =
                '<input type="number" class="chain-id-input" placeholder="Chain ID" value="' + chainId + '" required style="max-width: 120px;" />' +
                '<input type="text" class="rpc-url-input" placeholder="RPC URL" value="' + rpcUrl + '" required />' +
                '<div class="token-section">' +
                '<label>ERC20 Tokens (optional):</label>' +
                '<div class="token-list"></div>' +
                '<button type="button" class="add-token-btn">+ Add Token</button>' +
                '</div>' +
                '<button type="button" class="remove-chain-btn" title="Remove">&times;</button>';
            
            // Set up token functionality
            const addTokenBtn = row.querySelector('.add-token-btn');
            const tokenList = row.querySelector('.token-list');
            
            // Initialize with a default token row
            const defaultTokenRow = createTokenRow();
            tokenList.appendChild(defaultTokenRow);
            
            addTokenBtn.onclick = function() {
                addTokenRow(tokenList);
            };
            
            row.querySelector('.remove-chain-btn').onclick = function() {
                row.remove();
            };
            
            // Load tokens when chain ID changes
            const chainIdInput = row.querySelector('.chain-id-input');
            chainIdInput.addEventListener('change', function() {
                console.log('Chain ID changed to:', this.value);
                loadTokensForChain(parseInt(this.value), tokenList);
            });
            
            // If chain ID is pre-populated, load tokens immediately
            if (chainIdInput.value) {
                console.log('Pre-populated chain ID found:', chainIdInput.value);
                setTimeout(() => {
                    loadTokensForChain(parseInt(chainIdInput.value), tokenList);
                }, 200);
            } else {
                console.log('No pre-populated chain ID');
            }
            
            return row;
        }
        
        function createTokenRow(tokenAddress = '', isCustom = false) {
            const row = document.createElement('div');
            row.className = 'token-row';
            
            if (isCustom) {
                row.innerHTML = 
                    '<input type="text" class="token-address-input" placeholder="Token Contract Address" value="' + tokenAddress + '" required />' +
                    '<button type="button" class="remove-token-btn" title="Remove Token">&times;</button>';
            } else {
                row.innerHTML = 
                    '<select class="token-select">' +
                    '<option value="">Select a token...</option>' +
                    '</select>' +
                    '<button type="button" class="remove-token-btn" title="Remove Token">&times;</button>';
            }
            
            row.querySelector('.remove-token-btn').onclick = function() {
                row.remove();
            };
            
            console.log('Created token row:', row.outerHTML);
            return row;
        }
        
        async function loadTokensForSelect(chainId, selectElement) {
            if (!chainId || !selectElement) {
                console.log('loadTokensForSelect called with invalid parameters');
                return;
            }
            console.log('Loading tokens for select element, chain ID:', chainId);
            
            try {
                const response = await fetch('/api/tokens/' + chainId);
                const data = await response.json();
                console.log('Loaded tokens for select:', data.tokens.length);
                
                // Update the select with new tokens
                selectElement.innerHTML = '<option value="">Select a token...</option>';
                
                // Add all tokens to the dropdown
                data.tokens.forEach(token => {
                    const option = document.createElement('option');
                    option.value = token.address;
                    option.textContent = token.symbol + ' - ' + token.name;
                    selectElement.appendChild(option);
                });
                
                // Add "Add Custom Token" option
                const customOption = document.createElement('option');
                customOption.value = 'custom';
                customOption.textContent = '+ Add Custom Token';
                selectElement.appendChild(customOption);
                
                console.log('Updated select with', data.tokens.length, 'tokens');
                
            } catch (error) {
                console.error('Failed to load tokens for select:', error);
                selectElement.innerHTML = '<option value="">Error loading tokens</option>';
            }
        }

        function addTokenRow(tokenList, tokenAddress = '', isCustom = false) {
            const tokenRow = createTokenRow(tokenAddress, isCustom);
            tokenList.appendChild(tokenRow);
            
            if (!isCustom) {
                // Add custom token option
                const customOption = document.createElement('option');
                customOption.value = 'custom';
                customOption.textContent = '+ Add Custom Token';
                tokenRow.querySelector('.token-select').appendChild(customOption);
                
                tokenRow.querySelector('.token-select').addEventListener('change', function() {
                    if (this.value === 'custom') {
                        // Replace with custom input
                        const customRow = createTokenRow('', true);
                        tokenRow.parentNode.replaceChild(customRow, tokenRow);
                    }
                });
                
                // Load tokens for this new token row if there's a chain ID
                const chainRow = tokenList.closest('.chain-row');
                if (chainRow) {
                    const chainIdInput = chainRow.querySelector('.chain-id-input');
                    if (chainIdInput && chainIdInput.value) {
                        console.log('Loading tokens for new token row, chain ID:', chainIdInput.value);
                        loadTokensForSelect(parseInt(chainIdInput.value), tokenRow.querySelector('.token-select'));
                    }
                }
            }
        }
        
        async function loadTokensForChain(chainId, tokenList) {
            if (!chainId) {
                console.log('loadTokensForChain called with no chainId');
                return;
            }
            console.log('Loading tokens for chain:', chainId);
            
            try {
                const response = await fetch('/api/tokens/' + chainId);
                const data = await response.json();
                console.log('Loaded tokens:', data.tokens.length);
                
                // Find the existing select element
                const select = tokenList.querySelector('.token-select');
                if (!select) {
                    console.error('No select element found in tokenList');
                    return;
                }
                
                // Update the select with new tokens
                select.innerHTML = '<option value="">Select a token...</option>';
                
                // Add all tokens to the dropdown
                data.tokens.forEach(token => {
                    const option = document.createElement('option');
                    option.value = token.address;
                    option.textContent = token.symbol + ' - ' + token.name;
                    select.appendChild(option);
                });
                
                // Add "Add Custom Token" option
                const customOption = document.createElement('option');
                customOption.value = 'custom';
                customOption.textContent = '+ Add Custom Token';
                select.appendChild(customOption);
                
                // Add change event listener if not already present
                if (!select.hasAttribute('data-listener-added')) {
                    select.setAttribute('data-listener-added', 'true');
                    select.addEventListener('change', function() {
                        if (this.value === 'custom') {
                            // Replace with custom input
                            const tokenRow = this.closest('.token-row');
                            const customRow = createTokenRow('', true);
                            tokenRow.parentNode.replaceChild(customRow, tokenRow);
                        }
                    });
                }
                
                console.log('Updated select with', data.tokens.length, 'tokens');
                
            } catch (error) {
                console.error('Failed to load tokens:', error);
                // Add fallback with common tokens
                const select = tokenList.querySelector('.token-select');
                if (select) {
                    select.innerHTML = '<option value="">Error loading tokens</option>';
                }
            }
        }

        function addChainRow(chainId = '', rpcUrl = '') {
            const chainList = document.getElementById('chainList');
            chainList.appendChild(createChainRow(chainId, rpcUrl));
        }

        document.getElementById('addChainBtn').onclick = function(e) {
            e.preventDefault();
            addChainRow();
        };

        // Add a default row for user convenience
        window.onload = function() {
            console.log('window.onload running');
            // Prepopulate with popular chains
            const defaultChains = [
                { chainId: 1, rpcUrl: 'https://ethereum-rpc.publicnode.com', name: 'Ethereum' },
                { chainId: 10, rpcUrl: 'https://optimism-rpc.publicnode.com', name: 'Optimism' },
                { chainId: 42161, rpcUrl: 'https://arbitrum-one-rpc.publicnode.com', name: 'Arbitrum' },
                { chainId: 137, rpcUrl: 'https://polygon-bor-rpc.publicnode.com', name: 'Polygon' }
            ];
            
            defaultChains.forEach(chain => {
                addChainRow(chain.chainId, chain.rpcUrl);
            });
            
            // No need to manually load tokens - createChainRow handles it for pre-populated chain IDs
        };

        document.getElementById('balanceForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const fetchBtn = document.getElementById('fetchBtn');
            const loading = document.getElementById('loading');
            const results = document.getElementById('results');
            
            // Get chain configs
            const chainRows = document.querySelectorAll('.chain-row');
            const chains = [];
            for (const row of chainRows) {
                const chainId = row.querySelector('.chain-id-input').value.trim();
                const rpcUrl = row.querySelector('.rpc-url-input').value.trim();
                if (!chainId || !rpcUrl) continue;
                
                // Collect token addresses for this chain
                const tokenAddresses = [];
                const tokenRows = row.querySelectorAll('.token-row');
                for (const tokenRow of tokenRows) {
                    const tokenSelect = tokenRow.querySelector('.token-select');
                    const tokenInput = tokenRow.querySelector('.token-address-input');
                    
                    let tokenAddress = '';
                    if (tokenSelect) {
                        tokenAddress = tokenSelect.value.trim();
                    } else if (tokenInput) {
                        tokenAddress = tokenInput.value.trim();
                    }
                    
                    if (tokenAddress && tokenAddress !== 'custom') {
                        tokenAddresses.push(tokenAddress);
                    }
                }
                
                chains.push({ 
                    chainId: parseInt(chainId), 
                    rpcUrl,
                    tokenAddresses: tokenAddresses
                });
            }
            if (chains.length === 0) {
                alert('Please add at least one chain with ChainID and RPC URL.');
                return;
            }

            // Get addresses
            const addresses = document.getElementById('addresses').value.split('\n').filter(addr => addr.trim());
            if (addresses.length === 0) {
                alert('Please enter at least one address');
                return;
            }
            const blockNum = document.getElementById('blockNum').value.trim();
            
            // Show loading
            fetchBtn.disabled = true;
            loading.style.display = 'block';
            results.innerHTML = '';
            
            try {
                const response = await fetch('/fetch', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        chains: chains,
                        addresses: addresses,
                        blockNum: blockNum
                    })
                });
                
                const data = await response.json();
                
                if (data.errors && data.errors.length > 0) {
                    results.innerHTML = '<div class="error-message"><strong>Errors:</strong><br>' + data.errors.join('<br>') + '</div>';
                }
                
                if (data.results) {
                    let html = '<h2>Results:</h2>';
                    
                    for (const [chainId, chainResults] of Object.entries(data.results)) {
                        html += '<div class="chain-results">';
                        html += '<div class="chain-header">Chain ID: ' + chainId + '</div>';
                        if (chainResults['__chain_error__'] && chainResults['__chain_error__'].error) {
                            html += '<div class="error-message">' + chainResults['__chain_error__'].error + '</div>';
                        } else {
                            for (const [address, accountBalances] of Object.entries(chainResults)) {
                                html += '<div class="balance-item">';
                                html += '<div class="address">' + address + '</div>';
                                
                                // Native balance
                                const native = accountBalances.nativeBalance;
                                if (native.error) {
                                    html += '<div class="error">Native Error: ' + native.error + '</div>';
                                } else {
                                    html += '<div class="balance">';
                                    html += '<div class="balance-eth">' + native.balance + ' ETH</div>';
                                    html += '<div class="balance-wei">' + native.wei + ' wei</div>';
                                    html += '</div>';
                                }
                                
                                // ERC20 balances
                                if (accountBalances.erc20Balances && Object.keys(accountBalances.erc20Balances).length > 0) {
                                    html += '<div class="erc20-balances">';
                                    html += '<h4>ERC20 Tokens:</h4>';
                                    for (const [tokenAddress, tokenBalance] of Object.entries(accountBalances.erc20Balances)) {
                                        html += '<div class="token-balance">';
                                        if (tokenBalance.error) {
                                            html += '<div class="error">' + (tokenBalance.tokenSymbol || tokenAddress) + ' Error: ' + tokenBalance.error + '</div>';
                                        } else {
                                            html += '<div class="token-info">';
                                            html += '<span class="token-symbol">' + (tokenBalance.tokenSymbol || 'UNKNOWN') + '</span>';
                                            html += '<span class="token-address">(' + tokenAddress + ')</span>';
                                            html += '</div>';
                                            html += '<div class="token-balance-amount">';
                                            html += '<div class="balance-eth">' + tokenBalance.balance + '</div>';
                                            html += '<div class="balance-wei">' + tokenBalance.wei + ' wei</div>';
                                            html += '</div>';
                                        }
                                        html += '</div>';
                                    }
                                    html += '</div>';
                                }
                                
                                html += '</div>';
                            }
                        }
                        html += '</div>';
                    }
                    
                    results.innerHTML = html;
                }
                
            } catch (error) {
                results.innerHTML = '<div class="error-message">Error: ' + error.message + '</div>';
            } finally {
                fetchBtn.disabled = false;
                loading.style.display = 'none';
            }
        });
    </script>
</body>
</html>`

// handleHome serves the main web interface
func handleHome(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("home").Parse(homeTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}
