package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type process struct {
	id      int
	pid     int
	command string
	status  string
}

var (
	Process   = []*process{}
	processMu sync.Mutex
)

func processMangemnt(state int) {
	processMu.Lock()
	if state == 1 {
		for i := 0; i < len(Process); i++ {
			c := ""
			if Process[i].status == "Done" {
				c = " "
			} else {
				c = "&"
			}
			if i == len(Process)-1 {
				fmt.Printf("[%v]+  %v                 %v %v\n", Process[i].id, Process[i].status, Process[i].command, c)
			} else if i == len(Process)-2 {
				fmt.Printf("[%v]-  %v                 %v %v\n", Process[i].id, Process[i].status, Process[i].command, c)
			} else {
				fmt.Printf("[%v]   %v                 %v %v\n", Process[i].id, Process[i].status, Process[i].command, c)
			}
			if Process[i].status == "Done" {
				Process = append(Process[:i], Process[i+1:]...)
				i--
			}
		}
	} else {
		for i := 0; i < len(Process); i++ {
			if Process[i].status == "Done" {
				if i == len(Process)-1 {
					fmt.Printf("[%v]+  %v                 %v \n", Process[i].id, Process[i].status, Process[i].command)
				} else if i == len(Process)-2 {
					fmt.Printf("[%v]-  %v                 %v \n", Process[i].id, Process[i].status, Process[i].command)
				} else {
					fmt.Printf("[%v]  %v                 %v \n", Process[i].id, Process[i].status, Process[i].command)
				}
				Process = append(Process[:i], Process[i+1:]...)
				i--
			}
		}
	}
	processMu.Unlock()
}

func runCommand(p parsed) bool {
	args := p.args
	if len(args) == 0 {
		return false
	}
	if p.pipe {
		runPipeline(p.commands)
		return false
	}
	stdout, stderr := os.Stdout, os.Stderr
	if p.kind != noRedirect {
		flags := os.O_CREATE | os.O_WRONLY
		if p.appends {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}
		f, err := os.OpenFile(p.target, flags, 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return false
		}
		defer f.Close()
		if p.kind == redirectStderr {
			stderr = f
		} else {
			stdout = f
		}
	}

	switch args[0] {
	case "exit":
		return true

	case "echo":
		fmt.Fprintln(stdout, strings.Join(args[1:], " "))

	case "type":
		if len(args) != 2 {
			fmt.Fprintln(stdout, "wrong use to type command")
			return false
		}
		switch args[1] {
		case "echo", "type", "exit", "pwd", "cd", "jobs":
			fmt.Fprintf(stdout, "%s is a shell builtin\n", args[1])
		default:
			found := false
			for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
				path := filepath.Join(dir, args[1])
				info, err := os.Stat(path)
				if err != nil {
					continue
				}
				if !info.IsDir() && info.Mode().Perm()&0o111 != 0 {
					fmt.Fprintf(stdout, "%s is %s\n", args[1], path)
					found = true
					break
				}
			}
			if !found {
				fmt.Fprintf(stdout, "%s: not found\n", args[1])
			}
		}

	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return false
		}
		fmt.Fprintln(stdout, dir)

	case "cd":
		dir := ""
		if len(args) <= 1 || args[1] == "~" {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintln(stderr, err)
				return false
			}
			dir = home
		} else {
			dir = args[1]
		}
		if err := os.Chdir(dir); err != nil {
			switch {
			case errors.Is(err, fs.ErrNotExist):
				fmt.Fprintf(stderr, "cd: %s: No such file or directory\n", dir)
			case errors.Is(err, fs.ErrPermission):
				fmt.Fprintf(stderr, "cd: %s: Permission denied\n", dir)
			default:
				fmt.Fprintln(stderr, err)
			}
		}
	case "jobs":
		processMangemnt(1)

	default:

		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		if p.background {
			var prc process
			err := cmd.Start()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return false
			}
			prc.pid = cmd.Process.Pid
			prc.command = strings.Join(args, " ")
			prc.status = "Running"
			processMu.Lock()
			if len(Process) == 0 {
				prc.id = 1
			} else {
				prc.id = Process[len(Process)-1].id + 1
			}
			Process = append(Process, &prc)
			processMu.Unlock()

			go func() {
				_ = cmd.Wait()
				processMu.Lock()
				prc.status = "Done"
				processMu.Unlock()
			}()

			fmt.Printf("[%v] %v\n", prc.id, prc.pid)
		} else {
			cmd.Stdin = os.Stdin
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			err := cmd.Run()
			var exitErr *exec.ExitError
			if errors.Is(err, exec.ErrNotFound) {
				fmt.Fprintf(stderr, "%s: command not found\n", args[0])
			} else if err != nil && !errors.As(err, &exitErr) {
				fmt.Fprintln(stderr, err)
			}
		}
	}
	return false
}
