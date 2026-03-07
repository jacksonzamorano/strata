package main

import (
	"errors"
	"os"
	"slices"
	"strings"
)

type AppOptions string

const (
	AppOptionHostCli AppOptions = "cli"
)

type AppArgs struct {
	directory  string
	command    string
	subcommand string
	modulePath string
	opts       []AppOptions
}

func (a *AppArgs) Specifies(opt AppOptions) bool {
	return slices.Contains(a.opts, opt)
}

func ParseArgs() (*AppArgs, error) {
	return ParseArgList(os.Args[1:])
}

func ParseArgList(input []string) (*AppArgs, error) {
	if len(input) == 0 {
		return nil, nil
	}

	args := &AppArgs{
		command: input[0],
	}

	i := 1
	for i < len(input) {
		token := input[i]
		if trim, ok := strings.CutPrefix(token, "--"); ok {
			switch trim {
			case string(AppOptionHostCli):
				args.opts = append(args.opts, AppOptionHostCli)
			case "module":
				if i+1 >= len(input) {
					return nil, errors.New("missing value for --module")
				}
				args.modulePath = input[i+1]
				i++
			default:
				return nil, errors.New("unknown option --" + trim)
			}
		} else if args.command == "new" && len(args.subcommand) == 0 {
			args.subcommand = token
		} else if len(args.directory) == 0 {
			args.directory = token
		} else {
			return nil, errors.New("too many arguments")
		}
		i++
	}

	switch args.command {
	case "run":
		if len(args.directory) == 0 {
			return nil, errors.New("run requires a target directory")
		}
	case "new":
		if len(args.subcommand) == 0 {
			return nil, errors.New("new requires a project type")
		}
		if args.subcommand != "app" && args.subcommand != "component" {
			return nil, errors.New("new only supports app or component")
		}
		if len(args.directory) == 0 {
			return nil, errors.New("new requires a target directory")
		}
		if len(args.opts) > 0 {
			return nil, errors.New("--cli is only supported with run")
		}
	default:
		return nil, errors.New("unknown command")
	}

	return args, nil
}
