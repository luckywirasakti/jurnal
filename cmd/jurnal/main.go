package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/luckywirasakti/jurnal/internal/cli"
)

const version = "1.0.0"

func main() {
	var versionFlag bool
	flag.BoolVar(&versionFlag, "v", false, "Show version and credits")
	flag.BoolVar(&versionFlag, "version", false, "Show version and credits")
	flag.Parse()

	if versionFlag {
		cli.PrintCredits(version)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		cli.PrintUsage()
		return
	}

	switch args[0] {
	case "init":
		cli.HandleInit(args[1:])
	case "setup":
		cli.HandleSetup(args[1:])
	case "stage":
		cli.HandleStage()
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		os.Exit(1)
	}
}
