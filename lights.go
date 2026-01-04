package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/olexnzarov/processinfo"
	"github.com/struCoder/pidusage"
)

const LIST_COMMAND string = "list"
const METADATA_FILE_NAME string = ".lights.meta"

type LightsCmd struct {
	List    bool
	Cameras uint8
	Action  string
	Name    string
	Process string
	Binary  *string
}

type MetaDataProcess struct {
	Name      string `json:"name"`
	Pid       int    `json:"pids"`
	StartTime int64  `json:"startTime"`
}
type MetaData struct {
	Processes []MetaDataProcess `json:"processes"`
	UpdatedAt int64             `json:"updatedAt"`
}

func main() {
	var commandLineArgs []string = os.Args
	var lightsCmd LightsCmd = LightsCmd{
		List:    false,
		Cameras: 1,
		Action:  "",
		Name:    "",
		Process: "",
		Binary:  nil,
	}
	parseCommandLine(commandLineArgs, &lightsCmd)
	validateLightsCmd(lightsCmd)

	executeList(lightsCmd)
	executeOn(lightsCmd)
	executeOff(&lightsCmd)
}

func usage() {
	var msg string = "Options:\n\n" +
		"-action string\t\t" +
		"Whether to run the lights ON or OFF (start or stop processes)\n" +
		"-binary string\t\t" +
		"Binary to execute process [optional]\n" +
		"-cameras uint\t\t" +
		"Number of processes to run based on number of CPUs on machine (default %d)\n" +
		"-name string\t\t" +
		"Name of the process to execute\n\n" +
		"-process string\t\t" +
		"Process to execute\n\n" +
		"Commands:\n\n" +
		"list\t\t" +
		"List all the current processes\n"
	fmt.Printf(msg, 1)
	os.Exit(0)
}

func parseCommandLine(sysArgs []string, lightsCmd *LightsCmd) {
	if len(sysArgs) < 2 {
		usage()
	}
	var args []string = sysArgs[1:]

	for i := range args {
		switch args[i] {
		case "-action":
			lightsCmd.Action = args[i+1]
			continue
		case "-binary":
			lightsCmd.Binary = &args[i+1]
			continue
		case "-cameras":
			lightsCmd.Cameras = strToNum(args[i+1])
			continue
		case "-name":
			lightsCmd.Name = args[i+1]
			continue
		case "-process":
			lightsCmd.Process = args[i+1]
			continue
		case LIST_COMMAND:
			lightsCmd.List = true
			continue
		default:
			continue
		}
	}
}

func strToNum(numStr string) uint8 {
	num, err := strconv.ParseUint(numStr, 10, 8)
	if err != nil {
		usage()
	}
	return uint8(num)
}

func validateLightsCmd(lightsCmd LightsCmd) {
	if lightsCmd.List {
		return // if we're just listing processes, ignore other validations
	}
	if strings.ToLower(lightsCmd.Action) == "off" && lightsCmd.Name != "" {
		return // if action is off and name is provided, skip other validations
	}
	if lightsCmd.Action == "" {
		usage()
	}
	if lightsCmd.Name == "" {
		usage()
	}
	if lightsCmd.Process == "" {
		usage()
	}
	if lightsCmd.Cameras < 1 {
		fmt.Printf("-cameras must be greater than 1\n")
		os.Exit(1)
	}
	if lightsCmd.Binary == nil {
		fmt.Printf("-binary not provided - process must be an executable\n")
	}
}

func executeOn(lightsCmd LightsCmd) {
	if strings.ToLower(lightsCmd.Action) != "on" {
		return
	}
	var startTime time.Time = time.Now()
	// Check if exisiting metadata file exists
	metadataFile, err := readMetaDataFile()
	if err == nil {
		// Check if process with same name is already running
		for i := range metadataFile.Processes {
			if metadataFile.Processes[i].Name == lightsCmd.Name {
				fmt.Printf("Process %s is already running with PID %d\n", lightsCmd.Name, metadataFile.Processes[i].Pid)
				return
			}
		}
	}
	var executedProcesses []MetaDataProcess = make([]MetaDataProcess, lightsCmd.Cameras)
	var cmd *exec.Cmd
	for i := 0; i < int(lightsCmd.Cameras); i++ {
		if lightsCmd.Binary != nil {
			cmd = exec.Command(*lightsCmd.Binary, lightsCmd.Process)
			if err := cmd.Start(); err != nil {
				fmt.Printf("Error running %s process %d %s\n", *lightsCmd.Binary, i, err.Error())
				continue
			}
		} else {
			cmd = exec.Command(lightsCmd.Process)
			if err := cmd.Start(); err != nil {
				fmt.Printf("Error running %s process %d %s\n", lightsCmd.Process, i, err.Error())
				continue
			}
		}
		if cmd.Process == nil {
			continue
		}
		fmt.Printf("[%d] %s process started with PID %d\n", i+1, lightsCmd.Name, cmd.Process.Pid)
		executedProcesses[i] = MetaDataProcess{
			Name:      lightsCmd.Name,
			Pid:       cmd.Process.Pid,
			StartTime: time.Now().Unix(),
		}
	}
	var endTime time.Time = time.Now()
	if len(metadataFile.Processes) > 0 {
		executedProcesses = append(metadataFile.Processes, executedProcesses...)
	}
	// Create metadata file
	var metadata MetaData = MetaData{
		Processes: executedProcesses,
		UpdatedAt: time.Now().Unix(),
	}
	// Write metadata file
	writeMetaDataFile(metadata)
	var duration time.Duration = endTime.Sub(startTime)
	fmt.Printf("[ON] Process took %d ms\n", duration.Milliseconds())
}

func executeOff(lightsCmd *LightsCmd) {
	if strings.ToLower(lightsCmd.Action) != "off" {
		return
	}
	metadataFile, err := readMetaDataFile()
	if err != nil {
		fmt.Printf("Error reading metadata file: %v\n", err)
		return
	}
	if len(metadataFile.Processes) == 0 {
		fmt.Printf("No existing processing running\n")
		return
	}
	var startTime time.Time = time.Now()
	var k int = 0
	for i := range metadataFile.Processes {
		if metadataFile.Processes[i].Name == lightsCmd.Name {
			syscall.Kill(metadataFile.Processes[i].Pid, syscall.SIGINT)
			fmt.Printf("[%d] %s process killed\n", metadataFile.Processes[i].Pid, metadataFile.Processes[i].Name)
			continue
		}
		// Keep other processes using the same memory location
		metadataFile.Processes[k] = metadataFile.Processes[i]
		k++
	}
	metadataFile.Processes = metadataFile.Processes[:k]
	metadataFile.UpdatedAt = time.Now().Unix()
	writeMetaDataFile(metadataFile)
	var endTime time.Time = time.Now()
	var duration time.Duration = endTime.Sub(startTime)
	fmt.Printf("[OFF] Process took %d ms\n", duration.Milliseconds())
}

func executeList(lightsCmd LightsCmd) {
	if !lightsCmd.List {
		return
	}
	metaDataFile, err := readMetaDataFile()
	if err != nil {
		fmt.Printf("Error reading metadata file: %v\n", err)
		return
	}
	var processes []MetaDataProcess = slices.DeleteFunc(metaDataFile.Processes, func(n MetaDataProcess) bool {
		if lightsCmd.Name == "" {
			return false
		}
		if lightsCmd.Name != "" && n.Name == lightsCmd.Name {
			return false
		}
		return true
	})
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "Name\tCPU(%)\tMEMORY(MB)\tUptime")
	for i := range processes {
		var uptime time.Duration = time.Since(time.Unix(processes[i].StartTime, 0))
		switch runtime.GOOS {
		case "windows":
			sysInfo, err := processinfo.Get(processes[i].Pid)
			if err != nil {
				fmt.Printf("Error getting process state for %d: %v\n", processes[i].Pid, err)
				continue
			}
			fmt.Fprintf(w, "%s\t%.2f\t%d\t%s\n", processes[i].Name, sysInfo.CPU, sysInfo.Memory/(1024*1024), uptime)
		default:
			sysInfo, err := pidusage.GetStat(processes[i].Pid)
			if err != nil {
				fmt.Printf("Error getting process state for %d: %v\n", processes[i].Pid, err)
				continue
			}
			fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%s\n", processes[i].Name, sysInfo.CPU, sysInfo.Memory/(1024*1024), uptime)
		}
	}
	w.Flush()
}

func writeMetaDataFile(data MetaData) {
	// Marshal JSON data
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}

	// Write JSON data to file
	err = os.WriteFile(METADATA_FILE_NAME, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
}

func readMetaDataFile() (MetaData, error) {
	// Check if file exists
	if _, err := os.Stat(METADATA_FILE_NAME); os.IsNotExist(err) {
		return MetaData{}, fmt.Errorf("Metadata file does not exist\n")
	}
	// Read JSON file
	fileContent, err := os.ReadFile(METADATA_FILE_NAME)
	if err != nil {
		return MetaData{}, err
	}
	// Unmarshal JSON data into struct
	var metaData MetaData
	err = json.Unmarshal(fileContent, &metaData)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
		os.Exit(1)
	}
	return metaData, nil
}
