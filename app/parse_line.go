package main

import "errors"

type redirectKind int

const (
	noRedirect redirectKind = iota
	redirectStdout
	redirectStderr
)

type parsed struct {
	args    []string
	kind    redirectKind
	appends bool
	target  string
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
			p.appends = i+1 < len(line) && line[i+1] == '>'
			if p.appends {
				i++
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
