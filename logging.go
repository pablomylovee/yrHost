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
	ERROR int = 0
	ATTEMPT int = 1
	COMPLETE int = 2
)

func log(logType int, content string, addSeparator bool) string {
	switch logType {
	case ERROR:
		fmt.Printf("%s>> %s%s", RED, RESET, content)
		fmt.Print("\n")
		if addSeparator {
			fmt.Printf("%s-------------------------------------------------%s", PINK, RESET)
		}
		fmt.Print("\n")
		return ""
	case ATTEMPT:
		fmt.Printf("%s>> %s%s", BLUE, RESET, content)
		fmt.Print("\n")
		if addSeparator {
			fmt.Printf("%s-------------------------------------------------%s", PINK, RESET)
		}
		fmt.Print("\n")
		return ""
	case COMPLETE:
		fmt.Printf("%s>> %s%s", GREEN, RESET, content)
		fmt.Print("\n")
		if addSeparator {
			fmt.Printf("%s-------------------------------------------------%s", PINK, RESET)
		}
		fmt.Print("\n")
		return ""
	default:
		return "No such log type."
	}
}
