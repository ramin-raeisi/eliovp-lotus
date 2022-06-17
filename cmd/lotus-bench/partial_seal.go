package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	paramfetch "github.com/filecoin-project/go-paramfetch"
	"github.com/filecoin-project/go-state-types/abi"
	lapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/actors/policy"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper/basicfs"
	"github.com/filecoin-project/lotus/extern/sector-storage/storiface"
	saproof2 "github.com/filecoin-project/specs-actors/v2/actors/runtime/proof"
	"github.com/filecoin-project/specs-storage/storage"
	"github.com/minio/blake2b-simd"
	"github.com/mitchellh/go-homedir"
	"github.com/otiai10/copy"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var partialSealBenchCmd = &cli.Command{
	Name:      "partial-seal",
	Usage:     "Benchmark seal and winning post and window post using prepared PC1 output",
	ArgsUsage: "[input.json]",
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
	},
	Action: func(c *cli.Context) error {
		return partialSealInner(c)
	},
}

func partialSealInner(c *cli.Context) error {
	policy.AddSupportedProofTypes(abi.RegisteredSealProof_StackedDrg2KiBV1)

	if c.Bool("no-gpu") {
		err := os.Setenv("BELLMAN_NO_GPU", "1")
		if err != nil {
			return xerrors.Errorf("setting no-gpu flag: %w", err)
		}
	}

	if !c.Args().Present() {
		return xerrors.Errorf("Usage: lotus-bench partial-seal [input.json]")
	}

	p1OutFile, err := ioutil.ReadFile(c.Args().First())
	if err != nil {
		return xerrors.Errorf("reading input file: %w", err)
	}

	var p1out PrepareOut
	if err := json.Unmarshal(p1OutFile, &p1out); err != nil {
		return xerrors.Errorf("unmarshalling input file: %w", err)
	}

	p2ParNum := c.Int("p2-instances")
	var waitGroup sync.WaitGroup
	waitGroup.Add(p2ParNum)

	sectorBuilderPath := p1out.SectorBuilderPath

	exp, err := homedir.Expand(sectorBuilderPath)
	if err != nil {
		return err
	}
	sbdir := exp

	copiesPath := c.String("copies-path")
	sbdirSplitted := strings.Split(sbdir, "/")
	benchName := sbdirSplitted[len(sbdirSplitted)-1]
	sbdirs := make([]string, p2ParNum)
	for i := 0; i < p2ParNum; i++ {
		if copiesPath != "" {
			sbdirs[i] = copiesPath + benchName + "-" + strconv.Itoa(i)
		} else {
			sbdirs[i] = sbdir + "-" + strconv.Itoa(i)
		}
		err = os.RemoveAll(sbdirs[i]) // if partial-seal was used before
		if err != nil {
			return xerrors.Errorf("Removing old files: %w", err)
		}
		err = copy.Copy(sbdir, sbdirs[i])
		if err != nil {
			return xerrors.Errorf("Making additional copies of data: %w", err)
		}
	}
	defer func() {
		for i := 0; i < len(sbdirs); i++ {
			if err := os.RemoveAll(sbdirs[i]); err != nil {
				log.Warn("remove all: ", err)
			}
		}
	}()

	mid := p1out.Mid
	sectorSize := p1out.SectorSize

	// Only fetch parameters if actually needed
	skipc2 := c.Bool("skip-commit2")
	if !skipc2 {
		if err := paramfetch.GetParams(lcli.ReqContext(c), build.ParametersJSON(), build.SrsJSON(), uint64(sectorSize)); err != nil {
			return xerrors.Errorf("getting params: %w", err)
		}
	}

	sbfs := make([]*basicfs.Provider, p2ParNum)
	sb := make([]*ffiwrapper.Sealer, p2ParNum)
	for i := 0; i < p2ParNum; i++ {
		sbfs[i] = &basicfs.Provider{
			Root: sbdirs[i],
		}
		newSb, err := ffiwrapper.New(sbfs[i])
		if err != nil {
			return err
		}
		sb[i] = newSb
	}

	sectorNumber := p1out.NumSectors

	var sealTimings []SealingResult
	var sealedSectors []saproof2.SectorInfo
	var partSealErr error

	go runPartialSeals(&waitGroup, sb[0], sbfs[0], sectorNumber, c.String("save-commit2-input"), skipc2, c.Bool("skip-unseal"), &p1out, &sealTimings, &sealedSectors, &partSealErr)
	if partSealErr != nil {
		return xerrors.Errorf("failed to run seals: %w", err)
	}

	for i := 1; i < p2ParNum; i++ {
		time.Sleep(time.Second)
		var sealTimings []SealingResult
		var sealedSectors []saproof2.SectorInfo
		var partSealErr error
		go runPartialSeals(&waitGroup, sb[i], sbfs[i], sectorNumber, c.String("save-commit2-input"), skipc2, c.Bool("skip-unseal"), &p1out, &sealTimings, &sealedSectors, &partSealErr)
		if partSealErr != nil {
			return xerrors.Errorf("failed to run seals: %w", err)
		}
	}

	waitGroup.Wait()

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

	beforePost := time.Now()

	if !skipc2 {
		log.Info("generating winning post candidates")
		wipt, err := spt(sectorSize).RegisteredWinningPoStProof()
		if err != nil {
			return err
		}

		fcandidates, err := ffiwrapper.ProofVerifier.GenerateWinningPoStSectorChallenge(context.TODO(), wipt, mid, challenge[:], uint64(len(sealedSectors)))
		if err != nil {
			return err
		}

		candidates := make([]saproof2.SectorInfo, len(fcandidates))
		for i, fcandidate := range fcandidates {
			candidates[i] = sealedSectors[fcandidate]
		}

		gencandidates := time.Now()

		log.Info("computing winning post snark (cold)")
		proof1, err := sb[0].GenerateWinningPoSt(context.TODO(), mid, candidates, challenge[:])
		if err != nil {
			return err
		}

		winningpost1 := time.Now()

		log.Info("computing winning post snark (hot)")
		proof2, err := sb[0].GenerateWinningPoSt(context.TODO(), mid, candidates, challenge[:])
		if err != nil {
			return err
		}

		winnningpost2 := time.Now()

		pvi1 := saproof2.WinningPoStVerifyInfo{
			Randomness:        abi.PoStRandomness(challenge[:]),
			Proofs:            proof1,
			ChallengedSectors: candidates,
			Prover:            mid,
		}
		ok, err := ffiwrapper.ProofVerifier.VerifyWinningPoSt(context.TODO(), pvi1)
		if err != nil {
			return err
		}
		if !ok {
			log.Error("post verification failed")
		}

		verifyWinningPost1 := time.Now()

		pvi2 := saproof2.WinningPoStVerifyInfo{
			Randomness:        abi.PoStRandomness(challenge[:]),
			Proofs:            proof2,
			ChallengedSectors: candidates,
			Prover:            mid,
		}

		ok, err = ffiwrapper.ProofVerifier.VerifyWinningPoSt(context.TODO(), pvi2)
		if err != nil {
			return err
		}
		if !ok {
			log.Error("post verification failed")
		}
		verifyWinningPost2 := time.Now()

		log.Info("computing window post snark (cold)")
		wproof1, _, err := sb[0].GenerateWindowPoSt(context.TODO(), mid, sealedSectors, challenge[:])
		if err != nil {
			return err
		}

		windowpost1 := time.Now()

		log.Info("computing window post snark (hot)")
		wproof2, _, err := sb[0].GenerateWindowPoSt(context.TODO(), mid, sealedSectors, challenge[:])
		if err != nil {
			return err
		}

		windowpost2 := time.Now()

		wpvi1 := saproof2.WindowPoStVerifyInfo{
			Randomness:        challenge[:],
			Proofs:            wproof1,
			ChallengedSectors: sealedSectors,
			Prover:            mid,
		}
		ok, err = ffiwrapper.ProofVerifier.VerifyWindowPoSt(context.TODO(), wpvi1)
		if err != nil {
			return err
		}
		if !ok {
			log.Error("window post verification failed")
		}

		verifyWindowpost1 := time.Now()

		wpvi2 := saproof2.WindowPoStVerifyInfo{
			Randomness:        challenge[:],
			Proofs:            wproof2,
			ChallengedSectors: sealedSectors,
			Prover:            mid,
		}
		ok, err = ffiwrapper.ProofVerifier.VerifyWindowPoSt(context.TODO(), wpvi2)
		if err != nil {
			return err
		}
		if !ok {
			log.Error("window post verification failed")
		}

		verifyWindowpost2 := time.Now()

		bo.PostGenerateCandidates = gencandidates.Sub(beforePost)
		bo.PostWinningProofCold = winningpost1.Sub(gencandidates)
		bo.PostWinningProofHot = winnningpost2.Sub(winningpost1)
		bo.VerifyWinningPostCold = verifyWinningPost1.Sub(winnningpost2)
		bo.VerifyWinningPostHot = verifyWinningPost2.Sub(verifyWinningPost1)

		bo.PostWindowProofCold = windowpost1.Sub(verifyWinningPost2)
		bo.PostWindowProofHot = windowpost2.Sub(windowpost1)
		bo.VerifyWindowPostCold = verifyWindowpost1.Sub(windowpost2)
		bo.VerifyWindowPostHot = verifyWindowpost2.Sub(verifyWindowpost1)
	}

	if c.Bool("json-out") {
		data, err := json.MarshalIndent(bo, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(data))
	} else {
		fmt.Printf("----\nresults (v28) SectorSize:(%d), SectorNumber:(%d)\n", sectorSize, sectorNumber)
		fmt.Printf("seal: preCommit phase 2: %s (%s)\n", bo.SealingSum.PreCommit2, bps(bo.SectorSize, bo.SectorNumber, bo.SealingSum.PreCommit2))
		fmt.Printf("seal: commit phase 1: %s (%s)\n", bo.SealingSum.Commit1, bps(bo.SectorSize, bo.SectorNumber, bo.SealingSum.Commit1))
		fmt.Printf("seal: commit phase 2: %s (%s)\n", bo.SealingSum.Commit2, bps(bo.SectorSize, bo.SectorNumber, bo.SealingSum.Commit2))
		fmt.Printf("seal: verify: %s\n", bo.SealingSum.Verify)
		if !c.Bool("skip-unseal") {
			fmt.Printf("unseal: %s  (%s)\n", bo.SealingSum.Unseal, bps(bo.SectorSize, bo.SectorNumber, bo.SealingSum.Unseal))
		}
		fmt.Println("")
		if !skipc2 {
			fmt.Printf("generate candidates: %s (%s)\n", bo.PostGenerateCandidates, bps(bo.SectorSize, len(bo.SealingResults), bo.PostGenerateCandidates))
			fmt.Printf("compute winning post proof (cold): %s\n", bo.PostWinningProofCold)
			fmt.Printf("compute winning post proof (hot): %s\n", bo.PostWinningProofHot)
			fmt.Printf("verify winning post proof (cold): %s\n", bo.VerifyWinningPostCold)
			fmt.Printf("verify winning post proof (hot): %s\n\n", bo.VerifyWinningPostHot)

			fmt.Printf("compute window post proof (cold): %s\n", bo.PostWindowProofCold)
			fmt.Printf("compute window post proof (hot): %s\n", bo.PostWindowProofHot)
			fmt.Printf("verify window post proof (cold): %s\n", bo.VerifyWindowPostCold)
			fmt.Printf("verify window post proof (hot): %s\n", bo.VerifyWindowPostHot)
		}
	}
	return nil
}

func runPartialSeals(wg *sync.WaitGroup, sb *ffiwrapper.Sealer, sbfs *basicfs.Provider, numSectors int, saveC2inp string, skipc2, skipunseal bool, pc1out *PrepareOut, sealTimings *[]SealingResult, sealedSectors *[]saproof2.SectorInfo, errOut *error) {
	par := pc1out.Par
	mid := pc1out.Mid
	sectorSize := pc1out.SectorSize
	ticketPreimage := pc1out.TicketPreimage
	pieces := pc1out.Pieces

	*sealedSectors = make([]saproof2.SectorInfo, numSectors)
	*sealTimings = make([]SealingResult, numSectors)
	preCommit2Sema := make(chan struct{}, par.PreCommit2)
	commitSema := make(chan struct{}, par.Commit)

	sectorsPerWorker := numSectors / par.PreCommit1

	if numSectors%par.PreCommit1 != 0 {
		sealTimings = nil
		sealedSectors = nil
		*errOut = fmt.Errorf("parallelism factor must cleanly divide numSectors")
		return
	}

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

					trand := blake2b.Sum256(ticketPreimage)
					ticket := abi.SealRandomness(trand[:])

					preCommit2Sema <- struct{}{}
					pc2Start := time.Now()
					log.Infof("[%d] Running replication(2)...", i)
					cids, err := sb.SealPreCommit2(context.TODO(), sid, pc1out.PreCommit1[i])
					if err != nil {
						return xerrors.Errorf("commit: %w", err)
					}

					precommit2 := time.Now()
					<-preCommit2Sema

					(*sealedSectors)[i] = saproof2.SectorInfo{
						SealProof:    sid.ProofType,
						SectorNumber: i,
						SealedCID:    cids.Sealed,
					}

					seed := lapi.SealSeed{
						Epoch: 101,
						Value: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 255},
					}

					commitSema <- struct{}{}
					commitStart := time.Now()
					log.Infof("[%d] Generating PoRep for sector (1)", i)
					piece := []abi.PieceInfo{pieces[i]}
					c1o, err := sb.SealCommit1(context.TODO(), sid, ticket, seed.Value, piece, cids)
					if err != nil {
						return err
					}

					sealcommit1 := time.Now()

					log.Infof("[%d] Generating PoRep for sector (2)", i)

					if saveC2inp != "" {
						c2in := Commit2In{
							SectorNum:  int64(i),
							Phase1Out:  c1o,
							SectorSize: uint64(sectorSize),
						}

						b, err := json.Marshal(&c2in)
						if err != nil {
							return err
						}

						if err := ioutil.WriteFile(saveC2inp, b, 0664); err != nil {
							log.Warnf("%+v", err)
						}
					}

					var proof storage.Proof
					if !skipc2 {
						proof, err = sb.SealCommit2(context.TODO(), sid, c1o)
						if err != nil {
							return err
						}
					}

					sealcommit2 := time.Now()
					<-commitSema

					if !skipc2 {
						svi := saproof2.SealVerifyInfo{
							SectorID:              abi.SectorID{Miner: mid, Number: i},
							SealedCID:             cids.Sealed,
							SealProof:             sid.ProofType,
							Proof:                 proof,
							DealIDs:               nil,
							Randomness:            ticket,
							InteractiveRandomness: seed.Value,
							UnsealedCID:           cids.Unsealed,
						}

						ok, err := ffiwrapper.ProofVerifier.VerifySeal(svi)
						if err != nil {
							return err
						}
						if !ok {
							return xerrors.Errorf("porep proof for sector %d was invalid", i)
						}
					}

					verifySeal := time.Now()

					if !skipunseal {
						log.Infof("[%d] Unsealing sector", i)
						{
							p, done, err := sbfs.AcquireSector(context.TODO(), sid, storiface.FTUnsealed, storiface.FTNone, storiface.PathSealing)
							if err != nil {
								return xerrors.Errorf("acquire unsealed sector for removing: %w", err)
							}
							done()

							if err := os.Remove(p.Unsealed); err != nil {
								return xerrors.Errorf("removing unsealed sector: %w", err)
							}
						}

						err := sb.UnsealPiece(context.TODO(), sid, 0, abi.PaddedPieceSize(sectorSize).Unpadded(), ticket, cids.Unsealed)
						if err != nil {
							return err
						}
					}
					unseal := time.Now()

					(*sealTimings)[i].PreCommit2 = precommit2.Sub(pc2Start)
					(*sealTimings)[i].Commit1 = sealcommit1.Sub(commitStart)
					(*sealTimings)[i].Commit2 = sealcommit2.Sub(sealcommit1)
					(*sealTimings)[i].Verify = verifySeal.Sub(sealcommit2)
					(*sealTimings)[i].Unseal = unseal.Sub(verifySeal)
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
			sealTimings = nil
			sealedSectors = nil
			*errOut = err
			return
		}
	}

	errOut = nil
	wg.Done()
	return
}
