# Lights

A simple process manager for any executable.
Lights. Camera. Action.

## Usage

```bash
Options:

-action string
	Whether to run the lights ON or OFF (start or stop processes)
-binary string
	Binary to execute process [optional]
-cameras uint
	Number of processes to run based on number of CPUs on machine (default 1)
-name string
	Name of the process to execute
-process string
	Process to execute

Commands:

list
	List all the current processes
```

## Start Process

### With Binary Interpreter

Start a process using a binary interpreter (e.g., node, python, bun):

```bash
lights -cameras 2 -action ON -name server -binary node -process server.js
```

Example:

```bash
$ lights -cameras 2 -action ON -name server -binary node -process server.js
[1] server process started with PID 26702
[2] server process started with PID 26703
[ON] Process took 4 ms
```

### With Direct Executable

Start a process that is directly executable:

```bash
lights -cameras 2 -action ON -name api -process ./api-server
```

Example:

```bash
$ lights -cameras 2 -action ON -name api -process ./api-server
-binary not provided - process must be an executable
[1] api process started with PID 26710
[2] api process started with PID 26711
[ON] Process took 3 ms
```

## Stop Process

```bash
lights -action OFF -name server
```

Example:

```bash
$ lights -action OFF -name server
[26702] server process killed
[26703] server process killed
[OFF] Process took 0 ms
```

## List Processes

List all managed processes:

```bash
lights list
```

List specific process by name:

```bash
lights list -name server
```

Example:

```bash
$ lights list
Name   CPU(%) MEMORY(MB) Uptime
server 0.00   23.48      3.978249s
server 0.00   23.33      3.984656s
api    0.00   18.92      1.234567s
api    0.00   19.01      1.240123s
```

## Build From Source

**Note**: You'll need Go 1.22+ installed

1. Clone repository and build:

```bash
GOEXPERIMENT=jsonv2 go build -ldflags="-w -s" -o lights lights.go
```

2. Move binary to installation directory:

```bash
mv lights /usr/local/bin
```

3. Reload shell configuration:

```bash
source ~/.zshrc
```

## Notes

This is a lightweight implementation of a basic process manager. Features that could be added:

- Pipe all process logs into one unified log file
- Reload a process rather than stopping and starting (restart command)
- Pass environment variables into processes
- Process auto-restart on failure
- Resource limits and monitoring alerts
