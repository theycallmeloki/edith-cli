package edith

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	
)

var arbiApiKey string = ""
var etherApiKey string = ""
var githubToken string = ""

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure secret key(s) for use with edithctl",
	Long: `
Cache your login credentials for your managed services 
(openai, chroma, dockerhub, operators, ingress, egress, gh, do) 
in your host machine. 

(In the case you do not have managed services, 
	we'll setup dependencies to lazy install 
	so you can infrence locally. Be sure to run
    edith install to get started.)

edith configure
edith configure --arbiApiKey=<YOUR_ARBI_API_KEY>
edith configure --etherApiKey=<YOUR_ETHER_API_KEY>
	`,
	Run: func(cmd *cobra.Command, args []string) {

		// if arbiApiKey == "-" {
		// 	// fmt.Println("Enter your arbi API key")
		// 	scanner := bufio.NewScanner(os.Stdin)
		// 	if scanner.Scan() {
		// 		arbiApiKey = scanner.Text()
		// 	}
		// 	viper.Set("arbiApiKey", arbiApiKey)

		// }

		// if etherApiKey == "-" {
		// 	// fmt.Println("Enter your ether API key")
		// 	scanner := bufio.NewScanner(os.Stdin)
		// 	if scanner.Scan() {
		// 		etherApiKey = scanner.Text()
		// 	}
		// 	viper.Set("etherApiKey", etherApiKey)

		// }

		if githubToken == "-" {
			// fmt.Println("Enter your github token")
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				githubToken = scanner.Text()
			}
			viper.Set("githubToken", githubToken)

		}

		WriteLocalConfig()

		// fmt.Println("\n")
		// fmt.Println("Configuration complete! You can now use the CLI without specifying your secret key(s).")
		fmt.Println()

	},
}

func init() {

	// edith.SetEdithConfigFile()

	// Add a new flag to the root command
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(pipelineCmd)

	// Accepting flag for Secret API Key
	// configureCmd.Flags().StringVarP(&arbiApiKey, "arbiApiKey", "a", "", "Arbi API Key")
	// configureCmd.Flags().StringVarP(&etherApiKey, "etherApiKey", "e", "", "Ether API Key")
	configureCmd.Flags().StringVarP(&githubToken, "githubToken", "g", "", "Github Token")
}
