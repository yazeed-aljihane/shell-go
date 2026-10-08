package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	// TODO: Uncomment the code below to pass the first stage
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("$ ")
		if !scanner.Scan() {
			break
		}
		args := strings.Fields(scanner.Text())
		if len(args) == 0 {
			continue
		}
		if args[0] == "exit" {
			break
		} else if args[0] == "echo" {
			for i := 1; i < len(args); i++ {
				fmt.Printf("%s ", args[i])
			}
			fmt.Printf("\n")
		} else if args[0] == "type" {
			if len(args) <= 1 || len(args) > 2 {
				fmt.Printf("wrong use to type command\n")
				continue
			}
			if args[1] == "echo" || args[1] == "type" || args[1] == "exit" || args[1] == "pwd" || args[1] == "cd" {
				fmt.Printf("%s is a shell builtin\n", args[1])
			} else {
				dirs := filepath.SplitList(os.Getenv("PATH"))
				check := 0
				for i := 0; i < len(dirs); i++ {
					filePath := filepath.Join(dirs[i], args[1])
					fileInfo, err := os.Stat(filePath)
					if err != nil {
						continue
					}
					if !fileInfo.IsDir() && fileInfo.Mode().Perm()&0o111 != 0 {
						fmt.Printf("%s is %s\n", args[1], filePath)
						check = 1
						break
					}
				}
				if check == 0 {
					fmt.Printf("%s: not found\n", args[1])
				}

			}
		} else if args[0] == "pwd" {
			crdir, err := os.Getwd()
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			fmt.Printf("%s\n", crdir)

		} else if args[0] == "cd" {
			_, err := os.Stat(args[1])
			if err != nil {
				fmt.Printf("cd: %s: No such file or directory\n", args[1])
			} else {
				os.Chdir(args[1])
			}
		} else {
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if errors.Is(err, exec.ErrNotFound) {
				fmt.Printf("%s: command not found\n", args[0])
			} else if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}
