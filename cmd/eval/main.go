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
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		msg := fmt.Sprintf("failed to start: %s", err.Error())
		Exit(msg)
	}

	cmd.Wait()

	procState := cmd.ProcessState
	result := model.Run{
		Success: procState.Success(),
		Output:  output.String(),
	}

	if err := ctx.Err(); err != nil {
		result.Success = false
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
	result := model.Run{Message: message}
	content, _ := json.Marshal(result)
	fmt.Println(string(content))
	os.Exit(0)
}

func Matrics(wallTime time.Duration, exitCode int, usage *syscall.Rusage) model.Metrics {
	metrics := model.Metrics{
		WallTime: wallTime,
		ExitCode: exitCode,
		UserTime: time.Duration(usage.Utime.Nano()), // ns
		SysTime:  time.Duration(usage.Stime.Nano()), // ns
		Memory:   usage.Maxrss,                      // kb
	}

	return metrics
}
