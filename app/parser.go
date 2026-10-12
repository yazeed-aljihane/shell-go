package main

func Parser(line string) ([]string, map[string]int) {
	args := []string{}
	tempArg := ""
	checkone := 0
	checktwo := 0
	backslash := 0
	start := 0
	redirect := 0
	inx := 0
	redirecterr := 0
	hadQuoteOrEscape := false
	info := make(map[string]int)

	for i := 0; i < len(line); i++ {

		if backslash == 1 {
			tempArg += line[i : i+1]
			backslash = 0
			start = 1
			continue
		}
		if line[i] == '>' && checkone == 0 && checktwo == 0 && start == 0 {
			redirect = 1
			inx = len(args)
			args = append(args, line[i:i+1])
			continue
		}
		if line[i] == '>' && checkone == 0 && checktwo == 0 && start == 1 {
			start = 0
			redirect = 1
			if tempArg == "1" && !hadQuoteOrEscape {
				tempArg = ""
				inx = len(args)
				args = append(args, line[i:i+1])
			} else if tempArg == "2" && !hadQuoteOrEscape {
				tempArg = ""
				redirecterr = 1
				inx = len(args)
				args = append(args, line[i:i+1])
			} else {
				args = append(args, tempArg)
				tempArg = ""
				inx = len(args)
				args = append(args, line[i:i+1])
			}
			continue
		}

		if line[i] == '\\' && checkone == 0 {
			backslash = 1
			hadQuoteOrEscape = true
			continue
		}

		if line[i] == '\'' && checkone == 0 && checktwo == 0 {
			checkone = 1
			start = 1
			hadQuoteOrEscape = true
			continue
		}
		if line[i] == '"' && checktwo == 0 && checkone == 0 {
			checktwo = 1
			start = 1
			hadQuoteOrEscape = true
			continue
		}

		if line[i] == '\'' && checkone == 1 && checktwo == 0 {
			checkone = 0
			continue
		}

		if line[i] == '"' && checktwo == 1 && checkone == 0 {
			checktwo = 0
			continue
		}

		if checktwo == 1 {
			tempArg += line[i : i+1]
			continue
		}

		if checkone == 1 {
			tempArg += line[i : i+1]
			continue
		}
		if (line[i] == ' ' || line[i] == '	') && start == 0 {
			continue
		}
		if line[i] != ' ' && line[i] != '	' {
			tempArg += line[i : i+1]
			start = 1
		}
		if checkone == 0 && (line[i] == ' ' || line[i] == '	') && start == 1 {
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

	info["index"] = inx
	info["redirect"] = redirect
	info["redirectErr"] = redirecterr
	return args, info
}
