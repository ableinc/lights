package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/ableinc/lights/internals/utils"
	"github.com/olexnzarov/processinfo"
	"github.com/struCoder/pidusage"
)

// Map of parsed arguments
var PARSED_ARGUMENTS map[string]interface{} = make(map[string]interface{})
var NUMBER_OF_CPUS int = runtime.NumCPU()

func main() {
	// Define usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	// Number of CPUs
	cameras := flag.Int("cameras", 1, "Number of processes to run based on number of CPUs on machine.")
	// Action - ON or OFF
	action := flag.String("action", "", "Whether to run the lights ON or OFF (start or stop processes).")
	// Name of process
	name := flag.String("name", "", "Name of the process")
	// Process to execute
	process := flag.String("process", "", "Process to execute")
	// Binary
	binary := flag.String("binary", "bun", "Binary to execute process")
	// Check usage before sub-command checks
	if len(os.Args) == 1 {
		showUsage()
	}
	// List
	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	PARSED_ARGUMENTS["LIST_INVOKED"] = false
	command := os.Args[1]
	if command == "list" {
		PARSED_ARGUMENTS["LIST_INVOKED"] = true
		listCommand.Parse(os.Args[2:])
	}
	// Parse command-line arguments
	flag.Parse()
	// Assign arguments
	assignArguments(*cameras, *action, *name, *process, *binary, listCommand)
	// Validate arguments
	validateArguments()
	// Execute Actions
	executeOn()
	executeOff()
	executeList()
	// Exit program
	os.Exit(0)
}

func assignArguments(cameras int, action string, name string, process string, binary string, listCommand *flag.FlagSet) {
	if NUMBER_OF_CPUS < cameras {
		fmt.Printf(
			`The number of CPUs you provided (%d) is greater your machine: %d. We will use the max number of CPUs.\n`,
			cameras,
			NUMBER_OF_CPUS,
		)
		cameras = NUMBER_OF_CPUS
	}
	// Assign parsed arguments to global
	PARSED_ARGUMENTS["cameras"] = cameras
	PARSED_ARGUMENTS["action"] = action
	if name == "" {
		name = process
	}
	PARSED_ARGUMENTS["name"] = name
	PARSED_ARGUMENTS["process"] = process
	PARSED_ARGUMENTS["binary"] = binary
	if listCommand.NArg() < 1 {
		PARSED_ARGUMENTS["list"] = "*"
	} else {
		PARSED_ARGUMENTS["list"] = listCommand.Arg(0)
	}
}

func showUsage() {
	flag.Usage()
	os.Exit(1)
}

func validateArguments() {
	if PARSED_ARGUMENTS["action"] == "ON" {
		if PARSED_ARGUMENTS["process"] == "" {
			fmt.Println(
				"You must provide process (--process)",
			)
			showUsage()
		}
	}
	if PARSED_ARGUMENTS["action"] == "OFF" {
		if PARSED_ARGUMENTS["process"] == "" && PARSED_ARGUMENTS["name"] == "" {
			fmt.Println(
				"You must provide name of process (--process) or name (--name) to terminate.",
			)
			showUsage()
		}
	}
}

func executeOn() {
	if PARSED_ARGUMENTS["action"] != "ON" {
		return
	}
	startTime := time.Now()
	var EXECUTED_PROCESSES []int
	for i := 0; i < PARSED_ARGUMENTS["cameras"].(int); i++ {
		fmt.Printf("Executing: %s %s\n", PARSED_ARGUMENTS["binary"].(string), PARSED_ARGUMENTS["process"].(string))
		cmd := exec.Command(PARSED_ARGUMENTS["binary"].(string), PARSED_ARGUMENTS["process"].(string))
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error running %s process %d: %s\n", PARSED_ARGUMENTS["binary"].(string), i, err)
		}
		EXECUTED_PROCESSES = append(EXECUTED_PROCESSES, cmd.Process.Pid)
	}
	// Create meta data file
	var metaDataFile map[string]interface{} = make(map[string]interface{})
	var processes map[string]interface{} = make(map[string]interface{})
	processes["name"] = PARSED_ARGUMENTS["name"]
	processes["pids"] = EXECUTED_PROCESSES
	processes["startTime"] = time.Now().Unix()
	metaDataFile["processes"] = []map[string]interface{}{processes}
	metaDataFile["updatedAt"] = time.Now().Unix()
	// Write meta data
	utils.WriteMetaDataFile(metaDataFile)
	endTime := time.Now()
	// Calculate duration
	duration := endTime.Sub(startTime)
	// Print duration
	fmt.Printf("[ON] Process took: %s\n", duration)
}

func executeOff() {
	if PARSED_ARGUMENTS["action"] != "OFF" {
		return
	}
	startTime := time.Now()
	var metaDataFile utils.MetaData = utils.ReadMetaDataFile()
	var existingProcessesByName []utils.MetaDataProcesses
	var existingProcessesNotByName []utils.MetaDataProcesses

	for i := 0; i < len(metaDataFile.Processes); i++ {
		if metaDataFile.Processes[i].Name == PARSED_ARGUMENTS["name"] {
			existingProcessesByName = append(existingProcessesByName, metaDataFile.Processes[i])
		} else {
			existingProcessesNotByName = append(existingProcessesNotByName, metaDataFile.Processes[i])
		}
	}
	if len(existingProcessesByName) == 0 {
		log.Fatal("No existing processes found by name provided.")
	}
	for i := 0; i < len(existingProcessesByName[0].Pids); i++ {
		var PID int = existingProcessesByName[0].Pids[i]
		syscall.Kill(PID, syscall.SIGINT)
		fmt.Printf("[%s] Killed process %d\n", PARSED_ARGUMENTS["name"], PID)
	}
	// Update MetaData file
	updatedMetaData := make(map[string]interface{})
	updatedMetaData["processes"] = existingProcessesNotByName
	updatedMetaData["updatedAt"] = time.Now().Unix()
	utils.WriteMetaDataFile(updatedMetaData)
	endTime := time.Now()
	// Calculate duration
	duration := endTime.Sub(startTime)
	// Print duration
	fmt.Printf("[OFF] Process took: %s\n", duration)
}

func executeList() {
	if PARSED_ARGUMENTS["LIST_INVOKED"] == false {
		return
	}
	startTime := time.Now()
	var metaDataFile utils.MetaData = utils.ReadMetaDataFile()
	var existingProcesses []utils.MetaDataProcesses
	for i := 0; i < len(metaDataFile.Processes); i++ {
		if PARSED_ARGUMENTS["list"] == "*" {
			existingProcesses = append(existingProcesses, metaDataFile.Processes[i])
		} else if PARSED_ARGUMENTS["list"] == metaDataFile.Processes[i].Name {
			existingProcesses = append(existingProcesses, metaDataFile.Processes[i])
		}
	}
	if len(existingProcesses) == 0 {
		log.Fatal("No existing processes found by name provided.")
	}
	var NAME string = existingProcesses[0].Name
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "Name\tCPU(%)\tMEMORY(MB)\tUptime")
	for i := 0; i < len(existingProcesses[0].Pids); i++ {
		var PID int = existingProcesses[0].Pids[i]
		var UPTIME time.Duration = time.Now().Sub(time.Unix(existingProcesses[0].StartTime, 0))
		if runtime.GOOS == "darwin" {
			sysInfo, err := pidusage.GetStat(PID)
			if err != nil {
				log.Fatal("Error getting process state: ", err)
			}
			fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%s\n", NAME, sysInfo.CPU, sysInfo.Memory/(1024*1024), UPTIME)
		} else {
			processState, err := processinfo.Get(PID)
			if err != nil {
				log.Fatal("Error getting process state: ", err)
			}
			fmt.Fprintf(w, "%s\t%.2f\t%d\t%s\n", NAME, processState.CPU, processState.Memory/(1024*1024), UPTIME)
		}
	}
	endTime := time.Now()
	// Calculate duration
	duration := endTime.Sub(startTime)
	// Print duration
	fmt.Fprintf(w, "[list] Process took: %s\n", duration)
	w.Flush()
}

func listCurrentRunningProcesses() {
	var metaDataFile utils.MetaData = utils.ReadMetaDataFile()
	var existingProcessesByName []utils.MetaDataProcesses
	var existingProcessesNotByName []utils.MetaDataProcesses

	for i := 0; i < len(metaDataFile.Processes); i++ {
		if metaDataFile.Processes[i].Name == PARSED_ARGUMENTS["name"] {
			existingProcessesByName = append(existingProcessesByName, metaDataFile.Processes[i])
		} else {
			existingProcessesNotByName = append(existingProcessesNotByName, metaDataFile.Processes[i])
		}
	}
}
