package blockchain

import (
	"github.com/tin-incognito/simulate-consensus/utils"
	"time"
)

//Transaction ...
type Transaction struct{
	Index int64 `json:"index"`
	Hash string `json:"hash"`
	Type string `json:"type"`
	LockTime string `json:"lock_time"`
	Version int8 `json:"version"`
	Fee uint64 `json:"fee"`
	Metadata []string `json:"metadata"`
	Timestamp uint64
}

func GenerateTxs() map[string]Transaction{

	m := make(map[string]Transaction)

	for i := 0 ; i < 5; i++{
		tx := Transaction{
			Index:     int64(i),
			Hash:      utils.GenerateHashV1(),
			Type:      "normal tx",
			LockTime:  "",
			Version:   1,
			Fee:       100,
			Metadata:  nil,
			Timestamp: uint64(time.Now().Unix()),
		}
		m[tx.Hash] = tx
	}

	return m
}
