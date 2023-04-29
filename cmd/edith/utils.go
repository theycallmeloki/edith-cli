package edith

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	// "log"
	"strconv"

	// "net/http"
	"os"
	"path/filepath"
	"strings"

	// "github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/viper"
	"github.com/theycallmeloki/Edith-cli/pkg/edith"
)

// create config file function
func SetEdithConfigFile() {
	// get home directory
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// xdg config directory
	ConfigDir = filepath.Join(home, ".config", ConfigName)

	// create config file
	ConfigPath = filepath.Join(ConfigDir, ConfigName+"."+ConfigType)

	viper.SetConfigFile(ConfigPath)
	viper.SetConfigName(ConfigName)
	viper.SetConfigType(ConfigType)
	viper.AddConfigPath(ConfigDir)
	viper.SetDefault("arbiApiKey", "")
	viper.SetDefault("etherApiKey", "")
}


// Recursive function to print nested structure
func printNestedStructure(m map[string]interface{}, prefix string) {
	for key, value := range m {
		// Check if the value is a nested map
		if nestedMap, ok := value.(map[string]interface{}); ok {
			fmt.Printf("%s%s:\n", prefix, key)
			// Call the function recursively with the nested map and an updated prefix
			printNestedStructure(nestedMap, prefix+"  ")
		} else {
			// Print the key and value if it's not a nested map
			fmt.Printf("%s%s: %v\n", prefix, key, value)
		}
	}
}

func getValueAtKeyPath(m map[string]interface{}, keyPath, separator string) (interface{}, error) {
	keys := strings.Split(keyPath, separator)
	var currentValue interface{} = m

	for _, key := range keys {
		switch typedValue := currentValue.(type) {
		case map[string]interface{}:
			if value, ok := typedValue[key]; ok {
				currentValue = value
			} else {
				return nil, fmt.Errorf("key not found: %s", key)
			}
		case []interface{}:
			index, err := strconv.Atoi(key)
			if err != nil || index < 0 || index >= len(typedValue) {
				return nil, fmt.Errorf("invalid index: %s", key)
			}
			currentValue = typedValue[index]
		default:
			return nil, fmt.Errorf("key path not found")
		}
	}

	return currentValue, nil
}

// sanitized input function
func SanitizeInput(query string) string {
	fmt.Print(query)
	reader := bufio.NewReader(os.Stdin)
	inputLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return ""
	}
	inputLine = strings.TrimSpace(inputLine)
	return inputLine
}

// FileExists function, used to look for existing config file
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// EnsureFileExists creates a file if it does not exist
func EnsureFileExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		_, err := os.Create(path)
		return err
	}
	return nil
}

// function to read in the local configuration
func ReadLocalConfig() *edith.Config {

	err := viper.ReadInConfig()

	if err == nil {
		c := edith.Config{}
		c.ArbiApiKey = viper.GetString("arbiApiKey")
		c.EtherApiKey = viper.GetString("etherApiKey")
		return &c
	} else {
		e := edith.Config{}
		return &e
	}
}

// function to write the configuration in viper memory to the local config file
func WriteLocalConfig() {
	_, fErr := os.Stat(ConfigDir)
	if os.IsNotExist(fErr) {
		err := os.Mkdir(ConfigDir, 0700) // since it's a user directory, we don't need to worry about group or other permissions
		if err != nil {
			fmt.Println(err)
		}
	}

	err := EnsureFileExists(ConfigPath)
	if err != nil {
		fmt.Println(err)
	}

	if err := viper.WriteConfig(); err != nil {
		fmt.Println(err)
	}
}


func CollectFiles(root string) ([]map[string]string, error) {
	inputFiles := make([]map[string]string, 0)

	gitignorePath := filepath.Join(root, ".gitignore")
	ignore, err := gitignore.CompileIgnoreFile(gitignorePath)
	if os.IsNotExist(err) {
		ignore = &gitignore.GitIgnore{}
	} else if err != nil {
		return nil, fmt.Errorf("error reading .gitignore file: %v", err)
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Check if the path is ignored
			relPath, _ := filepath.Rel(root, path)
			if !ignore.MatchesPath(relPath) {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return fmt.Errorf("error reading file %s: %v", path, err)
				}
				inputFiles = append(inputFiles, map[string]string{"filename": relPath, "data": string(data)})
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return inputFiles, nil
}

func filesToFileDict(inputFiles []map[string]string) map[string]string {
	output := make(map[string]string)
	for _, f := range inputFiles {
		output[f["filename"]] = f["data"]
	}

	return output
}

func generateMD5Hash(input string) string {
	hasher := md5.New()
	hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}