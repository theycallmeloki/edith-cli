package edith

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version string = "-DEV_BUILD"

var ConfigName string = "edith"
var ConfigType string = "json"
var ConfigDir string
var ConfigPath string

var rootCmd = &cobra.Command{
    Use:  "edith",
	Version: version,
    Short: "edith - a CLI to manage your AI agents",
    Long: `

                    _______  ______  __________________         
                    (  ____ \(  __  \ \__   __/\__   __/|\     /|
                    | (    \/| (  \  )   ) (      ) (   | )   ( |
                    | (__    | |   ) |   | |      | |   | (___) |
                    |  __)   | |   | |   | |      | |   |  ___  |
                    | (      | |   ) |   | |      | |   | (   ) |
                    | (____/\| (__/  )___) (___   | |   | )   ( |
                    (_______/(______/ \_______/   )_(   |/     \|
                                             
_________ _______ _________ _______  _______ ___________________________ _______  _______ 
\__   __/(  ____ )\__   __/(  ____ )(  ____ \\__   __/\__   __/\__   __/(  ____ \(  ____ )
   ) (   | (    )|   ) (   | (    )|| (    \/   ) (      ) (      ) (   | (    \/| (    )|
   | |   | (____)|   | |   | (____)|| (_____    | |      | |      | |   | (__    | (____)|
   | |   |     __)   | |   |  _____)(_____  )   | |      | |      | |   |  __)   |     __)
   | |   | (\ (      | |   | (            ) |   | |      | |      | |   | (      | (\ (   
   | |   | ) \ \_____) (___| )      /\____) |___) (___   | |      | |   | (____/\| ) \ \__
   )_(   |/   \__/\_______/|/       \_______)\_______/   )_(      )_(   (_______/|/   \__/
                                                                                          

   
edithctl - v` + version,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
        return initializeConfig(cmd)
    },
    Run: func(cmd *cobra.Command, args []string) {

    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
        os.Exit(1)
    }
}


func initializeConfig(cmd *cobra.Command) error {
    
    // create a spheron configuration file if it doesn't exist
    SetEdithConfigFile()

    if(!FileExists(ConfigPath)) {
        
        fmt.Println("Start by configuring edith with your secret API keys")
        fmt.Println("\n")
        fmt.Println("Example usage: \n")
        fmt.Println("edith configure")
        fmt.Println("edith configure --secret=<YOUR_SECRET_API_KEY>")
        fmt.Println("\n")
    }

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

    return nil
}