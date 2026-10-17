package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		processMangemnt()
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
