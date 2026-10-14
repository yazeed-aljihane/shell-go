package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runCommand(p parsed) bool {
	args := p.args
	if len(args) == 0 {
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
		case "echo", "type", "exit", "pwd", "cd", "job":
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
	case "job":
		fmt.Printf("")

	default:
		cmd := exec.Command(args[0], args[1:]...)
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
	return false
}
