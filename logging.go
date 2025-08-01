package main

import "fmt"

const (
	RED   string = "\033[31m"
	BLUE  string = "\033[34m"
	CYAN  string = "\033[36m"
	PINK  string = "\033[35m"
	GREEN string = "\033[32m"
	RESET string = "\033[0m"
)
const (
	ERROR    int = 0
	ATTEMPT  int = 1
	COMPLETE int = 2
	STEP     int = 3
)

func log(logType int, content string, addSeparator bool) bool {
	switch logType {
	case ERROR:
		fmt.Printf("%s>> %s%s\n", RED, RESET, content)
	case ATTEMPT:
		fmt.Printf("%s>> %s%s\n", BLUE, RESET, content)
	case COMPLETE:
		fmt.Printf("%s>> %s%s\n", GREEN, RESET, content)
	case STEP:
		fmt.Printf("	%s•%s %s", CYAN, RESET, content)
	default:
		return false
	}

	if addSeparator {
		fmt.Printf("%s-------------------------------------------------%s\n", PINK, RESET)
	}
	return true
}
