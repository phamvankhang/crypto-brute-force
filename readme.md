# Ethereum Vanity Brute-Force Wallet Scanner

This project is a high-performance Ethereum wallet generator and brute-force scanner written in Go. It generates random wallets (either from raw private keys or BIP39 seed phrases), compares them against a known list of target addresses, and logs any matches found.

## 🔍 Features:
- ⚡ Fast wallet generation using Go’s lightweight goroutines

- 🔐 Supports both random private key generation and BIP39 mnemonic seeds

- 📄 Customizable input for address list (CSV file with top Ethereum holders)

- 🧠 Optimized memory usage using fixed-size byte comparisons

- 📦 Fully modular with Go Modules (go.mod, go.sum)

## Design Strategy: Optimizing the Problem Itself
Typical brute-force Ethereum wallet scanners rely on generating a private key and checking the corresponding address via HTTP API calls (e.g., to Etherscan or Infura). However, this approach is fundamentally limited by:

- API rate limits and access restrictions

- HTTP latency and concurrency overhead

- Potential costs when querying at scale

This project takes a different route: instead of querying each generated address online, it preloads a large static dataset (e.g., the top 30 million Ethereum addresses by balance) into memory using a Go map or fixed-size byte set. All checks are performed entirely offline, resulting in:

- Massively improved throughput (hundreds of thousands of wallets/sec)

- No dependency on third-party providers

- A restructured problem where lookup is O(1), not O(network-bound)

By redefining the problem from "check with an API" to "check against an in-memory set", the solution achieves dramatically better performance — shifting the bottleneck from the network to raw computation.

## Benmark

- Device: Macbook M4

- Number of Gorutines: 12

![Alt text](./benmark.png?raw=true "Title")

## ⚙️ Installation & Run
### 1. Clone the repository:
```bash
git clone https://github.com/phamvankhang/crypto-brute-force.git
cd eth-wallet-scanner
```
### 2. Install Go modules:
```bash
go mod tidy
```
### 3. Prepare input files:
Ensure you have a CSV file (e.g. top10m.csv) with the following format:

```bash
address,eth_balance
0xabc123...,1234.56
...
```


✅ You can generate this file using **Google BigQuery**, by querying the public dataset `bigquery-public-data.crypto_ethereum.balances` with:
```sql
SELECT address, eth_balance FROM `bigquery-public-data.crypto_ethereum.balances`
```
Then save the result to a BigQuery table and export it as a CSV to your Google Drive or local disk.

(Optional) If using seed phrase generation, provide `wordlist.txt` with your BIP39 dictionary.

### 4. Run the scanner:
```bash
go run main.go
```
To benchmark for 30 seconds (default): just run the above

To run indefinitely: set isBenmark := false in main.go

## 🛠 Use Case:
Research project or educational experiment to test the feasibility and performance of brute-force attacks on Ethereum wallets using publicly known addresses.

⚠️ Disclaimer: This project is for educational and research purposes only. Unauthorized access to wallets is illegal and unethical.

## 📂 Structure:
- main.go: Entry point of the scanner

- top20m.csv: Example address list (top holders)

- matches.txt: Output file containing matched wallets

