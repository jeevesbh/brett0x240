package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
)

type EtherscanResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

type TokenDetailsResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Decimals string `json:"decimals"`
	} `json:"result"`
}

const etherscanAPIURL = "https://api.etherscan.io/api"

func fetchTotalSupply(contractAddress, apiKey string) (*big.Int, error) {
	url := fmt.Sprintf("%s?module=stats&action=tokensupply&contractaddress=%s&apikey=%s", etherscanAPIURL, contractAddress, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching total supply: %v", err)
	}
	defer resp.Body.Close()

	var etherscanResp EtherscanResponse
	if err := json.NewDecoder(resp.Body).Decode(&etherscanResp); err != nil {
		return nil, fmt.Errorf("error decoding total supply response: %v", err)
	}

	if etherscanResp.Status != "1" {
		return nil, fmt.Errorf("error from Etherscan: %s", etherscanResp.Message)
	}

	// Convert the result to a big integer
	totalSupply, ok := new(big.Int).SetString(etherscanResp.Result, 10)
	if !ok {
		return nil, fmt.Errorf("invalid total supply value")
	}

	return totalSupply, nil
}

// Convert total supply from raw value to human-readable format
func convertToHumanReadable(totalSupply *big.Int, decimals int) string {
	// Create a factor to divide by 10^decimals
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	// Divide totalSupply by 10^decimals
	denomTotalSupply := new(big.Int).Div(totalSupply, factor)

	// Return the result as a string
	return denomTotalSupply.String()
}

func supplyHandler(w http.ResponseWriter, r *http.Request) {
	// Set address and API key
	contractAddress := "0x240D6FAF8c3B1A7394e371792A3bf9D28DD65515"
	apiKey := os.Getenv("ETHERSCAN_API_KEY")

	totalSupplyRaw, err := fetchTotalSupply(contractAddress, apiKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// token's decimals
	decimals := 9

	humanReadableTotalSupply := convertToHumanReadable(totalSupplyRaw, decimals)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(humanReadableTotalSupply))
}

func main() {
	http.HandleFunc("/supply", supplyHandler)
	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
