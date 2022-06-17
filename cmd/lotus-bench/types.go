package main

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/specs-storage/storage"
)

// PrepareOut represents all data what is needed to continue sealing after P1 phase
type PrepareOut struct {
	PreCommit1        []storage.PreCommit1Out
	Pieces            []abi.PieceInfo
	NumSectors        int
	Par               ParCfg
	Mid               abi.ActorID
	SectorSize        abi.SectorSize
	TicketPreimage    []byte
	SectorBuilderPath string
}
