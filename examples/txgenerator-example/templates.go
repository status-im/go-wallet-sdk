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
    <title>Transaction Generator</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 1000px;
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
            box-sizing: border-box;
        }
        select:focus, input:focus, textarea:focus {
            outline: none;
            border-color: #007bff;
        }
        textarea {
            resize: vertical;
            min-height: 100px;
            font-family: monospace;
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
            width: 100%;
        }
        button:hover {
            background: #0056b3;
        }
        button:disabled {
            background: #6c757d;
            cursor: not-allowed;
        }
        .radio-group {
            display: flex;
            gap: 20px;
            margin-top: 10px;
        }
        .radio-option {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .radio-option input[type="radio"] {
            width: auto;
        }
        .loading {
            display: none;
            text-align: center;
            margin: 20px 0;
        }
        .results {
            margin-top: 30px;
        }
        .result-box {
            background: #f8f9fa;
            border: 1px solid #e1e5e9;
            border-radius: 8px;
            padding: 20px;
            margin-top: 20px;
        }
        .result-box pre {
            margin: 0;
            white-space: pre-wrap;
            word-wrap: break-word;
            font-family: 'Courier New', monospace;
            font-size: 12px;
        }
        .error {
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
        .param-group {
            display: none;
        }
        .param-group.active {
            display: block;
        }
        .param-group label {
            font-size: 13px;
            margin-top: 10px;
        }
        .param-group input {
            margin-bottom: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔷 Transaction Generator</h1>
        
        <div class="info">
            <strong>Instructions:</strong> Select a transaction type, choose fee type (Legacy or EIP-1559), 
            fill in the required parameters, and click "Generate Transaction" to get the transaction in JSON format.
        </div>

        <form id="txForm">
            <div class="form-group">
                <label for="txType">Transaction Type:</label>
                <select id="txType" required>
                    <option value="">-- Select Transaction Type --</option>
                    <option value="transferETH">Transfer ETH</option>
                    <option value="transferERC20">Transfer ERC20</option>
                    <option value="approveERC20">Approve ERC20</option>
                    <option value="transferFromERC721">Transfer ERC721 (transferFrom)</option>
                    <option value="safeTransferFromERC721">Transfer ERC721 (safeTransferFrom)</option>
                    <option value="approveERC721">Approve ERC721</option>
                    <option value="setApprovalForAllERC721">Set Approval For All ERC721</option>
                    <option value="transferERC1155">Transfer ERC1155</option>
                    <option value="batchTransferERC1155">Batch Transfer ERC1155</option>
                    <option value="setApprovalForAllERC1155">Set Approval For All ERC1155</option>
                </select>
            </div>

            <div class="form-group">
                <label>Fee Type:</label>
                <div class="radio-group">
                    <div class="radio-option">
                        <input type="radio" id="feeLegacy" name="feeType" value="legacy" checked>
                        <label for="feeLegacy" style="margin: 0; font-weight: normal;">Legacy (GasPrice)</label>
                    </div>
                    <div class="radio-option">
                        <input type="radio" id="feeEIP1559" name="feeType" value="eip1559">
                        <label for="feeEIP1559" style="margin: 0; font-weight: normal;">EIP-1559 (MaxFeePerGas + MaxPriorityFeePerGas)</label>
                    </div>
                </div>
            </div>

            <div class="form-group">
                <label for="nonce">Nonce:</label>
                <input type="number" id="nonce" value="0" required>
            </div>

            <div class="form-group">
                <label for="gasLimit">Gas Limit:</label>
                <input type="number" id="gasLimit" value="21000" required>
            </div>

            <div class="form-group">
                <label for="chainID">Chain ID:</label>
                <input type="number" id="chainID" value="1" required>
            </div>

            <div class="form-group" id="legacyFeeGroup">
                <label for="gasPrice">Gas Price (wei):</label>
                <input type="text" id="gasPrice" value="20000000000" placeholder="20000000000">
            </div>

            <div class="form-group" id="eip1559FeeGroup" style="display: none;">
                <label for="maxFeePerGas">Max Fee Per Gas (wei):</label>
                <input type="text" id="maxFeePerGas" value="30000000000" placeholder="30000000000">
                <label for="maxPriorityFeePerGas" style="margin-top: 10px;">Max Priority Fee Per Gas (wei):</label>
                <input type="text" id="maxPriorityFeePerGas" value="2000000000" placeholder="2000000000">
            </div>

            <!-- Dynamic parameter fields -->
            <div id="paramsContainer"></div>

            <button type="submit" id="generateBtn">Generate Transaction</button>
        </form>

        <div class="loading" id="loading">
            <p>Generating transaction...</p>
        </div>

        <div class="results" id="results"></div>
    </div>

    <script>
        // Transaction type parameter definitions
        const txTypeParams = {
            transferETH: [
                { name: 'to', label: 'To Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'value', label: 'Value (wei)', type: 'text', placeholder: '1000000000000000000', required: true }
            ],
            transferERC20: [
                { name: 'tokenAddress', label: 'Token Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'to', label: 'To Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'amount', label: 'Amount (in token units)', type: 'text', placeholder: '1000000', required: true }
            ],
            approveERC20: [
                { name: 'tokenAddress', label: 'Token Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'spender', label: 'Spender Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'amount', label: 'Amount (in token units)', type: 'text', placeholder: '1000000', required: true }
            ],
            transferFromERC721: [
                { name: 'tokenAddress', label: 'Token Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'from', label: 'From Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'to', label: 'To Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'tokenID', label: 'Token ID', type: 'text', placeholder: '1234', required: true }
            ],
            safeTransferFromERC721: [
                { name: 'tokenAddress', label: 'Token Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'from', label: 'From Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'to', label: 'To Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'tokenID', label: 'Token ID', type: 'text', placeholder: '1234', required: true }
            ],
            approveERC721: [
                { name: 'tokenAddress', label: 'Token Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'to', label: 'To Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'tokenID', label: 'Token ID', type: 'text', placeholder: '1234', required: true }
            ],
            setApprovalForAllERC721: [
                { name: 'tokenAddress', label: 'Token Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'operator', label: 'Operator Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'approved', label: 'Approved (true/false)', type: 'text', placeholder: 'true', required: true }
            ],
            transferERC1155: [
                { name: 'tokenAddress', label: 'Token Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'from', label: 'From Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'to', label: 'To Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'tokenID', label: 'Token ID', type: 'text', placeholder: '1234', required: true },
                { name: 'value', label: 'Value (amount)', type: 'text', placeholder: '1', required: true }
            ],
            batchTransferERC1155: [
                { name: 'tokenAddress', label: 'Token Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'from', label: 'From Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'to', label: 'To Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'tokenIDs', label: 'Token IDs (comma-separated)', type: 'text', placeholder: '1,2,3', required: true },
                { name: 'values', label: 'Values (comma-separated)', type: 'text', placeholder: '1,2,3', required: true }
            ],
            setApprovalForAllERC1155: [
                { name: 'tokenAddress', label: 'Token Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'operator', label: 'Operator Address', type: 'text', placeholder: '0x...', required: true },
                { name: 'approved', label: 'Approved (true/false)', type: 'text', placeholder: 'true', required: true }
            ]
        };

        // Handle transaction type change
        document.getElementById('txType').addEventListener('change', function() {
            updateParamsFields(this.value);
        });

        // Handle fee type change
        document.querySelectorAll('input[name="feeType"]').forEach(radio => {
            radio.addEventListener('change', function() {
                if (this.value === 'legacy') {
                    document.getElementById('legacyFeeGroup').style.display = 'block';
                    document.getElementById('eip1559FeeGroup').style.display = 'none';
                    document.getElementById('gasPrice').required = true;
                    document.getElementById('maxFeePerGas').required = false;
                    document.getElementById('maxPriorityFeePerGas').required = false;
                } else {
                    document.getElementById('legacyFeeGroup').style.display = 'none';
                    document.getElementById('eip1559FeeGroup').style.display = 'block';
                    document.getElementById('gasPrice').required = false;
                    document.getElementById('maxFeePerGas').required = true;
                    document.getElementById('maxPriorityFeePerGas').required = true;
                }
            });
        });

        // Update parameter fields based on transaction type
        function updateParamsFields(txType) {
            const container = document.getElementById('paramsContainer');
            container.innerHTML = '';

            if (!txType || !txTypeParams[txType]) {
                return;
            }

            const params = txTypeParams[txType];
            params.forEach(param => {
                const group = document.createElement('div');
                group.className = 'form-group';
                
                const label = document.createElement('label');
                label.setAttribute('for', param.name);
                label.textContent = param.label + (param.required ? ' *' : '');
                
                const input = document.createElement('input');
                input.type = param.type;
                input.id = param.name;
                input.name = param.name;
                input.placeholder = param.placeholder;
                if (param.required) {
                    input.required = true;
                }
                
                group.appendChild(label);
                group.appendChild(input);
                container.appendChild(group);
            });
        }

        // Handle form submission
        document.getElementById('txForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const generateBtn = document.getElementById('generateBtn');
            const loading = document.getElementById('loading');
            const results = document.getElementById('results');
            
            // Get form values
            const txType = document.getElementById('txType').value;
            const useEIP1559 = document.getElementById('feeEIP1559').checked;
            const nonce = document.getElementById('nonce').value;
            const gasLimit = document.getElementById('gasLimit').value;
            const chainID = document.getElementById('chainID').value;
            const gasPrice = document.getElementById('gasPrice').value;
            const maxFeePerGas = document.getElementById('maxFeePerGas').value;
            const maxPriorityFeePerGas = document.getElementById('maxPriorityFeePerGas').value;
            
            // Get dynamic parameters
            const params = {};
            const paramInputs = document.querySelectorAll('#paramsContainer input');
            paramInputs.forEach(input => {
                if (input.value.trim()) {
                    params[input.name] = input.value.trim();
                }
            });
            
            // Build request
            const request = {
                txType: txType,
                useEIP1559: useEIP1559,
                nonce: nonce,
                gasLimit: gasLimit,
                chainID: chainID,
                gasPrice: gasPrice,
                maxFeePerGas: maxFeePerGas,
                maxPriorityFeePerGas: maxPriorityFeePerGas,
                params: params
            };
            
            // Show loading
            generateBtn.disabled = true;
            loading.style.display = 'block';
            results.innerHTML = '';
            
            try {
                const response = await fetch('/generate', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(request)
                });
                
                const data = await response.json();
                
                if (data.error) {
                    results.innerHTML = '<div class="error"><strong>Error:</strong> ' + data.error + '</div>';
                } else {
                    results.innerHTML = '<div class="result-box"><h3>Generated Transaction:</h3><pre>' + JSON.stringify(data, null, 2) + '</pre></div>';
                }
                
            } catch (error) {
                results.innerHTML = '<div class="error"><strong>Error:</strong> ' + error.message + '</div>';
            } finally {
                generateBtn.disabled = false;
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
