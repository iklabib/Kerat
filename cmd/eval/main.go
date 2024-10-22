package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"codeberg.org/iklabib/kerat/model"
)

func main() {
	if len(os.Args) < 2 {
		Exit("arguments not provided")
	}

	v, ok := os.LookupEnv("TIMEOUT")
	if !ok {
		Exit("env timeout not defined")
	}

	timeout, err := strconv.Atoi(v)
	if err != nil {
		Exit("failed to parse env timeout")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	executable := os.Args[1]
	args := os.Args[2:]
	cmd := exec.CommandContext(ctx, executable, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = os.Stdin

	start := time.Now()
	if err := cmd.Start(); err != nil {
		Exit("failed to start")
	}

	cmd.Wait()

	wallTime := time.Since(start)

	procState := cmd.ProcessState
	usage, ok := procState.SysUsage().(*syscall.Rusage)
	if !ok {
		Exit("failed to get usage")
	}

	metrics := model.Metrics{
		WallTime: wallTime,
		ExitCode: procState.ExitCode(),
		UserTime: time.Duration(usage.Utime.Nano()), // ns
		SysTime:  time.Duration(usage.Stime.Nano()), // ns
		Memory:   usage.Maxrss,                      // kb
	}

	result := model.Result{
		Metric: metrics,
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			result.Message = "time limit exceeded"
		} else if errors.Is(err, context.Canceled) {
			result.Message = "canceled"
		} else {
			result.Message = "error"
		}
	}

	marshaled, _ := json.Marshal(result)
	fmt.Println(string(marshaled))
}

func GetSignal(procState *os.ProcessState) (os.Signal, bool) {
	if procState.Exited() {
		return syscall.Signal(0), false
	}

	wt := procState.Sys().(syscall.WaitStatus)
	if wt.Signaled() {
		return wt.Signal(), true
	}

	return syscall.Signal(0), false
}

func Exit(message string) {
	result := model.Result{Message: message}
	content, _ := json.Marshal(result)
	fmt.Println(string(content))
	os.Exit(0)
}
