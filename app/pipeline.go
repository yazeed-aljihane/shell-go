package main

import (
	"fmt"
	"os"
	"os/exec"
)

// runPipeline handles this stage's two external commands.
func runPipeline(commands [][]string) {
	if len(commands) != 2 || len(commands[0]) == 0 || len(commands[1]) == 0 {
		fmt.Fprintln(os.Stderr, "pipeline requires two commands")
		return
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer reader.Close()
	defer writer.Close()

	first := exec.Command(commands[0][0], commands[0][1:]...)
	first.Stdin, first.Stdout, first.Stderr = os.Stdin, writer, os.Stderr
	second := exec.Command(commands[1][0], commands[1][1:]...)
	second.Stdin, second.Stdout, second.Stderr = reader, os.Stdout, os.Stderr

	if err := second.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if err := first.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		reader.Close()
		writer.Close()
		_ = second.Process.Kill()
		_ = second.Wait()
		return
	}
	// The children now own their copies. Parent copies must not keep the pipe open.
	reader.Close()
	writer.Close()

	firstDone := make(chan struct{})
	go func() {
		_ = first.Wait()
		close(firstDone)
	}()
	_ = second.Wait()
	// A consumer such as head can exit while tail -f is still waiting for data.
	select {
	case <-firstDone:
	default:
		_ = first.Process.Kill()
		<-firstDone
	}
}
