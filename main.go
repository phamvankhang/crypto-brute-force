package main

import (
	"bufio"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
	"encoding/csv"
    // "encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/common"
    bip39 "github.com/tyler-smith/go-bip39"
)

var (
	topSet    = make(map[string]struct{})
    topSetBytes = make(map[[20]byte]struct{})

	logChan   = make(chan string, 100)
	stopChan  = make(chan struct{})
	countAddr uint64 // dùng atomic để đếm số lượng ví
    wordlist []string
)

func loadWordlist(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("Failed to read wordlist")
	}
	words := strings.Split(string(data), "\n")
	filtered := []string{}
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w != "" {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

func generateMnemonic(numWords int) (string, error) {
	entropySize := (numWords * 11 * 32) / 33
	entropy, err := bip39.NewEntropy(entropySize)
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}

func derivePrivateKeyFromMnemonic(mnemonic string) (*ecdsa.PrivateKey, error) {
	seed := bip39.NewSeed(mnemonic, "")
	return crypto.ToECDSA(seed[:32])
}

func generateWalletsByMnemonic(numWords int) {
    select {
    case <-stopChan:
        return
    default:
        for {
            mnemonic, err := generateMnemonic(numWords)
            if err != nil {
                continue
            }
            privKey, err := derivePrivateKeyFromMnemonic(mnemonic)
            if err != nil {
                continue
            }
            atomic.AddUint64(&countAddr, 1)
            address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
            if _, exists := topSet[address]; exists {
                logChan <- fmt.Sprintf("🎯 MATCH FOUND: %s", privKey)
            }
        }
    }
}

func loadTopAddresses(filePath string) {
    file, err := os.Open(filePath)
    if err != nil {
        log.Fatalf("❌ Failed to open file: %v", err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        log.Fatalf("❌ Failed to read CSV: %v", err)
    }

    if len(records) < 1 || len(records[0]) < 1 || records[0][0] != "address" {
        log.Fatalf("❌ CSV file is not in expected format")
    }

    count := 0
    fmt.Println("🔍 First 5 addresses loaded:")
    for _, row := range records[1:] {
        if len(row) >= 1 {
            address := strings.ToLower(strings.TrimSpace(row[0]))
            if len(address) == 42 && strings.HasPrefix(address, "0x") {
                addrBytes := [20]byte{}
                copy(addrBytes[:], common.FromHex(address)[12:]) // remove '0x' and trim to 20 bytes
                topSetBytes[addrBytes] = struct{}{}

                topSet[address] = struct{}{}
                count++
                if count <= 5 {
                    fmt.Printf("  %d. %s\n", count, address)
                }
            }
        }
    }

    log.Printf("✅ Loaded %d addresses from CSV\n", count)
}

func bruteForceWorker(id int) {
	for {
		select {
		case <-stopChan:
			return
		default:
			privKey, err := crypto.GenerateKey()
			if err != nil {
				continue
			}
			pubKey := privKey.Public().(*ecdsa.PublicKey)
			address := strings.ToLower(crypto.PubkeyToAddress(*pubKey).Hex())

            if _, exists := topSet[address]; exists {
				logChan <- fmt.Sprintf("🎯 MATCH FOUND: %s", privKey)
			}

			atomic.AddUint64(&countAddr, 1)

			
		}
	}
}

func startLogger() {
	file, err := os.Create("matches.txt")
	if err != nil {
		log.Fatalf("❌ Cannot create log file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for addr := range logChan {
		fmt.Println(addr)
		writer.WriteString(addr + "\n")
	}
}

func generateRandomWallets(numCPU int) {
    rand.Seed(time.Now().UnixNano())

	for i := 0; i < numCPU; i++ {
		go bruteForceWorker(i)
	}
}

func generateBySeed(numCPU int) {
    wordlist = loadWordlist("wordlist.txt")
    numWords := 12
    for i := 0; i < numCPU; i++ {
        go generateWalletsByMnemonic(numWords)
    }
}

func main() {
    loadTopAddresses("top10m.csv")
	// Start logger goroutine
	go startLogger()
    numCPU := runtime.NumCPU()
	log.Printf("🚀 Starting brute-force on %d cores...\n", numCPU)
    
    useSeed := false
    isBenmark := true

    if useSeed {
        generateBySeed(numCPU)
    } else {
        generateRandomWallets(numCPU)
    }
    
    if !isBenmark {
        // keep running
        select {}
    } else {
        // Run for a fixed time
        start := time.Now()
        duration := 30 * time.Second
        time.Sleep(duration)
        close(stopChan)

        time.Sleep(1 * time.Second) // wait goroutine stop
        close(logChan)

        elapsed := time.Since(start)
        total := atomic.LoadUint64(&countAddr)
        speed := float64(total) / elapsed.Seconds()

        log.Printf("🧾 Checked %d wallets in %.2f seconds → %.0f wallets/sec\n", total, elapsed.Seconds(), speed)
    }

	
	log.Println("🛑 Brute-force completed.")
}
