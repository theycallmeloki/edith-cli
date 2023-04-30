package edith

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	// "time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// "github.com/spf13/viper"
	// etherscan "github.com/nanmu42/etherscan-api"
	"github.com/go-resty/resty/v2"
	edith "github.com/theycallmeloki/Edith-cli/pkg/edith"
)

var walletAddress string
var chain string

var arbCmd = &cobra.Command{
	Use:   "arb",
	Short: "Arbitrage your assets",
	Long: `
	Arbitrage your assets across multiple exchanges.
	
	`,
	Run: func(cmd *cobra.Command, args []string) {

		if walletAddress == "" {
			walletAddress = viper.GetString("wallet")
		}
		if chain == "" {
			chain = viper.GetString("chain")
		}

		if githubToken == "" {
			githubToken = viper.GetString("githubToken")
		}


		if arbiApiKey == "" {
			arbiApiKey = viper.GetString("arbiApiKey")
			if arbiApiKey == "" {
				
				// going to ask edith for the api key



				// Initialize a Resty client
				client := resty.New()

				// Set the base URL for the Resty client
				client.SetHostURL(edith.EDITH_BASE_URL)

				// Set the Authorization header with the GitHub token
				client.SetHeader("Authorization", "Bearer "+githubToken)

				// Define the desired key type
				keyType := "arbiscan"

				// Send the request
				response, err := client.R().
				SetHeader("Authorization", "Bearer "+githubToken).
				SetQueryParams(map[string]string{
					"keyType": keyType,
				}).
				Get(edith.EDITH_BASE_URL+"/apiKey")
				if err != nil {
					log.Fatalf("Request failed: %v", err)
				}

				if response.StatusCode() != http.StatusOK {
					log.Fatalf("Request returned non-200 status code: %d", response.StatusCode())
				}

				// Read the response body
				body := response.Body()

				// Parse the response JSON as a generic map
				var responseData map[string]interface{}

				printNestedStructure(responseData,"")

				err = json.Unmarshal(body, &responseData)
				if err != nil {
					log.Fatalf("Failed to parse JSON response: %v", err)
				}

				// Extract the API key from the map
				apiKey, ok := responseData["apiKey"].(string)
				if !ok {
					log.Fatal("Failed to extract API key from the response")
				}


				

				// Save the received API key to Viper
				// viper.Set("arbiApiKey", apiKeyResponse.APIKey)

				// Use the API key in your application
				// apiKey := viper.GetString("arbiApiKey")
				// fmt.Printf("API key for %s: %s\n", keyType, apiKey)
				viper.Set("arbiApiKey", apiKey)
				arbiApiKey = apiKey
			


			}
		}
		if etherApiKey == "" {
			etherApiKey = viper.GetString("etherApiKey")
			if etherApiKey == "" {

				// Initialize a Resty client
				client := resty.New()

				// Set the base URL for the Resty client
				client.SetHostURL(edith.EDITH_BASE_URL)

				// Set the Authorization header with the GitHub token
				client.SetHeader("Authorization", "Bearer "+githubToken)

				// Define the desired key type
				keyType := "etherscan"

				// Send the request
				response, err := client.R().
				SetHeader("Authorization", "Bearer "+githubToken).
				SetQueryParams(map[string]string{
					"keyType": keyType,
				}).
				Get(edith.EDITH_BASE_URL+"/apiKey")
				if err != nil {
					log.Fatalf("Request failed: %v", err)
				}

				if response.StatusCode() != http.StatusOK {
					log.Fatalf("Request returned non-200 status code: %d", response.StatusCode())
				}

				// Read the response body
				body := response.Body()

				// Parse the response JSON as a generic map
				var responseData map[string]interface{}

				printNestedStructure(responseData,"")

				err = json.Unmarshal(body, &responseData)
				if err != nil {
					log.Fatalf("Failed to parse JSON response: %v", err)
				}

				// Extract the API key from the map
				apiKey, ok := responseData["apiKey"].(string)
				if !ok {
					log.Fatal("Failed to extract API key from the response")
				}


				

				// Save the received API key to Viper
				// viper.Set("arbiApiKey", apiKeyResponse.APIKey)

				// Use the API key in your application
				// apiKey := viper.GetString("arbiApiKey")
				// fmt.Printf("API key for %s: %s\n", keyType, apiKey)
				viper.Set("etherApiKey", apiKey)
				etherApiKey = apiKey
			
				
			}

		}
	
		
		var base_url string
		if chain == "eth" {
			base_url = fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&page=1&offset=10000&sort=desc&apikey=%s", walletAddress, etherApiKey)
		} else if chain == "arb" {
			base_url = fmt.Sprintf("https://api.arbiscan.io/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&page=1&offset=10000&sort=desc&apikey=%s", walletAddress, arbiApiKey)
		} else if chain == "canto" {
			base_url = fmt.Sprintf("https://evm.explorer.canto.io/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&page=1&offset=10000&sort=desc", walletAddress)
		} else {
			fmt.Println("Invalid chain")
			return
		}

		req4, _ := http.NewRequest("GET", base_url, nil)

		res4, _ := http.DefaultClient.Do(req4)

		defer res4.Body.Close()
		body4, err := ioutil.ReadAll(res4.Body)
		if err != nil {
			fmt.Println("Error reading API response:", err)
			return
		}

		// Check if the API response is a valid JSON
		var jsonResult map[string]interface{}
		if err := json.Unmarshal(body4, &jsonResult); err != nil {
			fmt.Println("Error parsing API response:", err)
			return
		}

		// Check if the API response contains an error message
		if msg, ok := jsonResult["message"]; ok && msg == "NOTOK" {
			fmt.Println("Error from API:", jsonResult["result"])
			return
		}

		fmt.Println(string(body4))
		

	},
}

func init() {

	rootCmd.AddCommand(arbCmd)

	// Accepting flag for Wallet
	rootCmd.PersistentFlags().StringVar(&walletAddress, "wallet", "", "Wallet address")

	// Accepting flag for Chain
	rootCmd.PersistentFlags().StringVar(&chain, "chain", "eth", "Chain name (eth/arb/canto)")

	// Accepting flag for Arbiscan API Key
	rootCmd.PersistentFlags().StringVar(&arbiApiKey, "arbiApiKey", "", "ArbiScan API Key")

	// Accepting flag for Etherscan API Key
	rootCmd.PersistentFlags().StringVar(&etherApiKey, "etherApiKey", "", "EtherScan API Key")
}
