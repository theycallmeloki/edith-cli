package edith

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	// "time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// "github.com/spf13/viper"
	// etherscan "github.com/nanmu42/etherscan-api"
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
		if arbiApiKey == "" {
			arbiApiKey = viper.GetString("arbiApiKey")
		}
		if etherApiKey == "" {
			etherApiKey = viper.GetString("etherApiKey")
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
