package main

import (
	"html/template"
	"net/http"
)

// homeTemplate contains the HTML template for the main page with two-column layout
const homeTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Accounts Example - Extended vs Standard Keystore</title>
    <style>
        * {
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 0;
            background-color: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            text-align: center;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
        }
        .container {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 0;
            height: calc(100vh - 80px);
        }
        @media (max-width: 1024px) {
            .container {
                grid-template-columns: 1fr;
                height: auto;
            }
        }
        .column {
            background: white;
            padding: 20px;
            overflow-y: auto;
            border-right: 2px solid #e1e5e9;
        }
        .column:last-child {
            border-right: none;
        }
        .column-header {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            text-align: center;
            border: 2px solid #007bff;
        }
        .column-header h2 {
            margin: 0;
            color: #007bff;
            font-size: 18px;
        }
        .accounts-section {
            margin-bottom: 20px;
        }
        .accounts-list {
            max-height: 200px;
            overflow-y: auto;
            border: 1px solid #e1e5e9;
            border-radius: 6px;
            background: white;
            margin-bottom: 15px;
        }
        .account-item {
            padding: 10px;
            border-bottom: 1px solid #e1e5e9;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        .account-item:hover {
            background-color: #f8f9fa;
        }
        .account-item.selected {
            background-color: #e7f3ff;
            border-left: 4px solid #007bff;
        }
        .account-item:last-child {
            border-bottom: none;
        }
        .account-address {
            font-family: monospace;
            font-size: 12px;
            color: #007bff;
            word-break: break-all;
        }
        .account-status {
            font-size: 11px;
            color: #6c757d;
            margin-top: 4px;
        }
        .status-unlocked {
            color: #28a745;
        }
        .status-locked {
            color: #dc3545;
        }
        .section {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 15px;
        }
        .section h3 {
            margin-top: 0;
            color: #495057;
            font-size: 16px;
            border-bottom: 2px solid #007bff;
            padding-bottom: 8px;
        }
        .form-group {
            margin-bottom: 12px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: 600;
            color: #555;
            font-size: 13px;
        }
        input, textarea, select, button {
            width: 100%;
            padding: 8px;
            border: 2px solid #e1e5e9;
            border-radius: 6px;
            font-size: 13px;
            font-family: inherit;
        }
        input:focus, textarea:focus, select:focus {
            outline: none;
            border-color: #007bff;
        }
        textarea {
            resize: vertical;
            min-height: 60px;
        }
        button {
            background: #007bff;
            color: white;
            border: none;
            cursor: pointer;
            font-weight: 600;
            transition: background-color 0.3s;
            margin-top: 8px;
        }
        button:hover {
            background: #0056b3;
        }
        button:disabled {
            background: #6c757d;
            cursor: not-allowed;
        }
        .btn-secondary {
            background: #6c757d;
        }
        .btn-secondary:hover {
            background: #5a6268;
        }
        .btn-success {
            background: #28a745;
        }
        .btn-success:hover {
            background: #218838;
        }
        .btn-danger {
            background: #dc3545;
        }
        .btn-danger:hover {
            background: #c82333;
        }
        .result {
            margin-top: 10px;
            padding: 10px;
            border-radius: 6px;
            background: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
            font-size: 12px;
            word-break: break-all;
        }
        .error {
            background: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
        }
        .loading {
            display: none;
            text-align: center;
            padding: 8px;
            color: #007bff;
            font-size: 12px;
        }
        .mnemonic-display {
            background: #fff3cd;
            border: 1px solid #ffc107;
            padding: 12px;
            border-radius: 6px;
            font-family: monospace;
            font-size: 12px;
            line-height: 1.6;
            word-break: break-all;
            margin-top: 8px;
        }
        .inline-group {
            display: flex;
            gap: 8px;
            align-items: center;
        }
        .inline-group input {
            flex: 1;
        }
        .checkbox-group {
            display: flex;
            align-items: center;
            gap: 6px;
        }
        .checkbox-group input[type="checkbox"] {
            width: auto;
        }
        .account-info {
            background: #e7f3ff;
            padding: 12px;
            border-radius: 6px;
            margin-top: 15px;
            font-size: 12px;
        }
        .account-info h4 {
            margin-top: 0;
            color: #007bff;
            font-size: 14px;
        }
        .info-item {
            margin-bottom: 8px;
        }
        .info-label {
            font-weight: 600;
            color: #495057;
            font-size: 12px;
        }
        .info-value {
            font-family: monospace;
            font-size: 12px;
            color: #333;
            word-break: break-all;
        }
        .top-section {
            background: white;
            padding: 20px;
            margin: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            border: 2px solid #28a745;
        }
        .top-section h2 {
            margin-top: 0;
            color: #28a745;
            font-size: 18px;
            border-bottom: 2px solid #28a745;
            padding-bottom: 8px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔐 Accounts Example - Extended Keystore vs Standard Keystore</h1>
    </div>
    
    <!-- Mnemonic Generation Section -->
    <div class="top-section">
        <h2>🌱 Generate Random Seed Phrase</h2>
        <div class="form-group">
            <label for="extMnemonicLength">Word Count:</label>
            <select id="extMnemonicLength" style="max-width: 200px;">
                <option value="12">12 words</option>
                <option value="15">15 words</option>
                <option value="18">18 words</option>
                <option value="21">21 words</option>
                <option value="24">24 words</option>
            </select>
        </div>
        <button onclick="generateExtMnemonic()" style="max-width: 200px;">Generate Mnemonic</button>
        <div id="extMnemonicResult"></div>
    </div>
    
    <div class="container">
        <!-- Left Column: Extended Keystore -->
        <div class="column">
            <div class="column-header">
                <h2>Extended Keystore (extkeystore)</h2>
            </div>
            
            <!-- Accounts List -->
            <div class="accounts-section">
                <h3>Accounts</h3>
                <button onclick="refreshExtAccounts()" class="btn-secondary" style="width: auto; margin-bottom: 10px;">Refresh</button>
                <div id="extAccountsList" class="accounts-list"></div>
                <div id="extAccountInfo" class="account-info" style="display: none;">
                    <h4>Account Information</h4>
                    <div id="extAccountInfoContent"></div>
                </div>
            </div>

            <!-- Create Account from Mnemonic -->
            <div class="section">
                <h3>Create Account from Seed Phrase</h3>
                <div class="form-group">
                    <label for="extMnemonicInput">Mnemonic Phrase:</label>
                    <textarea id="extMnemonicInput" placeholder="Enter your mnemonic phrase"></textarea>
                </div>
                <div class="form-group">
                    <label for="extCreatePassphrase">Passphrase (optional):</label>
                    <input type="password" id="extCreatePassphrase" placeholder="Leave empty for no encryption">
                </div>
                <button onclick="createExtAccount()">Create Account</button>
                <div id="extCreateAccountResult"></div>
            </div>

            <!-- Derive Child Account -->
            <div class="section">
                <h3>Derive Child Account</h3>
                <div class="form-group">
                    <label for="extDeriveAddress">Parent Address:</label>
                    <select id="extDeriveAddress">
                        <option value="">Select an account...</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="extDerivePath">Derivation Path:</label>
                    <input type="text" id="extDerivePath" placeholder="m/44'/60'/0'/0/0" value="m/44'/60'/0'/0/0">
                </div>
                <div class="form-group">
                    <label for="extDerivePassphrase">Parent Passphrase:</label>
                    <input type="password" id="extDerivePassphrase" placeholder="Required to decrypt parent account">
                </div>
                <div class="form-group checkbox-group">
                    <input type="checkbox" id="extDerivePin" onchange="toggleDeriveNewPassphrase()">
                    <label for="extDerivePin" style="margin: 0;">Pin derived account (save to keystore)</label>
                </div>
                <div class="form-group" id="extDeriveNewPassphraseGroup" style="display: none;">
                    <label for="extDeriveNewPassphrase">Derived Account Passphrase (optional):</label>
                    <input type="password" id="extDeriveNewPassphrase" placeholder="Leave empty to use parent passphrase">
                </div>
                <button onclick="deriveExtAccount()">Derive Account</button>
                <div id="extDeriveResult"></div>
            </div>

            <!-- Import/Export -->
            <div class="section">
                <h3>Import/Export JSON Keys</h3>
                <div class="form-group">
                    <label for="extImportKeyJson">JSON Key (for import):</label>
                    <textarea id="extImportKeyJson" placeholder="Paste JSON key here"></textarea>
                </div>
                <div class="form-group">
                    <label for="extImportPassphrase">Current Passphrase:</label>
                    <input type="password" id="extImportPassphrase" placeholder="Required">
                </div>
                <div class="form-group">
                    <label for="extImportNewPassphrase">New Passphrase (optional):</label>
                    <input type="password" id="extImportNewPassphrase" placeholder="Leave empty to keep current">
                </div>
                <div class="form-group">
                    <label for="extExportAddress">Address (for export):</label>
                    <select id="extExportAddress">
                        <option value="">Select an account...</option>
                    </select>
                </div>
                <div class="inline-group">
                    <button onclick="importExtKey()" class="btn-success">Import Key</button>
                    <button onclick="exportExtKey()" class="btn-secondary">Export Key</button>
                </div>
                <div id="extImportExportResult"></div>
            </div>

            <!-- Export Private Key (Standard Keystore) -->
            <div class="section">
                <h3>Export Private Key (Standard Keystore)</h3>
                <div class="form-group">
                    <label for="extExportPrivAddress">Address:</label>
                    <select id="extExportPrivAddress">
                        <option value="">Select an account...</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="extExportPrivPassphrase">Current Passphrase:</label>
                    <input type="password" id="extExportPrivPassphrase" placeholder="Required">
                </div>
                <div class="form-group">
                    <label for="extExportPrivNewPassphrase">New Passphrase:</label>
                    <input type="password" id="extExportPrivNewPassphrase" placeholder="Required">
                </div>
                <button onclick="exportExtPriv()" class="btn-secondary">Export Private Key</button>
                <div id="extExportPrivResult"></div>
            </div>

            <!-- Sign Message -->
            <div class="section">
                <h3>Sign Message</h3>
                <div class="form-group">
                    <label for="extSignAddress">Address:</label>
                    <select id="extSignAddress">
                        <option value="">Select an account...</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="extSignMessage">Message:</label>
                    <textarea id="extSignMessage" placeholder="Enter message to sign"></textarea>
                </div>
                <div class="form-group">
                    <label for="extSignPassphrase">Passphrase (if locked):</label>
                    <input type="password" id="extSignPassphrase" placeholder="Leave empty if unlocked">
                </div>
                <button onclick="signExtMessage()">Sign Message</button>
                <div id="extSignResult"></div>
            </div>

            <!-- Unlock/Lock -->
            <div class="section">
                <h3>Unlock/Lock Account</h3>
                <div class="form-group">
                    <label for="extUnlockAddress">Address:</label>
                    <select id="extUnlockAddress">
                        <option value="">Select an account...</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="extUnlockPassphrase">Passphrase:</label>
                    <input type="password" id="extUnlockPassphrase" placeholder="Required">
                </div>
                <div class="form-group">
                    <label for="extUnlockTimeout">Timeout (seconds, 0 for indefinite):</label>
                    <input type="number" id="extUnlockTimeout" value="300" min="0">
                </div>
                <div class="inline-group">
                    <button onclick="unlockExtAccount()" class="btn-success">Unlock</button>
                    <button onclick="lockExtAccount()" class="btn-danger">Lock</button>
                </div>
                <div id="extUnlockResult"></div>
            </div>
        </div>

        <!-- Right Column: Standard Keystore -->
        <div class="column">
            <div class="column-header">
                <h2>Standard Keystore (keystore.KeyStore)</h2>
            </div>
            
            <!-- Accounts List -->
            <div class="accounts-section">
                <h3>Accounts</h3>
                <button onclick="refreshStdAccounts()" class="btn-secondary" style="width: auto; margin-bottom: 10px;">Refresh</button>
                <div id="stdAccountsList" class="accounts-list"></div>
                <div id="stdAccountInfo" class="account-info" style="display: none;">
                    <h4>Account Information</h4>
                    <div id="stdAccountInfoContent"></div>
                </div>
            </div>

            <!-- Create Account -->
            <div class="section">
                <h3>Create New Account</h3>
                <div class="form-group">
                    <label for="stdCreatePassphrase">Passphrase:</label>
                    <input type="password" id="stdCreatePassphrase" placeholder="Required">
                </div>
                <button onclick="createStdAccount()">Create Account</button>
                <div id="stdCreateAccountResult"></div>
            </div>

            <!-- Import/Export -->
            <div class="section">
                <h3>Import/Export JSON Keys</h3>
                <div class="form-group">
                    <label for="stdImportKeyJson">JSON Key (for import):</label>
                    <textarea id="stdImportKeyJson" placeholder="Paste JSON key here"></textarea>
                </div>
                <div class="form-group">
                    <label for="stdImportPassphrase">Current Passphrase:</label>
                    <input type="password" id="stdImportPassphrase" placeholder="Required">
                </div>
                <div class="form-group">
                    <label for="stdImportNewPassphrase">New Passphrase (optional):</label>
                    <input type="password" id="stdImportNewPassphrase" placeholder="Leave empty to keep current">
                </div>
                <div class="form-group">
                    <label for="stdExportAddress">Address (for export):</label>
                    <select id="stdExportAddress">
                        <option value="">Select an account...</option>
                    </select>
                </div>
                <div class="inline-group">
                    <button onclick="importStdKey()" class="btn-success">Import Key</button>
                    <button onclick="exportStdKey()" class="btn-secondary">Export Key</button>
                </div>
                <div id="stdImportExportResult"></div>
            </div>

            <!-- Sign Message -->
            <div class="section">
                <h3>Sign Message</h3>
                <div class="form-group">
                    <label for="stdSignAddress">Address:</label>
                    <select id="stdSignAddress">
                        <option value="">Select an account...</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="stdSignMessage">Message:</label>
                    <textarea id="stdSignMessage" placeholder="Enter message to sign"></textarea>
                </div>
                <div class="form-group">
                    <label for="stdSignPassphrase">Passphrase (if locked):</label>
                    <input type="password" id="stdSignPassphrase" placeholder="Leave empty if unlocked">
                </div>
                <button onclick="signStdMessage()">Sign Message</button>
                <div id="stdSignResult"></div>
            </div>

            <!-- Unlock/Lock -->
            <div class="section">
                <h3>Unlock/Lock Account</h3>
                <div class="form-group">
                    <label for="stdUnlockAddress">Address:</label>
                    <select id="stdUnlockAddress">
                        <option value="">Select an account...</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="stdUnlockPassphrase">Passphrase:</label>
                    <input type="password" id="stdUnlockPassphrase" placeholder="Required">
                </div>
                <div class="form-group">
                    <label for="stdUnlockTimeout">Timeout (seconds, 0 for indefinite):</label>
                    <input type="number" id="stdUnlockTimeout" value="300" min="0">
                </div>
                <div class="inline-group">
                    <button onclick="unlockStdAccount()" class="btn-success">Unlock</button>
                    <button onclick="lockStdAccount()" class="btn-danger">Lock</button>
                </div>
                <div id="stdUnlockResult"></div>
            </div>
        </div>
    </div>

    <script>
        let selectedExtAccount = null;
        let selectedStdAccount = null;

        // Helper function to extract error message from response
        async function extractErrorMessage(response) {
            try {
                // Read response as text first
                const text = await response.text();
                
                // Try to parse as JSON
                try {
                    const data = JSON.parse(text);
                    if (data.error) {
                        return data.error;
                    }
                    // If no error field, return the full JSON stringified for debugging
                    return JSON.stringify(data, null, 2);
                } catch (e) {
                    // Not JSON, return as text
                    return text || response.statusText || 'Unknown error';
                }
            } catch (error) {
                return response.statusText || error.message || 'Failed to parse error response';
            }
        }

        // Helper function to escape HTML
        function escapeHtml(text) {
            const map = {
                '&': '&amp;',
                '<': '&lt;',
                '>': '&gt;',
                '"': '&quot;',
                "'": '&#039;'
            };
            return text.replace(/[&<>"']/g, function(m) { return map[m]; });
        }

        // Extended Keystore Functions
        async function generateExtMnemonic() {
            const length = parseInt(document.getElementById('extMnemonicLength').value);
            const resultDiv = document.getElementById('extMnemonicResult');
            resultDiv.innerHTML = '<div class="loading">Generating...</div>';

            try {
                const response = await fetch('/api/ext/generate-mnemonic', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({length: length})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="mnemonic-display">' + data.mnemonic + '</div>';
                    document.getElementById('extMnemonicInput').value = data.mnemonic;
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function createExtAccount() {
            const mnemonic = document.getElementById('extMnemonicInput').value.trim();
            const passphrase = document.getElementById('extCreatePassphrase').value;
            const resultDiv = document.getElementById('extCreateAccountResult');
            resultDiv.innerHTML = '<div class="loading">Creating account...</div>';

            try {
                const response = await fetch('/api/ext/create-account', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({mnemonic: mnemonic, passphrase: passphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Account created: ' + data.address + '</div>';
                    refreshExtAccounts();
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        function toggleDeriveNewPassphrase() {
            const pinChecked = document.getElementById('extDerivePin').checked;
            const newPassphraseGroup = document.getElementById('extDeriveNewPassphraseGroup');
            if (pinChecked) {
                newPassphraseGroup.style.display = 'block';
            } else {
                newPassphraseGroup.style.display = 'none';
                document.getElementById('extDeriveNewPassphrase').value = '';
            }
        }

        async function deriveExtAccount() {
            const address = document.getElementById('extDeriveAddress').value;
            if (!address) {
                document.getElementById('extDeriveResult').innerHTML = '<div class="result error">Please select an account</div>';
                return;
            }
            const path = document.getElementById('extDerivePath').value.trim();
            const passphrase = document.getElementById('extDerivePassphrase').value;
            const newPassphrase = document.getElementById('extDeriveNewPassphrase').value;
            const pin = document.getElementById('extDerivePin').checked;
            const resultDiv = document.getElementById('extDeriveResult');
            resultDiv.innerHTML = '<div class="loading">Deriving account...</div>';

            try {
                const response = await fetch('/api/ext/derive-account', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        address: address,
                        path: path,
                        passphrase: passphrase,
                        newPassphrase: newPassphrase,
                        pin: pin
                    })
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Derived account: ' + data.address + (data.pinned ? ' (pinned)' : '') + '</div>';
                    if (data.pinned) {
                        refreshExtAccounts();
                    }
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function importExtKey() {
            const keyJson = document.getElementById('extImportKeyJson').value.trim();
            const passphrase = document.getElementById('extImportPassphrase').value;
            const newPassphrase = document.getElementById('extImportNewPassphrase').value;
            const resultDiv = document.getElementById('extImportExportResult');
            resultDiv.innerHTML = '<div class="loading">Importing key...</div>';

            try {
                const response = await fetch('/api/ext/import-key', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({keyJson: keyJson, passphrase: passphrase, newPassphrase: newPassphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Key imported: ' + data.address + '</div>';
                    refreshExtAccounts();
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function exportExtKey() {
            const address = document.getElementById('extExportAddress').value;
            if (!address) {
                document.getElementById('extImportExportResult').innerHTML = '<div class="result error">Please select an account</div>';
                return;
            }
            const passphrase = document.getElementById('extImportPassphrase').value;
            const newPassphrase = document.getElementById('extImportNewPassphrase').value || passphrase;
            const resultDiv = document.getElementById('extImportExportResult');
            resultDiv.innerHTML = '<div class="loading">Exporting key...</div>';

            try {
                const response = await fetch('/api/ext/export-key', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address, passphrase: passphrase, newPassphrase: newPassphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result"><strong>Exported Key:</strong><br><textarea style="width:100%;height:100px;font-family:monospace;font-size:11px;">' + data.keyJson + '</textarea></div>';
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function exportExtPriv() {
            const address = document.getElementById('extExportPrivAddress').value;
            if (!address) {
                document.getElementById('extExportPrivResult').innerHTML = '<div class="result error">Please select an account</div>';
                return;
            }
            const passphrase = document.getElementById('extExportPrivPassphrase').value;
            const newPassphrase = document.getElementById('extExportPrivNewPassphrase').value;
            const resultDiv = document.getElementById('extExportPrivResult');
            resultDiv.innerHTML = '<div class="loading">Exporting private key...</div>';

            try {
                const response = await fetch('/api/ext/export-priv', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address, passphrase: passphrase, newPassphrase: newPassphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result"><strong>Exported Private Key (Standard Keystore):</strong><br><textarea style="width:100%;height:100px;font-family:monospace;font-size:11px;">' + data.keyJson + '</textarea></div>';
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function signExtMessage() {
            const address = document.getElementById('extSignAddress').value;
            if (!address) {
                document.getElementById('extSignResult').innerHTML = '<div class="result error">Please select an account</div>';
                return;
            }
            const message = document.getElementById('extSignMessage').value.trim();
            const passphrase = document.getElementById('extSignPassphrase').value;
            const resultDiv = document.getElementById('extSignResult');
            resultDiv.innerHTML = '<div class="loading">Signing message...</div>';

            try {
                const response = await fetch('/api/ext/sign-message', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address, message: message, passphrase: passphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result"><strong>Signature:</strong><br>' + data.signature + '<br><br><strong>Message Hash:</strong><br>' + data.messageHash + '</div>';
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function unlockExtAccount() {
            const address = document.getElementById('extUnlockAddress').value;
            if (!address) {
                document.getElementById('extUnlockResult').innerHTML = '<div class="result error">Please select an account</div>';
                return;
            }
            const passphrase = document.getElementById('extUnlockPassphrase').value;
            const timeout = parseInt(document.getElementById('extUnlockTimeout').value) || 0;
            const resultDiv = document.getElementById('extUnlockResult');
            resultDiv.innerHTML = '<div class="loading">Unlocking account...</div>';

            try {
                const response = await fetch('/api/ext/unlock-account', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address, passphrase: passphrase, timeout: timeout})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Account unlocked' + (timeout > 0 ? ' for ' + timeout + ' seconds' : ' indefinitely') + '</div>';
                    refreshExtAccounts();
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function lockExtAccount() {
            const address = document.getElementById('extUnlockAddress').value;
            if (!address) {
                document.getElementById('extUnlockResult').innerHTML = '<div class="result error">Please select an account</div>';
                return;
            }
            const resultDiv = document.getElementById('extUnlockResult');
            resultDiv.innerHTML = '<div class="loading">Locking account...</div>';

            try {
                const response = await fetch('/api/ext/lock-account', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Account locked</div>';
                    refreshExtAccounts();
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        function populateExtAddressDropdowns(accounts) {
            const dropdownIds = ['extDeriveAddress', 'extExportAddress', 'extExportPrivAddress', 'extSignAddress', 'extUnlockAddress'];
            
            dropdownIds.forEach(dropdownId => {
                const dropdown = document.getElementById(dropdownId);
                if (!dropdown) return;
                
                const currentValue = dropdown.value;
                dropdown.innerHTML = '<option value="">Select an account...</option>';
                
                accounts.forEach(acc => {
                    const option = document.createElement('option');
                    option.value = acc.address;
                    option.textContent = acc.address;
                    dropdown.appendChild(option);
                });
                
                if (currentValue && accounts.some(acc => acc.address === currentValue)) {
                    dropdown.value = currentValue;
                }
            });
        }

        async function refreshExtAccounts() {
            const listDiv = document.getElementById('extAccountsList');
            listDiv.innerHTML = '<div class="loading">Loading accounts...</div>';

            try {
                const response = await fetch('/api/ext/accounts');
                if (response.ok) {
                    const data = await response.json();
                    populateExtAddressDropdowns(data.accounts);
                    
                    if (data.accounts.length === 0) {
                        listDiv.innerHTML = '<div style="padding:15px;text-align:center;color:#6c757d;">No accounts found</div>';
                    } else {
                        let html = '';
                        data.accounts.forEach(acc => {
                            const isSelected = selectedExtAccount === acc.address;
                            html += '<div class="account-item' + (isSelected ? ' selected' : '') + '" onclick="selectExtAccount(\'' + acc.address + '\')">';
                            html += '<div class="account-address">' + acc.address + '</div>';
                            html += '<div class="account-status">Click to view details</div>';
                            html += '</div>';
                        });
                        listDiv.innerHTML = html;
                    }
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    listDiv.innerHTML = '<div class="result error">Error loading accounts: ' + errorMsg + '</div>';
                }
            } catch (error) {
                listDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function selectExtAccount(address) {
            selectedExtAccount = address;
            const infoDiv = document.getElementById('extAccountInfo');
            const contentDiv = document.getElementById('extAccountInfoContent');
            contentDiv.innerHTML = '<div class="loading">Loading account info...</div>';
            infoDiv.style.display = 'block';

            try {
                const response = await fetch('/api/ext/account/' + address);
                if (response.ok) {
                    const data = await response.json();
                    let html = '';
                    html += '<div class="info-item"><span class="info-label">Address:</span><br><span class="info-value">' + data.address + '</span></div>';
                    html += '<div class="info-item"><span class="info-label">Status:</span><br><span class="info-value ' + (data.unlocked ? 'status-unlocked' : 'status-locked') + '">' + (data.unlocked ? 'Unlocked' : 'Locked') + '</span></div>';
                    if (data.filePath) {
                        html += '<div class="info-item"><span class="info-label">File Path:</span><br><span class="info-value" style="word-break: break-all;">' + escapeHtml(data.filePath) + '</span></div>';
                    }
                    if (data.fileContents) {
                        html += '<div class="info-item"><span class="info-label">Keystore File Contents:</span><br><textarea readonly style="width:100%;height:200px;font-family:monospace;font-size:11px;background:#f8f9fa;border:1px solid #dee2e6;padding:8px;resize:vertical;">' + escapeHtml(data.fileContents) + '</textarea></div>';
                    }
                    html += '<div class="info-item" style="margin-top: 15px; padding-top: 15px; border-top: 1px solid #dee2e6;">';
                    html += '<label for="extDeletePassphrase">Passphrase to delete:</label>';
                    html += '<input type="password" id="extDeletePassphrase" placeholder="Enter passphrase to confirm deletion" style="margin-bottom: 8px;">';
                    html += '<button onclick="deleteExtAccount(\'' + data.address + '\')" class="btn-danger" style="width: 100%;">Delete Account</button>';
                    html += '<div id="extDeleteResult" style="margin-top: 8px;"></div>';
                    html += '</div>';
                    contentDiv.innerHTML = html;

                    const addressSelects = ['extDeriveAddress', 'extExportAddress', 'extExportPrivAddress', 'extSignAddress', 'extUnlockAddress'];
                    addressSelects.forEach(selectId => {
                        const select = document.getElementById(selectId);
                        if (select) {
                            select.value = address;
                        }
                    });
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    contentDiv.innerHTML = '<div class="result error">Error loading account info: ' + errorMsg + '</div>';
                }
            } catch (error) {
                contentDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }

            refreshExtAccounts();
        }

        async function deleteExtAccount(address) {
            if (!address) {
                document.getElementById('extDeleteResult').innerHTML = '<div class="result error">No account selected</div>';
                return;
            }
            
            if (!confirm('Are you sure you want to delete this account? This action cannot be undone!')) {
                return;
            }
            
            const passphrase = document.getElementById('extDeletePassphrase').value;
            const resultDiv = document.getElementById('extDeleteResult');
            resultDiv.innerHTML = '<div class="loading">Deleting account...</div>';

            try {
                const response = await fetch('/api/ext/delete-account', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address, passphrase: passphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Account deleted successfully</div>';
                    selectedExtAccount = null;
                    document.getElementById('extAccountInfo').style.display = 'none';
                    refreshExtAccounts();
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        // Standard Keystore Functions
        async function createStdAccount() {
            const passphrase = document.getElementById('stdCreatePassphrase').value;
            const resultDiv = document.getElementById('stdCreateAccountResult');
            resultDiv.innerHTML = '<div class="loading">Creating account...</div>';

            try {
                const response = await fetch('/api/std/create-account', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({passphrase: passphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Account created: ' + data.address + '</div>';
                    refreshStdAccounts();
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function importStdKey() {
            const keyJson = document.getElementById('stdImportKeyJson').value.trim();
            const passphrase = document.getElementById('stdImportPassphrase').value;
            const newPassphrase = document.getElementById('stdImportNewPassphrase').value;
            const resultDiv = document.getElementById('stdImportExportResult');
            resultDiv.innerHTML = '<div class="loading">Importing key...</div>';

            try {
                const response = await fetch('/api/std/import-key', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({keyJson: keyJson, passphrase: passphrase, newPassphrase: newPassphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Key imported: ' + data.address + '</div>';
                    refreshStdAccounts();
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function exportStdKey() {
            const address = document.getElementById('stdExportAddress').value;
            if (!address) {
                document.getElementById('stdImportExportResult').innerHTML = '<div class="result error">Please select an account</div>';
                return;
            }
            const passphrase = document.getElementById('stdImportPassphrase').value;
            const newPassphrase = document.getElementById('stdImportNewPassphrase').value || passphrase;
            const resultDiv = document.getElementById('stdImportExportResult');
            resultDiv.innerHTML = '<div class="loading">Exporting key...</div>';

            try {
                const response = await fetch('/api/std/export-key', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address, passphrase: passphrase, newPassphrase: newPassphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result"><strong>Exported Key:</strong><br><textarea style="width:100%;height:100px;font-family:monospace;font-size:11px;">' + data.keyJson + '</textarea></div>';
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function signStdMessage() {
            const address = document.getElementById('stdSignAddress').value;
            if (!address) {
                document.getElementById('stdSignResult').innerHTML = '<div class="result error">Please select an account</div>';
                return;
            }
            const message = document.getElementById('stdSignMessage').value.trim();
            const passphrase = document.getElementById('stdSignPassphrase').value;
            const resultDiv = document.getElementById('stdSignResult');
            resultDiv.innerHTML = '<div class="loading">Signing message...</div>';

            try {
                const response = await fetch('/api/std/sign-message', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address, message: message, passphrase: passphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result"><strong>Signature:</strong><br>' + data.signature + '<br><br><strong>Message Hash:</strong><br>' + data.messageHash + '</div>';
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function unlockStdAccount() {
            const address = document.getElementById('stdUnlockAddress').value;
            if (!address) {
                document.getElementById('stdUnlockResult').innerHTML = '<div class="result error">Please select an account</div>';
                return;
            }
            const passphrase = document.getElementById('stdUnlockPassphrase').value;
            const timeout = parseInt(document.getElementById('stdUnlockTimeout').value) || 0;
            const resultDiv = document.getElementById('stdUnlockResult');
            resultDiv.innerHTML = '<div class="loading">Unlocking account...</div>';

            try {
                const response = await fetch('/api/std/unlock-account', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address, passphrase: passphrase, timeout: timeout})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Account unlocked' + (timeout > 0 ? ' for ' + timeout + ' seconds' : ' indefinitely') + '</div>';
                    refreshStdAccounts();
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function lockStdAccount() {
            const address = document.getElementById('stdUnlockAddress').value;
            if (!address) {
                document.getElementById('stdUnlockResult').innerHTML = '<div class="result error">Please select an account</div>';
                return;
            }
            const resultDiv = document.getElementById('stdUnlockResult');
            resultDiv.innerHTML = '<div class="loading">Locking account...</div>';

            try {
                const response = await fetch('/api/std/lock-account', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Account locked</div>';
                    refreshStdAccounts();
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        function populateStdAddressDropdowns(accounts) {
            const dropdownIds = ['stdExportAddress', 'stdSignAddress', 'stdUnlockAddress'];
            
            dropdownIds.forEach(dropdownId => {
                const dropdown = document.getElementById(dropdownId);
                if (!dropdown) return;
                
                const currentValue = dropdown.value;
                dropdown.innerHTML = '<option value="">Select an account...</option>';
                
                accounts.forEach(acc => {
                    const option = document.createElement('option');
                    option.value = acc.address;
                    option.textContent = acc.address;
                    dropdown.appendChild(option);
                });
                
                if (currentValue && accounts.some(acc => acc.address === currentValue)) {
                    dropdown.value = currentValue;
                }
            });
        }

        async function refreshStdAccounts() {
            const listDiv = document.getElementById('stdAccountsList');
            listDiv.innerHTML = '<div class="loading">Loading accounts...</div>';

            try {
                const response = await fetch('/api/std/accounts');
                if (response.ok) {
                    const data = await response.json();
                    populateStdAddressDropdowns(data.accounts);
                    
                    if (data.accounts.length === 0) {
                        listDiv.innerHTML = '<div style="padding:15px;text-align:center;color:#6c757d;">No accounts found</div>';
                    } else {
                        let html = '';
                        data.accounts.forEach(acc => {
                            const isSelected = selectedStdAccount === acc.address;
                            html += '<div class="account-item' + (isSelected ? ' selected' : '') + '" onclick="selectStdAccount(\'' + acc.address + '\')">';
                            html += '<div class="account-address">' + acc.address + '</div>';
                            html += '<div class="account-status">Click to view details</div>';
                            html += '</div>';
                        });
                        listDiv.innerHTML = html;
                    }
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    listDiv.innerHTML = '<div class="result error">Error loading accounts: ' + errorMsg + '</div>';
                }
            } catch (error) {
                listDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        async function selectStdAccount(address) {
            selectedStdAccount = address;
            const infoDiv = document.getElementById('stdAccountInfo');
            const contentDiv = document.getElementById('stdAccountInfoContent');
            contentDiv.innerHTML = '<div class="loading">Loading account info...</div>';
            infoDiv.style.display = 'block';

            try {
                const response = await fetch('/api/std/account/' + address);
                if (response.ok) {
                    const data = await response.json();
                    let html = '';
                    html += '<div class="info-item"><span class="info-label">Address:</span><br><span class="info-value">' + data.address + '</span></div>';
                    html += '<div class="info-item"><span class="info-label">Status:</span><br><span class="info-value ' + (data.unlocked ? 'status-unlocked' : 'status-locked') + '">' + (data.unlocked ? 'Unlocked' : 'Locked') + '</span></div>';
                    if (data.filePath) {
                        html += '<div class="info-item"><span class="info-label">File Path:</span><br><span class="info-value" style="word-break: break-all;">' + escapeHtml(data.filePath) + '</span></div>';
                    }
                    if (data.fileContents) {
                        html += '<div class="info-item"><span class="info-label">Keystore File Contents:</span><br><textarea readonly style="width:100%;height:200px;font-family:monospace;font-size:11px;background:#f8f9fa;border:1px solid #dee2e6;padding:8px;resize:vertical;">' + escapeHtml(data.fileContents) + '</textarea></div>';
                    }
                    html += '<div class="info-item" style="margin-top: 15px; padding-top: 15px; border-top: 1px solid #dee2e6;">';
                    html += '<label for="stdDeletePassphrase">Passphrase to delete:</label>';
                    html += '<input type="password" id="stdDeletePassphrase" placeholder="Enter passphrase to confirm deletion" style="margin-bottom: 8px;">';
                    html += '<button onclick="deleteStdAccount(\'' + data.address + '\')" class="btn-danger" style="width: 100%;">Delete Account</button>';
                    html += '<div id="stdDeleteResult" style="margin-top: 8px;"></div>';
                    html += '</div>';
                    contentDiv.innerHTML = html;

                    const addressSelects = ['stdExportAddress', 'stdSignAddress', 'stdUnlockAddress'];
                    addressSelects.forEach(selectId => {
                        const select = document.getElementById(selectId);
                        if (select) {
                            select.value = address;
                        }
                    });
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    contentDiv.innerHTML = '<div class="result error">Error loading account info: ' + errorMsg + '</div>';
                }
            } catch (error) {
                contentDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }

            refreshStdAccounts();
        }

        async function deleteStdAccount(address) {
            if (!address) {
                document.getElementById('stdDeleteResult').innerHTML = '<div class="result error">No account selected</div>';
                return;
            }
            
            if (!confirm('Are you sure you want to delete this account? This action cannot be undone!')) {
                return;
            }
            
            const passphrase = document.getElementById('stdDeletePassphrase').value;
            const resultDiv = document.getElementById('stdDeleteResult');
            resultDiv.innerHTML = '<div class="loading">Deleting account...</div>';

            try {
                const response = await fetch('/api/std/delete-account', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({address: address, passphrase: passphrase})
                });
                if (response.ok) {
                    const data = await response.json();
                    resultDiv.innerHTML = '<div class="result">Account deleted successfully</div>';
                    selectedStdAccount = null;
                    document.getElementById('stdAccountInfo').style.display = 'none';
                    refreshStdAccounts();
                } else {
                    const errorMsg = await extractErrorMessage(response);
                    resultDiv.innerHTML = '<div class="result error">Error: ' + errorMsg + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error">Error: ' + error.message + '</div>';
            }
        }

        // Load accounts on page load
        window.onload = function() {
            refreshExtAccounts();
            refreshStdAccounts();
        };
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
