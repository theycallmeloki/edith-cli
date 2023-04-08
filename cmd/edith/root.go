package edith

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version string = "-DEV_BUILD"

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
    Run: func(cmd *cobra.Command, args []string) {

    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
        os.Exit(1)
    }
}