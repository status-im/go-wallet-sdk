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
    </style>
</head>
<body>
    <div class="container">
        <h1>🌐 Native Balance Fetcher</h1>
        
        <div class="info">
            <strong>Note:</strong> Enter one or more (ChainID, RPC URL) pairs below. You can use any EVM-compatible chain.<br>
            Example RPC URLs: Infura, Alchemy, public endpoints, or your own node.
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
                '<input type="number" min="1" step="1" class="chain-id-input" placeholder="ChainID" value="' + chainId + '" required />' +
                '<input type="text" class="rpc-url-input" placeholder="RPC URL (https://...)" value="' + rpcUrl + '" required />' +
                '<button type="button" class="remove-chain-btn" title="Remove">&times;</button>';
            row.querySelector('.remove-chain-btn').onclick = function() {
                row.remove();
            };
            return row;
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
                chains.push({ chainId: parseInt(chainId), rpcUrl });
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
                            for (const [address, result] of Object.entries(chainResults)) {
                                html += '<div class="balance-item">';
                                html += '<div class="address">' + address + '</div>';
                                if (result.error) {
                                    html += '<div class="error">Error: ' + result.error + '</div>';
                                } else {
                                    html += '<div class="balance">';
                                    html += '<div class="balance-eth">' + result.balance + ' ETH</div>';
                                    html += '<div class="balance-wei">' + result.wei + ' wei</div>';
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
