package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"codeberg.org/iklabib/kerat/model"
)

func main() {
	v, ok := os.LookupEnv("TIMEOUT")
	if !ok {
		Exit("env timeout not defined")
	}

	timeout, err := strconv.Atoi(v)
	if err != nil {
		Exit("failed to parse env timeout")
	}

	bin, err := io.ReadAll(os.Stdin)
	if err != nil {
		Exit("failed to read stdin")
	}

	if err := os.WriteFile("Main", bin, 0555); err != nil {
		Exit("failed to write binary to file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	executable := "./Main"
	cmd := exec.CommandContext(ctx, executable)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	if err := cmd.Start(); err != nil {
		msg := fmt.Sprintf("failed to start: %s", err.Error())
		Exit(msg)
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
	result := model.RunResult{Message: message}
	content, _ := json.Marshal(result)
	fmt.Println(string(content))
	os.Exit(0)
}
