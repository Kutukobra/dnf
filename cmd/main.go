package main

import (
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v2"

	"github.com/fre5gc/dnf/internal/logger"
	"github.com/free5gc/util/version"
)

func main() {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.MainLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	app := cli.NewApp()
	app.Name = "dnf"
	app.Usage = "5G Dummy Network Function (DNF)"
	app.Action = action
	app.Flags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "log",
			Aliases: []string{"l"},
			Usage:   "Output NF log to `FILE`",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.MainLog.Errorf("DNF Run error: %v\n", err)
	}
}

func action(cliCtx *cli.Context) error {
	logger.MainLog.Infoln("DNF version: ", version.GetVersion())
	return nil
}
