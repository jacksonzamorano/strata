package main

import (
	"fmt"
)

var help string = `
Strata CLI

stratacli run [your_dir]
`

func showHelp() {
	fmt.Print(help)
}



func main() {
	args := ParseArgs()
	if args == nil {
		showHelp()
		return
	}
	switch args.command {
	case "run":
		RunApp(args)
	default:
		showHelp()
	}

}
