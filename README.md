# Lights

A simple process manager for any executable.
Lights. Camera. Action.

## Usage

```bash
Usage: lights [options]
Options:
  -action string
    	Whether to run the lights ON or OFF (start or stop processes).
  -binary string
    	Binary to execute process (default "bun")
  -cameras int
    	Number of processes to run based on number of CPUs on machine. (default 1)
  -name string
    	Name of the process
  -process string
    	Process to execute
```

## Start Process

```bash
lights -cameras 2 -action ON -name server -process test.ts
```

Example:

```bash
dev@M2-Air lights % ./lights -cameras 2 -action ON -name server -process
Executing: bun test.ts
Executing: bun test.ts
[ON] Process took: 4.713084ms
```

## Stop Process

```bash
lights -action OFF -name server
```

Example:

```bash
dev@M2-Air lights % ./lights -action OFF -name server
[server] Killed process 26702
[server] Killed process 26703
[OFF] Process took: 557.25µs
```

## List Processes

```bash
lights list
```

or

```bash
lights list server
```

Example:

```bash
dev@M2-Air lights % ./lights list
Name   CPU(%) MEMORY(MB) Uptime
server 0.00   23.48      3.978249s
server 0.00   23.33      3.984656s
[list] Process took: 10.734667ms
```

## Build From Source

```bash
go build cmd/lights/lights.go
```

## Notes

This is a lightweight implementation of a basic process manager. A few things it doesn't do, but could with little effort:

- Pipe all process logs into one unified log
- Reload a process, rather than stopping and starting (restart)
- Pass environment variables into process
