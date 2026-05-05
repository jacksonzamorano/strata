package main

import "testing"

func TestParseArgListRun(t *testing.T) {
	args, err := ParseArgList([]string{"run", "./demo"})
	if err != nil {
		t.Fatalf("ParseArgList returned error: %v", err)
	}
	if args.command != "run" {
		t.Fatalf("command = %q, want run", args.command)
	}
	if args.target != "./demo" {
		t.Fatalf("directory = %q, want ./demo", args.target)
	}
}

func TestParseArgListNewComponent(t *testing.T) {
	args, err := ParseArgList([]string{"new", "component", "demo-component", "--module", "example.com/acme/component"})
	if err != nil {
		t.Fatalf("ParseArgList returned error: %v", err)
	}
	if args.command != "new" {
		t.Fatalf("command = %q, want new", args.command)
	}
	if args.subcommand != "component" {
		t.Fatalf("subcommand = %q, want component", args.subcommand)
	}
	if args.modulePath != "example.com/acme/component" {
		t.Fatalf("modulePath = %q, want example.com/acme/component", args.modulePath)
	}
	if !args.Specifies(AppOptionFlagModule) {
		t.Fatalf("expected --module to be parsed")
	}
}

func TestParseArgListRejectsIncompleteNew(t *testing.T) {
	_, err := ParseArgList([]string{"new", "app"})
	if err == nil {
		t.Fatalf("expected error for incomplete new command")
	}
}

func TestParseArgListRejectsMissingModuleValue(t *testing.T) {
	_, err := ParseArgList([]string{"new", "component", "demo-component", "--module"})
	if err == nil {
		t.Fatalf("expected error for missing --module value")
	}
}
