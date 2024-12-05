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

	// 4MiB buffer size
	if err := StdinToFile("Main", 4*1024*1024); err != nil {
		Exit("failed to transfer binary to file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	executable := "./Main"
	cmd := exec.CommandContext(ctx, executable)
	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Start(); err != nil {
		msg := fmt.Sprintf("failed to start: %s", err.Error())
		Exit(msg)
	}

	cmd.Wait()

	var testResults []model.TestResult
	if err := json.Unmarshal(output.Bytes(), &testResults); err != nil {
		msg := fmt.Sprintf("failed to unmarshal test results: %s", err.Error())
		Exit(msg)
	}

	procState := cmd.ProcessState
	result := model.EvalResult{
		Success: procState.Success(),
		Output:  testResults,
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

func StdinToFile(filename string, bufferSize int) error {
	outFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0555)
	if err != nil {
		return err
	}
	defer outFile.Close()

	buffer := make([]byte, bufferSize)

	// read binary at once could exceed memory usage
	// so read and write to file in chunks
	for {
		n, err := os.Stdin.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if n > 0 {
			if _, err := outFile.Write(buffer[:n]); err != nil {
				return err
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
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
	result := model.EvalResult{Message: message}
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
