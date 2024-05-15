package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// File name
var FILE_NAME string = ".lights.meta"

// Define struct to unmarshal JSON data into
type MetaDataProcesses struct {
	Name      string `json:"name"`
	Pids      []int  `json:"pids"`
	StartTime int64  `json:"startTime"`
}
type MetaData struct {
	Processes []MetaDataProcesses `json:"processes"`
	UpdatedAt int64               `json:"updatedAt"`
}

func WriteMetaDataFile(data map[string]interface{}) {
	// Marshal JSON data
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}

	// Write JSON data to file
	err = os.WriteFile(FILE_NAME, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
}

func ReadMetaDataFile() MetaData {
	// Read JSON file
	fileContent, err := os.ReadFile(FILE_NAME)
	if err != nil {
		err := fmt.Errorf("Error reading meta file data file: %v\n", err)
		panic(err)
	}

	// Unmarshal JSON data into struct
	var metaData MetaData
	err = json.Unmarshal(fileContent, &metaData)
	if err != nil {
		err := fmt.Errorf("Error unmarshalling JSON: %v\n", err)
		panic(err)
	}
	return metaData
}
