package main

import (
	"fmt"
)

var help string = `
Strata CLI

strata run <your_dir> [--cli]
strata new app <your_dir> [--module <module-path>]
strata new component <your_dir> [--module <module-path>]
`

func showHelp() {
	fmt.Print(help)
}

func main() {
	args, err := ParseArgs()
	if err != nil {
		fmt.Printf("%s\n\n", err.Error())
		showHelp()
		return
	}
	if args == nil {
		showHelp()
		return
	}
	switch args.command {
	case "run":
		RunApp(args)
	case "new":
		NewProject(args)
	default:
		showHelp()
	}
}
