package main

import (
	"bufio"
	"fmt"
	"os"
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
			if args[1] == "echo" || args[1] == "type" || args[1] == "exit" {
				fmt.Printf("%s is a shell builtin\n", args[1])
			} else {
				fmt.Printf("%s: not found\n", args[1])
			}
		} else {
			fmt.Printf("%s: command not found\n", args[0])
		}
	}
}
