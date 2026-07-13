package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/luckywirasakti/jurnal/internal/cli"
)

const version = "1.0.0"

func main() {
	versionFlag := flag.Bool("v", false, "Show version and credits")
	versionLongFlag := flag.Bool("version", false, "Show version and credits")
	flag.Parse()

	if *versionFlag || *versionLongFlag {
		cli.PrintCredits(version)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		cli.PrintUsage()
		return
	}

	switch args[0] {
	case "setup":
		cli.HandleSetup()
	case "stage":
		cli.HandleStage()
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		os.Exit(1)
	}
}
