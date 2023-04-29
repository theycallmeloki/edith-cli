package edith

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	// "path/filepath"

	// "os/exec"
	"strings"

	"github.com/hjson/hjson-go"
	"github.com/spf13/cobra"
	edith "github.com/theycallmeloki/Edith-cli/pkg/edith"
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

				// Check if the command starts with "!", if yes, run it locally
				if actualCmd[0][0] == '!' {
					actualCmd[0] = actualCmd[0][1:]
					runCmd := exec.Command(actualCmd[0], actualCmd[1:]...)
					runCmd.Stdout = os.Stdout
					runCmd.Stderr = os.Stderr
					err = runCmd.Run()
					if err != nil {
						fmt.Println("Error running local pre-push hook: \n", err)
					}
				} else {
					// Stub for running command on the cloud
					fmt.Println("Running command on the cloud:", cmd)
					// Your logic for running the command on the cloud goes here

					// Run the command on the cloud using the runCommandOnCloud function
					var stdinBytes []byte = nil
					// if strings.Contains(strings.Join(actualCmd, " "), "-") {
						// stdinBytes, _ = ioutil.ReadAll(os.Stdin)
					// }
					err := runCommandOnCloud(actualCmd[0], actualCmd[1:], stdinBytes)
					if err != nil {
						fmt.Println("Error running pre-push hook on the cloud: \n", err)
					}
				}
			}
		}


		// read input files from the current directory
		builderPath := "."
	

		// fmt.Println(inputFiles)
		inputFiles, err := CollectFiles(builderPath)
		if err != nil {
			fmt.Printf("Error collecting files from directory %s: %v\n", builderPath, err)
			return
		}

		// printNestedStructure(inputFiles, "")

		// we need to drop the "edith" key from the inputFiles map

		// filteredInputFiles := make([]map[string]string, 0)
		// for _, file := range inputFiles {
			
		// 	// rule system for filtering out files
			
		// 	if file["filename"] != "edith" {
		// 		filteredInputFiles = append(filteredInputFiles, file)
		// 	}

		// }

		// fmt.Println(filteredInputFiles)


		buildPayload := map[string]interface{}{
			"files": filesToFileDict(inputFiles),
			"name":  containerName,
		}

		buildPayloadBytes, _ := json.Marshal(buildPayload)
		buildPayload["tag"] = generateMD5Hash(string(buildPayloadBytes))
		
		builderImageTag := "laneone/edith-images:" + buildPayload["name"].(string) + "_" + buildPayload["tag"].(string)
		// print builderImageTag
		fmt.Println(builderImageTag)


		// Send the build payload to the /buildContainer endpoint
		url := edith.EDITH_BASE_URL + "/buildContainer"

		payloadBytes, _ := json.Marshal(buildPayload)
		payloadReader := bytes.NewReader(payloadBytes)
		resp, err := http.Post(url, "application/json", payloadReader)

		if err != nil {
			fmt.Printf("Error sending build payload: %v\n", err)
			return
		}

		defer resp.Body.Close()

		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			return
		}

		
		var responseMap map[string]interface{}
		err = json.Unmarshal(responseBody, &responseMap)
		if err != nil {
			fmt.Printf("Error unmarshalling response body: %v\n", err)
			return
		}

		edithImageTag, err := getValueAtKeyPath(responseMap, "edithImageTag", ".")
		if err != nil {
			fmt.Printf("Error getting value at key path: %v\n", err)
		} else {
			// fmt.Printf("Edith image tag: %v\n", edithImageTag)
			fmt.Printf("%s\n", edithImageTag)
		}

		if edithImageTag != builderImageTag {
			fmt.Printf("Warning! Edith image tag does not match builder image tag, things might not add up\n")
		}

		pipelineConfigInput, err := getValueAtKeyPath(jsonObj, "postPushHook.pipeline.input", ".")
		if err != nil {
			fmt.Printf("Error getting value at key path: %v\n", err)
		} else {
			fmt.Printf("Pipeline input: %v\n", pipelineConfigInput)
		}


		// TODO: first check if you're able to access Edith baseurl, 
		// prefer to use Edith's docker daemon to build the container
		
		// if not, then use local docker daemon to build the container
		
		// collect postPushHookPipeline from HJSON
		postPushHookPipeline, err := getValueAtKeyPath(jsonObj, "postPushHook.pipeline", ".")
		if err != nil {
			fmt.Printf("Error getting value at key path: %v\n", err)
		} else {
			fmt.Printf("Running pipeline: %v\n", postPushHookPipeline)

			// run the pipeline

			transformCmd, err := getValueAtKeyPath(jsonObj, "postPushHook.pipeline.transform.cmd", ".")

			pachctlPipelinePayload := map[string]interface{}{
				"pipeline": map[string]interface{}{
					"name": containerName,
				},
				"input": pipelineConfigInput,
				"transform": map[string]interface{}{
					"cmd": transformCmd,
					"image_pull_secrets": []string{"laneonekey"},
					"image": edithImageTag,
				},
			}

			// Prepare the payload with the command arguments
			stdinBytes, _ := json.Marshal(pachctlPipelinePayload)

			// Prepare the payload with the command arguments
			pachctlCommandLinePayload := map[string]interface{}{
				"args":  []string{"create", "pipeline", "-f", "-"},
				"stdin": string(stdinBytes),
			}

			// Marshal the payload to JSON
			pachctlCommandLinePayloadBytes, _ := json.Marshal(pachctlCommandLinePayload)

			// Create a reader from the JSON payload bytes
			pachctlCommandLinePayloadReader := bytes.NewReader(pachctlCommandLinePayloadBytes)

			// Set the target URL for the API
			pachctlUrl := edith.EDITH_BASE_URL + "/runPachctlCommand"

			// Make the POST request
			pachctlResp, err := http.Post(pachctlUrl, "application/json", pachctlCommandLinePayloadReader)

			if err != nil {
				fmt.Printf("Error making API request: %v\n", err)
				return
			}

			defer pachctlResp.Body.Close()


			// Read the entire response body
			responseBody2, err := ioutil.ReadAll(pachctlResp.Body)
			if err != nil {
				fmt.Printf("Error reading response body: %v\n", err)
				return
			}

			// Print the response status and content
			fmt.Printf("API response status: %s\n", pachctlResp.Status)
			fmt.Printf("API response body: %s\n", responseBody2)
			
		}

		postPushHookK8s, err := getValueAtKeyPath(jsonObj, "postPushHook.k8s", ".")
		if err != nil {
			fmt.Printf("Error getting value at key path: %v\n", err)
		} else {
			fmt.Printf("Running k8s: %v\n", postPushHookK8s)

			replicas, err := getValueAtKeyPath(jsonObj, "postPushHook.k8s.replicas", ".")
			if err != nil {
				fmt.Printf("Error getting value at key path: %v\n", err)
			}

			minReadySeconds, err := getValueAtKeyPath(jsonObj, "postPushHook.k8s.minReadySeconds", ".")
			if err != nil {
				fmt.Printf("Error getting value at key path: %v\n", err)
			}

			containerPort, err := getValueAtKeyPath(jsonObj, "postPushHook.k8s.containerPort", ".")
			if err != nil {
				fmt.Printf("Error getting value at key path: %v\n", err)
			}

			servicePort, err := getValueAtKeyPath(jsonObj, "postPushHook.k8s.servicePort", ".")
			if err != nil {
				fmt.Printf("Error getting value at key path: %v\n", err)
			}

			nodePort, err := getValueAtKeyPath(jsonObj, "postPushHook.k8s.nodePort", ".")
			if err != nil {
				fmt.Printf("Error getting value at key path: %v\n", err)
			}

			containerNameStr, ok := containerName.(string)
			if !ok {
				fmt.Println("Error: containerName is not a string")
				return
			}
			serviceName := fmt.Sprintf("%v-svc", containerNameStr)


			blueGreenDeployment := map[string]interface{}{
				"apiVersion": "ctl.enisoc.com/v1",
				"kind":       "BlueGreenDeployment",
				"metadata": map[string]interface{}{
					"name": containerName,
					"labels": map[string]interface{}{
						"app": containerName,
					},
				},
				"spec": map[string]interface{}{
					"replicas":         replicas,
					"minReadySeconds":  minReadySeconds,
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": containerName,
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": containerName,
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  containerName,
									"image": edithImageTag,
									"ports": []interface{}{
										map[string]interface{}{
											"containerPort": containerPort,
										},
									},
									"imagePullPolicy": "Always",
								},
							},
							"imagePullSecrets": []interface{}{
								map[string]interface{}{
									"name": "laneonekey",
								},
							},
						},
					},
					"service": map[string]interface{}{
						"metadata": map[string]interface{}{
							"name": serviceName,
							"labels": map[string]interface{}{
								"app": containerName,
							},
						},
						"spec": map[string]interface{}{
							"type": "NodePort",
							"selector": map[string]interface{}{
								"app": containerName,
							},
							"ports": []interface{}{
								map[string]interface{}{
									"port":       servicePort,
									"targetPort": containerPort,
									"protocol":   "TCP",
									"nodePort":   nodePort,
								},
							},
						},
					},
				},
			}

			// print the JSON payload

			fmt.Println("Blue Green Deployment JSON Payload:")

			printNestedStructure(blueGreenDeployment, "")

			blueGreenDeploymentBytes, _ := json.Marshal(blueGreenDeployment)

			// Create a reader from the JSON payload bytes

			
			// write the file to disk
			
			// err = ioutil.WriteFile("blueGreenDeployment.json", blueGreenDeploymentBytes, 0644)
			
			// Prepare the payload with the command arguments
			kubectlCommandLinePayload := map[string]interface{}{
				"args":  []string{"apply", "-f", "-"},
				"stdin": string(blueGreenDeploymentBytes),
			}

			kubectlCommandLinePayloadBytes, err := json.Marshal(kubectlCommandLinePayload)
			if err != nil {
				fmt.Printf("Error marshaling kubectl command line payload: %v\n", err)
				return
			}

			// Create a reader from the JSON payload bytes
			kubectlCommandLinePayloadReader := bytes.NewReader(kubectlCommandLinePayloadBytes)
			
			// kubectlCommandLinePayloadReader := bytes.NewReader(kubectlCommandLinePayload)
			// Set the target URL for the API

			blueGreenDeploymentUrl := edith.EDITH_BASE_URL + "/runKubectlCommand"

			// Make the POST request

			blueGreenDeploymentResp, err := http.Post(blueGreenDeploymentUrl, "application/json", kubectlCommandLinePayloadReader)
			
			if err != nil {
				fmt.Printf("Error making API request: %v\n", err)
				return
			}

			defer blueGreenDeploymentResp.Body.Close()

			// Read the entire response body

			responseBody3, err := ioutil.ReadAll(blueGreenDeploymentResp.Body)

			if err != nil {
				fmt.Printf("Error reading response body: %v\n", err)
				return
			}

			// Print the response status and content
			fmt.Printf("API response status: %s\n", blueGreenDeploymentResp.Status)
			fmt.Printf("API response body: %s\n", responseBody3)
		}

		

		// TODO: Unmount any repos that are already mounted

		// TODO: Implement MOUNT_SERVER functionality as in the original Python script

		
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
