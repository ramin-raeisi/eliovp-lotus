package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/actors/policy"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper/basicfs"
	"github.com/filecoin-project/specs-storage/storage"
	"github.com/minio/blake2b-simd"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var generateLabelsBenchCmd = &cli.Command{
	Name:      "labels-bench",
	Usage:     "",
	ArgsUsage: "",
	Flags: []cli.Flag{
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
		&cli.BoolFlag{
			Name:  "no-gpu",
			Usage: "disable gpu usage for the benchmark run",
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
	},
	Action: func(c *cli.Context) error {
		policy.AddSupportedProofTypes(abi.RegisteredSealProof_StackedDrg2KiBV1)

		if c.Bool("no-gpu") {
			err := os.Setenv("BELLMAN_NO_GPU", "1")
			if err != nil {
				return xerrors.Errorf("setting no-gpu flag: %w", err)
			}
		}

		var sbdir string

		sdir, err := homedir.Expand(c.String("storage-dir"))
		if err != nil {
			return err
		}

		err = os.MkdirAll(sdir, 0775) //nolint:gosec
		if err != nil {
			return xerrors.Errorf("creating sectorbuilder dir: %w", err)
		}

		tsdir, err := ioutil.TempDir(sdir, "bench")
		if err != nil {
			return err
		}
		defer func() {
			if err := os.RemoveAll(tsdir); err != nil {
				log.Warn("remove all: ", err)
			}
		}()

		// TODO: pretty sure this isnt even needed?
		if err := os.MkdirAll(tsdir, 0775); err != nil {
			return err
		}

		sbdir = tsdir

		// miner address
		maddr, err := address.NewFromString(c.String("miner-addr"))
		if err != nil {
			return err
		}
		amid, err := address.IDFromAddress(maddr)
		if err != nil {
			return err
		}
		mid := abi.ActorID(amid)

		// sector size
		sectorSizeInt, err := units.RAMInBytes(c.String("sector-size"))
		if err != nil {
			return err
		}
		sectorSize := abi.SectorSize(sectorSizeInt)

		sbfs := &basicfs.Provider{
			Root: sbdir,
		}

		sb, err := ffiwrapper.New(sbfs)
		if err != nil {
			return err
		}

		sectorNumber := c.Int("num-sectors")

		var sealTimings []SealingResult

		parCfg := ParCfg{
			PreCommit1: c.Int("parallel"),
			PreCommit2: 1,
			Commit:     1,
		}
		sealTimings, _, err = runGenerateLabelsBench(sb, sbfs, sectorNumber, parCfg, mid, sectorSize, []byte(c.String("ticket-preimage")))
		if err != nil {
			return xerrors.Errorf("failed to run P1: %w", err)
		}

		bo := BenchResults{
			SectorSize:     sectorSize,
			SectorNumber:   sectorNumber,
			SealingResults: sealTimings,
		}
		if err := bo.SumSealingTime(); err != nil {
			return err
		}

		var challenge [32]byte
		rand.Read(challenge[:])

		if c.Bool("json-out") {
			data, err := json.MarshalIndent(bo, "", "  ")
			if err != nil {
				return err
			}

			fmt.Println(string(data))
		} else {
			fmt.Printf("----\nresults (v28) SectorSize:(%d), SectorNumber:(%d)\n", sectorSize, sectorNumber)
			fmt.Printf("seal: addPiece: %s (%s)\n", bo.SealingSum.AddPiece, bps(bo.SectorSize, bo.SectorNumber, bo.SealingSum.AddPiece))
			fmt.Printf("seal: preCommit phase 1: %s (%s)\n", bo.SealingSum.PreCommit1, bps(bo.SectorSize, bo.SectorNumber, bo.SealingSum.PreCommit1))
			fmt.Println("")
		}
		return nil
	},
}

func runGenerateLabelsBench(sb *ffiwrapper.Sealer, sbfs *basicfs.Provider, numSectors int, par ParCfg, mid abi.ActorID, sectorSize abi.SectorSize, ticketPreimage []byte) ([]SealingResult, *PrepareOut, error) {
	var pieces []abi.PieceInfo
	sealTimings := make([]SealingResult, numSectors)
	var res PrepareOut = PrepareOut{}
	res.PreCommit1 = make([]storage.PreCommit1Out, numSectors)
	res.NumSectors = numSectors
	res.Par = par
	res.Mid = mid
	res.SectorSize = sectorSize
	res.TicketPreimage = ticketPreimage

	if numSectors%par.PreCommit1 != 0 {
		return nil, &res, fmt.Errorf("parallelism factor must cleanly divide numSectors")
	}

	for i := abi.SectorNumber(0); i < abi.SectorNumber(numSectors); i++ {
		sid := storage.SectorRef{
			ID: abi.SectorID{
				Miner:  mid,
				Number: i,
			},
			ProofType: spt(sectorSize),
		}

		start := time.Now()
		log.Infof("[%d] Writing piece into sector...", i)

		r := rand.New(rand.NewSource(100 + int64(i)))

		pi, err := sb.AddPiece(context.TODO(), sid, nil, abi.PaddedPieceSize(sectorSize).Unpadded(), r)
		if err != nil {
			return nil, &res, err
		}

		pieces = append(pieces, pi)

		sealTimings[i].AddPiece = time.Since(start)
	}

	res.Pieces = pieces
	sectorsPerWorker := numSectors / par.PreCommit1

	errs := make(chan error, par.PreCommit1)
	for wid := 0; wid < par.PreCommit1; wid++ {
		go func(worker int) {
			sealerr := func() error {
				start := worker * sectorsPerWorker
				end := start + sectorsPerWorker
				for i := abi.SectorNumber(start); i < abi.SectorNumber(end); i++ {
					sid := storage.SectorRef{
						ID: abi.SectorID{
							Miner:  mid,
							Number: i,
						},
						ProofType: spt(sectorSize),
					}

					start := time.Now()

					trand := blake2b.Sum256(ticketPreimage)
					ticket := abi.SealRandomness(trand[:])

					log.Infof("[%d] Running replication(1)...", i)
					piece := []abi.PieceInfo{pieces[i]}
					pc1o, err := sb.SealGenerateLabelsBench(context.TODO(), sid, ticket, piece)
					if err != nil {
						return xerrors.Errorf("commit: %w", err)
					}

					precommit1 := time.Now()

					res.PreCommit1[i] = pc1o

					sealTimings[i].PreCommit1 = precommit1.Sub(start)
				}
				return nil
			}()
			if sealerr != nil {
				errs <- sealerr
				return
			}
			errs <- nil
		}(wid)
	}

	for i := 0; i < par.PreCommit1; i++ {
		err := <-errs
		if err != nil {
			return nil, &res, err
		}
	}

	return sealTimings, &res, nil
}
