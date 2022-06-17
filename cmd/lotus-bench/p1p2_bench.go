package main

import (
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

var p1p2BenchCmd = &cli.Command{
	Name:      "p1p2",
	Usage:     "Run P1 and P2 phases in parallel for a prepared P1 output",
	ArgsUsage: "[p2input.json]",
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
			Name:  "storage-dir",
			Value: "~/.lotus-bench",
			Usage: "path to the storage directory that will store sectors long term",
		},
		&cli.StringFlag{
			Name:  "sector-size",
			Value: "512MiB",
			Usage: "size of the sectors in bytes, i.e. 32GiB",
		},
		&cli.StringFlag{
			Name:  "miner-addr",
			Usage: "pass miner address (only necessary if using existing sectorbuilder)",
			Value: "t01000",
		},
		&cli.StringFlag{
			Name:  "ticket-preimage",
			Usage: "ticket random",
		},
		&cli.IntFlag{
			Name:  "num-sectors",
			Usage: "select number of sectors to seal",
			Value: 1,
		},
		&cli.IntFlag{
			Name:  "parallel",
			Usage: "num run in parallel",
			Value: 1,
		},
		&cli.IntFlag{
			Name:  "p1-delay",
			Usage: "delay (in seconds) before p1 start",
			Value: 0,
		},
		&cli.IntFlag{
			Name:  "p2-delay",
			Usage: "delay (in seconds) before p2 start",
			Value: 0,
		},
	},
	Action: func(c *cli.Context) error {
		var waitGroup sync.WaitGroup
		waitGroup.Add(2)
		var p2Err error
		var p1Err error
		var p1Delay = c.Int("p1-delay")
		var p2Delay = c.Int("p2-delay")
		if p1Delay >= p2Delay {
			time.Sleep(time.Duration(p2Delay) * time.Second)
			go runP2(&waitGroup, c, &p2Err)
			time.Sleep(time.Duration(p1Delay-p2Delay) * time.Second)
			go p1Clean(&waitGroup, c, &p1Err)
		} else {
			time.Sleep(time.Duration(p1Delay) * time.Second)
			go p1Clean(&waitGroup, c, &p1Err)
			time.Sleep(time.Duration(p2Delay-p1Delay) * time.Second)
			go runP2(&waitGroup, c, &p2Err)
		}
		waitGroup.Wait()
		if p2Err != nil {
			return p2Err
		}
		if p1Err != nil {
			return p1Err
		}
		return nil
	},
}

func p1Clean(wg *sync.WaitGroup, c *cli.Context, err *error) {
	*err = p1Inner(c, false)
	wg.Done()
}
