package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func NewProject(args *AppArgs) {
	kind := ProjectKind(args.subcommand)
	result, err := GenerateProject(context.Background(), GenerateProjectOptions{
		Kind:       kind,
		Directory:  args.directory,
		ModulePath: args.modulePath,
	})
	if err != nil {
		var bootstrapErr *BootstrapError
		if errors.As(err, &bootstrapErr) && result != nil {
			fmt.Printf(
				"Created %s project in %s with module %s.\n`go mod tidy` failed:\n%s\n\nNext steps:\n%s\n",
				kind,
				result.Directory,
				result.ModulePath,
				bootstrapErr.Detail(),
				nextSteps(*result),
			)
			return
		}

		fmt.Printf("Could not create %s project: %s\n", kind, err.Error())
		return
	}

	fmt.Printf(
		"Created %s project in %s with module %s.\n\nNext steps:\n%s\n",
		kind,
		result.Directory,
		result.ModulePath,
		nextSteps(*result),
	)
}

func nextSteps(result GenerateProjectResult) string {
	lines := []string{
		fmt.Sprintf("  cd %s", result.Directory),
	}

	if result.Kind == ProjectKindApp {
		lines = append(lines, "  strata run . --cli")
	} else {
		lines = append(lines, "  go build .")
	}

	return strings.Join(lines, "\n")
}
