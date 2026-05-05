package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

type AppOptions string

const (
	AppOptionFlagModule AppOptions = "module"
	AppOptionRunMode    AppOptions = "run"
	AppOptionNewMode    AppOptions = "new"
	AppOptionAddMode    AppOptions = "add"
)

type AppArgs struct {
	target     string
	command    string
	subcommand string

	modulePath string

	opts []AppOptions
	cwd  string
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

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	args := &AppArgs{
		command: input[0],
		cwd:     cwd,
	}

	i := 1
	for i < len(input) {
		token := input[i]
		if trim, ok := strings.CutPrefix(token, "--"); ok {
			switch AppOptions(trim) {
			case AppOptionFlagModule:
				if i+1 >= len(input) {
					return nil, errors.New("missing value for --module")
				}
				args.modulePath = input[i+1]
				args.opts = append(args.opts, AppOptionFlagModule)
			default:
				return nil, fmt.Errorf("Unknown flag '%s'", trim)
			}
			i++
		} else if args.command == "new" && len(args.subcommand) == 0 {
			args.subcommand = token
		} else if len(args.target) == 0 {
			args.target = token
		} else {
			return nil, errors.New("too many arguments")
		}
		i++
	}

	switch AppOptions(args.command) {
	case AppOptionRunMode:
		if len(args.target) == 0 {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, errors.New("Could not get current directory.")
			}
			args.target = cwd
		}
	case AppOptionNewMode:
		if len(args.subcommand) == 0 {
			return nil, errors.New("new requires a project type")
		}
		if args.subcommand != "app" && args.subcommand != "component" {
			return nil, errors.New("new only supports app or component")
		}
		if len(args.target) == 0 {
			return nil, errors.New("new requires a target directory")
		}
	case AppOptionAddMode:
		if len(args.target) == 0 {
			return nil, errors.New("Add requires a module to import.")
		}
	default:
		return nil, errors.New("unknown command")
	}

	return args, nil
}
