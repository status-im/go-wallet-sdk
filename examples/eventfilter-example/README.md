# EventFilter Example

This example demonstrates how to use the `eventfilter` package to detect and display ERC20, ERC721, and ERC1155 transfer events for a specific account within a block range.

## Features

- **Command-line interface** with flexible options
- **Multi-token support** for ERC20, ERC721, and ERC1155 transfers
- **Direction filtering** (send, receive, or both)
- **Comprehensive transfer details** extracted from the `Unpacked` field
- **Enhanced formatting** with shortened addresses and scientific notation for large numbers
- **Raw event metadata** including event signatures, data length, and log properties
- **Debug information** showing contract keys, event keys, and unpacked types
- **Error handling** and validation

## Usage

### Basic Usage

```bash
go run main.go -account 0x1234567890123456789012345678901234567890 -start 18000000 -end 18001000
```

### With Custom RPC

```bash
go run main.go -rpc https://mainnet.infura.io/v3/YOUR_KEY -account 0x1234567890123456789012345678901234567890 -start 18000000 -end 18001000
```

### Filter by Direction

```bash
# Only outgoing transfers
go run main.go -account 0x1234567890123456789012345678901234567890 -start 18000000 -end 18001000 -direction send

# Only incoming transfers
go run main.go -account 0x1234567890123456789012345678901234567890 -start 18000000 -end 18001000 -direction receive
```

## Command Line Options

| Option | Description | Required | Default |
|--------|-------------|----------|---------|
| `-account` | Account address to filter transfers for | Yes | - |
| `-start` | Start block number | Yes | - |
| `-end` | End block number | Yes | - |
| `-rpc` | Ethereum RPC URL | No | `https://mainnet.infura.io/v3/YOUR_KEY` |
| `-direction` | Transfer direction (send/receive/both) | No | `both` |
| `-help` | Show help message | No | - |

## Output Format

The example displays transfers grouped by token type with comprehensive details extracted from the `Unpacked` field:

### ERC20 Transfers
```
=== ERC20 Transfers (5) ===
1. ERC20 Transfer
   Block: 18000001
   Transaction: 0x1234567890abcdef...
   From: 0x1234567890123456789012345678901234567890 (0x1234...7890)
   To: 0x5678901234567890123456789012345678901234 (0x5678...1234)
   Amount: 1000000000000000000
   Token Contract: 0xA0b86a33E6441c8C06Cdd238c2df0F7A8a8c8c8c (0xA0b8...c8c8)
   Log Index: 0
   Topics: 3
   Event Signature: 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
   Data Length: 32 bytes
   Removed: false
```

### ERC721 Transfers
```
=== ERC721 Transfers (2) ===
1. ERC721 Transfer (NFT)
   Block: 18000002
   Transaction: 0x1234567890abcdef...
   From: 0x1234567890123456789012345678901234567890 (0x1234...7890)
   To: 0x5678901234567890123456789012345678901234 (0x5678...1234)
   Token ID: 123
   NFT Contract: 0xB0b86a33E6441c8C06Cdd238c2df0F7A8a8c8c8c (0xB0b8...c8c8)
   Log Index: 1
   Topics: 4
   Event Signature: 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
   Data Length: 0 bytes
   Removed: false
```

### ERC1155 Transfers
```
=== ERC1155 Transfers (3) ===
1. ERC1155 Transfer Single
   Block: 18000003
   Transaction: 0x1234567890abcdef...
   Operator: 0x1234567890123456789012345678901234567890 (0x1234...7890)
   From: 0x1234567890123456789012345678901234567890 (0x1234...7890)
   To: 0x5678901234567890123456789012345678901234 (0x5678...1234)
   Token ID: 456
   Amount: 5
   Contract: 0xC0b86a33E6441c8C06Cdd238c2df0F7A8a8c8c8c (0xC0b8...c8c8)
   Log Index: 2
   Topics: 4
   Event Signature: 0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62
   Data Length: 64 bytes
   Removed: false

2. ERC1155 Transfer Batch
   Block: 18000004
   Transaction: 0x1234567890abcdef...
   Operator: 0x1234567890123456789012345678901234567890 (0x1234...7890)
   From: 0x1234567890123456789012345678901234567890 (0x1234...7890)
   To: 0x5678901234567890123456789012345678901234 (0x5678...1234)
   Contract: 0xD0b86a33E6441c8C06Cdd238c2df0F7A8a8c8c8c (0xD0b8...c8c8)
   Log Index: 3
   Topics: 4
   Batch Items (2):
     - Token ID: 789, Amount: 10
     - Token ID: 101, Amount: 2
   Event Signature: 0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb
   Data Length: 128 bytes
   Removed: false
```

### Raw Event Data (Debugging)
```
=== Raw Event Data (First 3 events) ===
Event 1:
  Contract Key: erc20
  Event Key: erc20transfer
  ABI Event Name: Transfer
  Unpacked Type: erc20.Erc20Transfer

Event 2:
  Contract Key: erc721
  Event Key: erc721transfer
  ABI Event Name: Transfer
  Unpacked Type: erc721.Erc721Transfer
```

## Setup

1. **Install dependencies**:
   ```bash
   go mod tidy
   ```

2. **Configure RPC endpoint**:
   - Replace `YOUR_KEY` in the default RPC URL with your Infura/Alchemy API key
   - Or use the `-rpc` flag to specify a custom endpoint

3. **Run the example**:
   ```bash
   go run main.go -account <ADDRESS> -start <START_BLOCK> -end <END_BLOCK>
   ```

## Example Output

```
Connecting to Ethereum RPC: https://mainnet.infura.io/v3/YOUR_KEY
Latest block: 18500000
Scanning blocks 18000000 to 18001000 for account 0x1234567890123456789012345678901234567890
Direction: both

Filtering transfer events...

Found 15 transfer events:

=== ERC20 Transfers (8) ===
1. Block 18000001 | 0x1234... -> 0x5678... | Amount: 1000000000000000000 | Token: 0xA0b86a33E6441c8C06Cdd238c2df0F7A8a8c8c8c
...

=== ERC721 Transfers (3) ===
1. Block 18000005 | 0x1234... -> 0x5678... | Token ID: 123 | Contract: 0xB0b86a33E6441c8C06Cdd238c2df0F7A8a8c8c8c
...

=== ERC1155 Transfers (4) ===
1. Block 18000008 | 0x1234... -> 0x5678... | Token ID: 456 | Amount: 5 | Contract: 0xC0b86a33E6441c8C06Cdd238c2df0F7A8a8c8c8c
...

Summary:
- ERC20 transfers: 8
- ERC721 transfers: 3
- ERC1155 transfers: 4
- Total transfers: 15
```

## Error Handling

The example includes comprehensive error handling for:
- Invalid account addresses
- Invalid block ranges
- RPC connection failures
- Network timeouts
- Invalid command line arguments

## Performance Notes

- The example uses the optimized `eventfilter.FilterTransfers` function
- Query optimization minimizes API calls to the Ethereum node
- Large block ranges may take time to process depending on network activity
