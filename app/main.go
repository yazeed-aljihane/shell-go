package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
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

		args := []string{}
		fullText := scanner.Text()
		tempArg := ""
		checkone := 0
		checktwo := 0
		backslash := 0
		start := 0
		redirect := 0
		redirecterr := 0
		hadQuoteOrEscape := false
		output := os.Stdout
		outputerr := os.Stderr
		inx := 0
		for i := 0; i < len(fullText); i++ {

			if backslash == 1 {
				tempArg += fullText[i : i+1]
				backslash = 0
				start = 1
				continue
			}
			if fullText[i] == '>' && checkone == 0 && checktwo == 0 && start == 0 {
				redirect = 1
				inx = len(args)
				args = append(args, fullText[i:i+1])
				continue
			}
			if fullText[i] == '>' && checkone == 0 && checktwo == 0 && start == 1 {
				start = 0
				redirect = 1
				if tempArg == "1" && !hadQuoteOrEscape {
					tempArg = ""
					inx = len(args)
					args = append(args, fullText[i:i+1])
				} else if tempArg == "2" && !hadQuoteOrEscape {
					tempArg = ""
					redirecterr = 1
					inx = len(args)
					args = append(args, fullText[i:i+1])
				} else {
					args = append(args, tempArg)
					tempArg = ""
					inx = len(args)
					args = append(args, fullText[i:i+1])
				}
				continue
			}

			if fullText[i] == '\\' && checkone == 0 {
				backslash = 1
				hadQuoteOrEscape = true
				continue
			}

			if fullText[i] == '\'' && checkone == 0 && checktwo == 0 {
				checkone = 1
				start = 1
				hadQuoteOrEscape = true
				continue
			}
			if fullText[i] == '"' && checktwo == 0 && checkone == 0 {
				checktwo = 1
				start = 1
				hadQuoteOrEscape = true
				continue
			}

			if fullText[i] == '\'' && checkone == 1 && checktwo == 0 {
				checkone = 0
				continue
			}

			if fullText[i] == '"' && checktwo == 1 && checkone == 0 {
				checktwo = 0
				continue
			}

			if checktwo == 1 {
				tempArg += fullText[i : i+1]
				continue
			}

			if checkone == 1 {
				tempArg += fullText[i : i+1]
				continue
			}
			if (fullText[i] == ' ' || fullText[i] == '	') && start == 0 {
				continue
			}
			if fullText[i] != ' ' && fullText[i] != '	' {
				tempArg += fullText[i : i+1]
				start = 1
			}
			if checkone == 0 && (fullText[i] == ' ' || fullText[i] == '	') && start == 1 {
				args = append(args, tempArg)
				tempArg = ""
				start = 0
				hadQuoteOrEscape = false
				continue
			}

		}
		if start == 1 {
			args = append(args, tempArg)
		}
		if len(args) == 0 {
			continue
		}
		argLen := len(args)

		if redirect == 1 {
			if len(args) == inx+1 {
				fmt.Fprintf(os.Stderr, "Error where the file?!\n")
				continue
			}
			f, err := os.OpenFile(args[inx+1], os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			if redirecterr == 1 {
				outputerr = f
			} else {
				output = f
			}
			argLen = len(args[0:inx])
		}

		if args[0] == "exit" {
			break
		} else if args[0] == "echo" {

			for i := 1; i < argLen; i++ {
				if i == argLen-1 {
					fmt.Fprintf(output, "%s", args[i])
					continue
				}
				fmt.Fprintf(output, "%s ", args[i])
			}
			fmt.Fprintf(output, "\n")
		} else if args[0] == "type" {
			if argLen <= 1 || argLen > 2 {
				fmt.Fprintf(output, "wrong use to type command\n")
				continue
			}
			if args[1] == "echo" || args[1] == "type" || args[1] == "exit" || args[1] == "pwd" || args[1] == "cd" {
				fmt.Fprintf(output, "%s is a shell builtin\n", args[1])
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
						fmt.Fprintf(output, "%s is %s\n", args[1], filePath)
						check = 1
						break
					}
				}
				if check == 0 {
					fmt.Fprintf(output, "%s: not found\n", args[1])
				}

			}
		} else if args[0] == "pwd" {
			crdir, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			fmt.Fprintf(output, "%s\n", crdir)

		} else if args[0] == "cd" {
			if argLen <= 1 {
				dir, err := os.UserHomeDir()
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					continue
				}
				errCd := os.Chdir(dir)
				if errCd != nil {
					fmt.Fprintf(os.Stderr, "%v\n", errCd)
				}
				continue

			}
			if args[1] == "~" {
				dir, err := os.UserHomeDir()
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					if redirect == 1 {
						output.Close()
					}
					continue
				}
				errCd := os.Chdir(dir)
				if errCd != nil {
					fmt.Fprintf(os.Stderr, "%v\n", errCd)
				}
				if redirect == 1 {
					output.Close()
				}
				continue

			}
			_, err := os.Stat(args[1])
			if errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "cd: %s: No such file or directory\n", args[1])
				if redirect == 1 {
					output.Close()
				}
				continue
			} else if errors.Is(err, fs.ErrPermission) {
				fmt.Fprintf(os.Stderr, "cd: %v: Permission denied\n", args[1])
				if redirect == 1 {
					output.Close()
				}
				continue
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				if redirect == 1 {
					output.Close()
				}
				continue
			}

			errCd := os.Chdir(args[1])
			if errCd != nil {
				fmt.Fprintf(os.Stderr, "%v\n", errCd)
			}

		} else {
			var cmd *exec.Cmd
			if redirect == 1 {
				cmd = exec.Command(args[0], args[1:inx]...)
			} else {
				cmd = exec.Command(args[0], args[1:]...)
			}
			cmd.Stdin = os.Stdin
			cmd.Stdout = output
			cmd.Stderr = outputerr
			err := cmd.Run()
			if errors.Is(err, exec.ErrNotFound) {
				fmt.Fprintf(os.Stderr, "%s: command not found\n", args[0])
			} else if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					continue
				} else {
					fmt.Fprintln(os.Stderr, err)
				}
			}
		}
		if redirect == 1 {
			output.Close()
		}
	}
	if scanner.Err() != nil {
		fmt.Fprintf(os.Stderr, "%v\n", scanner.Err())
	}
}
