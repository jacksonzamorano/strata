package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
)

func AddDependancy(args *AppArgs) {
	ctx := context.Background()
	_, err := goCommandRunner(ctx, args.cwd, "get", args.target)
	if err != nil {
		fmt.Printf("Could not add the module: %s", err.Error())
		return
	}

	depFile, err := os.OpenFile("components.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Could not write components file: %s", err.Error())
		return
	}
	defer depFile.Close()

	depFile.WriteString(args.target)
	depFile.WriteString("\n")

	name := path.Base(args.target)
	name = strings.TrimPrefix(name, "strata-")

	fmt.Printf("Added component!\n\nTo start using, import it's definitions:\nimport %s \"%s/definitions\"\n", name, args.target)
}
