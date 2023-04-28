package edith

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	// "os/exec"
	"strings"

	"github.com/hjson/hjson-go"
	"github.com/spf13/cobra"
)

// The main command group
var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Utilities for developing pipelines locally.",
}

// The 'run' command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run python file locally as if it were a pipeline",
	Run: func(cmd *cobra.Command, args []string) {

		hjsonFile := "edith-config.hjson"

		// Check if the HJSON file exists

		_, err := os.Stat(hjsonFile)

		// If the HJSON file doesn't exist or there is an error, read from stdin
		if os.IsNotExist(err) {
			fmt.Println("HJSON file not found, reading from stdin...")
			hjsonBytes, _ := ioutil.ReadAll(os.Stdin)
			err = ioutil.WriteFile(hjsonFile, hjsonBytes, 0644)
			if err != nil {
				fmt.Printf("Error writing HJSON to file: %v\n", err)
				os.Exit(1)
			}
		} else if err != nil {
			fmt.Printf("Error checking HJSON file: %v\n", err)
			os.Exit(1)
		}

		hjsonBytes, err := ioutil.ReadFile(hjsonFile)

		// Convert HJSON to JSON
		var jsonObj map[string]interface{}
		if err := hjson.Unmarshal(hjsonBytes, &jsonObj); err != nil {
			fmt.Printf("Error parsing HJSON: %v\n", err)
			os.Exit(1)
		}

		// Call the recursive function to print the nested structure
		printNestedStructure(jsonObj, "")

		// Collect values from the HJSON file, and print them
		// We will then proceed to create a container using Edith's API and run the pipeline locally

		// read containerName from HJSON
		containerName, err := getValueAtKeyPath(jsonObj, "containerName", ".")
		if err != nil {
			fmt.Printf("Error getting value at key path: %v\n", err)
		} else {
			fmt.Printf("Building container: %v\n", containerName)
		}

		// read prePushHook from HJSON
		prePushHook, err := getValueAtKeyPath(jsonObj, "prePushHook.cmdList", ".")
		if err != nil {
			fmt.Printf("Error getting value at key path: %v\n", err)
		} else {
			// for each command in the command list, print it, then run it
			for _, cmd := range prePushHook.([]interface{}) {
				fmt.Printf("%v\n", cmd)

				actualCmd := strings.Split(cmd.(string), " ")
				runCmd := exec.Command(actualCmd[0], actualCmd[1:]...)
				runCmd.Stdout = os.Stdout
				runCmd.Stderr = os.Stderr
				err = runCmd.Run()
				if err != nil {
					fmt.Println("Error running pre-push hook: \n", err)
				}

			}
		}

		// read input files from the current directory
		builderPath := "."
		builderFiles, err := ioutil.ReadDir(builderPath)
		if err != nil {
			fmt.Printf("Error reading directory %s: %v\n", builderPath, err)
			return
		} 

		inputFiles := make([]map[string]string, 0)
		for _, bf := range builderFiles {
			filePath := filepath.Join(builderPath, bf.Name())
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", filePath, err)
				continue
			}
			inputFiles = append(inputFiles, map[string]string{"filename": bf.Name(), "data": string(data)})
		}

		fmt.Println(inputFiles)

		buildPayload := map[string]interface{}{
			"files": filesToFileDict(inputFiles),
			"name":  containerName,
		}

		buildPayloadBytes, _ := json.Marshal(buildPayload)
		buildPayload["tag"] = generateMD5Hash(string(buildPayloadBytes))
		
		builderImageTag := "laneone/edith-images:" + buildPayload["name"].(string) + "_" + buildPayload["tag"].(string)
		// print builderImageTag
		fmt.Println(builderImageTag)

		// TODO: first check if you're able to access Edith baseurl, 
		// prefer to use Edith's docker daemon to build the container
		
		// if not, then use local docker daemon to build the container
		
		

		// TODO: Unmount any repos that are already mounted

		// TODO: Implement MOUNT_SERVER functionality as in the original Python script

		// entrypoint := args[0]
		// entrypointArgs := args[1:]

		// pythonCmd := exec.Command("python3", append([]string{entrypoint}, entrypointArgs...)...)
		// pythonCmd.Stdout = os.Stdout
		// pythonCmd.Stderr = os.Stderr
		// err = pythonCmd.Run()
		// if err != nil {
		// 	fmt.Println("Error running Python script:", err)
		// }

		// TODO: Implement unmounting repos and stopping the mount server as in the original Python script
	},
}

// instead of build, you're looking at deploy instead

// // The 'build' command
// var buildCmd = &cobra.Command{
// 	Use:   "build",
// 	Short: "Build pachyderm pipeline",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		name, _ := cmd.Flags().GetString("name")
// 		description, _ := cmd.Flags().GetString("description")
// 		image, _ := cmd.Flags().GetString("image")
// 		inputRepo, _ := cmd.Flags().GetStringSlice("input_repo")
// 		entrypoint, _ := cmd.Flags().GetString("entrypoint")
// 		entrypointArgs := args

// 		cmdStr := "python " + entrypoint + " " + strings.Join(entrypointArgs, " ")
// 		cmdList := strings.Split(cmdStr, " ")

// 		pipeline := map[string]interface{}{
// 			"pipeline":   map[string]string{"name": name},
// 			"description": description,
// 			"input":      map[string]interface{}{},
// 			"transform": map[string]interface{}{
// 				"image": image,
// 				"cmd":   cmdList,
// 			},
// 		}

// 		pipelineInputs := []map[string]interface{}{}
// 		for _, i := range inputRepo {
// 			splitInput := strings.Split(i, "@")
// 			repo := splitInput[0]
// 			branch := splitInput[1]
// 			pipelineInputs = append(pipelineInputs, map[string]interface{}{
// 				"pfs": map[string]string{"repo": repo, "branch": branch, "glob": "/"},
// 			})
// 		}

// 		if len(inputRepo) > 1 {
// 			pipeline["input"] = map[string]interface{}{"cross": pipelineInputs}
// 		} else {
// 			pipeline["input"] = pipelineInputs[0]
// 		}

// 		pipelineJson, _ := json.MarshalIndent(pipeline, "", "  ")
// 		fmt.Println(string(pipelineJson))
// 	},
// }

func init() {
	// Add the 'run' and 'build' commands to the main command group
	pipelineCmd.AddCommand(runCmd)
	// pipelineCmd.AddCommand(buildCmd)

	// Define flags for the 'build' command
	// buildCmd.Flags().String("name", "", "Name of pipeline")
	// buildCmd.Flags().String("description", "","description of pipeline")
	// buildCmd.Flags().String("image", "", "Name of docker image to be used for the entrypoint")
	// buildCmd.Flags().StringSlice("input_repo", []string{}, "Input repo(s) - format repo@branch")
	// buildCmd.Flags().String("entrypoint", "", "Path to the entrypoint script")
}
