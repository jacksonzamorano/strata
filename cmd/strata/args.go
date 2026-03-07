package main

import (
	"os"
	"slices"
	"strings"
)

type AppOptions string

const (
	AppOptionHostCli AppOptions = "cli"
)

type AppArgs struct {
	directory string
	command   string
	opts      []AppOptions
}

func (a *AppArgs) Specifies(opt AppOptions) bool {
	return slices.Contains(a.opts, opt)
}

func ParseArgs() *AppArgs {
	var args AppArgs

	osArgs := os.Args
	if len(osArgs) < 2 {
		return nil
	}

	args.command = osArgs[1]
	i := 2
	for i < len(osArgs) {
		if trim, ok := strings.CutPrefix(osArgs[i], "--"); ok {
			args.opts = append(args.opts, AppOptions(trim))
		} else if len(args.directory) == 0 {
			args.directory = osArgs[i]
		}
		i++
	}

	return &args
}
