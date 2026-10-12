package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type redirectKind int

const (
	noRedirect redirectKind = iota
	redirectStdout
	redirectStderr
)

type parsed struct {
	args   []string
	kind   redirectKind
	target string
}

func parseLine(line string) (parsed, error) {
	var p parsed
	tempArg := ""
	inSingle, inDouble := false, false
	escaped := false
	start := false
	wantFile := false
	hadQuoteOrEscape := false

	flush := func() {
		if wantFile {
			p.target = tempArg
			wantFile = false
		} else {
			p.args = append(p.args, tempArg)
		}
		tempArg = ""
		start = false
		hadQuoteOrEscape = false
	}

	for i := 0; i < len(line); i++ {
		c := line[i]

		if escaped {
			tempArg += string(c)
			escaped = false
			start = true
			continue
		}

		if c == '>' && !inSingle && !inDouble {
			if wantFile {
				return p, errors.New("syntax error near unexpected token `>'")
			}
			if start {
				switch {
				case tempArg == "2" && !hadQuoteOrEscape:
					p.kind = redirectStderr
				case tempArg == "1" && !hadQuoteOrEscape:
					p.kind = redirectStdout
				default:
					p.args = append(p.args, tempArg)
					p.kind = redirectStdout
				}
				tempArg = ""
				start = false
				hadQuoteOrEscape = false
			} else {
				p.kind = redirectStdout
			}
			wantFile = true
			continue
		}

		if c == '\\' && !inSingle {
			escaped = true
			hadQuoteOrEscape = true
			continue
		}
		if c == '\'' && !inDouble {
			inSingle = !inSingle
			start = true
			hadQuoteOrEscape = true
			continue
		}
		if c == '"' && !inSingle {
			inDouble = !inDouble
			start = true
			hadQuoteOrEscape = true
			continue
		}
		if inSingle || inDouble {
			tempArg += string(c)
			continue
		}
		if c == ' ' || c == '\t' {
			if start {
				flush()
			}
			continue
		}
		tempArg += string(c)
		start = true
	}

	if start {
		flush()
	}
	if wantFile {
		return p, errors.New("syntax error: missing redirect target")
	}
	return p, nil
}

func runCommand(p parsed) bool {
	args := p.args
	if len(args) == 0 {
		return false
	}

	stdout, stderr := os.Stdout, os.Stderr
	if p.kind != noRedirect {
		f, err := os.OpenFile(p.target, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
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
		case "echo", "type", "exit", "pwd", "cd":
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

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("$ ")
		if !scanner.Scan() {
			break
		}
		p, err := parseLine(scanner.Text())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if runCommand(p) {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
