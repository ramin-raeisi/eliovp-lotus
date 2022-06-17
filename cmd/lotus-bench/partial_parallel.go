package main

import (
	"sync"

	"github.com/urfave/cli/v2"
)

var p2C2BenchCmd = &cli.Command{
	Name:      "p2c2",
	Usage:     "Run P2 and C2 phases in parallel for prepared P1 and C1 outputs",
	ArgsUsage: "[p2input.json] [c2input.json]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "no-gpu",
			Usage: "disable gpu usage for the benchmark run",
		},
		&cli.BoolFlag{
			Name:  "json-out",
			Usage: "output results in json format",
		},
		&cli.BoolFlag{
			Name:  "skip-commit2",
			Usage: "skip the commit2 (snark) portion of the benchmark",
		},
		&cli.BoolFlag{
			Name:  "skip-unseal",
			Usage: "skip the unseal portion of the benchmark",
		},
		&cli.StringFlag{
			Name:  "save-commit2-input",
			Usage: "save commit2 input to a file",
		},
		&cli.IntFlag{
			Name:  "p2-instances",
			Usage: "run several intsances p2 in parallel",
			Value: 1,
		},
		&cli.StringFlag{
			Name:  "copies-path",
			Usage: "path for p2-instances copies",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "miner-addr",
			Usage: "pass miner address (only necessary if using existing sectorbuilder)",
			Value: "t01000",
		},
	},
	Action: func(c *cli.Context) error {
		var waitGroup sync.WaitGroup
		waitGroup.Add(2)
		var p2Err error
		var c2Err error
		go runP2(&waitGroup, c, &p2Err)
		go runC2(&waitGroup, c, &c2Err)
		waitGroup.Wait()
		if p2Err != nil {
			return p2Err
		}
		if c2Err != nil {
			return c2Err
		}
		return nil
	},
}

func runC2(wg *sync.WaitGroup, c *cli.Context, err *error) {
	*err = proveInner(c, true)
	wg.Done()
}

func runP2(wg *sync.WaitGroup, c *cli.Context, err *error) {
	*err = partialSealInner(c)
	wg.Done()
}
