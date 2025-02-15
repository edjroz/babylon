package datagen

import (
	"time"

	extendedkeeper "github.com/babylonchain/babylon/x/zoneconcierge/extended-client-keeper"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func GenRandomTMHeader(chainID string, height uint64) *tmproto.Header {
	return &tmproto.Header{
		ChainID:        chainID,
		Height:         int64(height),
		Time:           time.Now(),
		LastCommitHash: GenRandomByteArray(32),
	}
}

func GenRandomIBCTMHeader(chainID string, height uint64) *ibctmtypes.Header {
	return &ibctmtypes.Header{
		SignedHeader: &tmproto.SignedHeader{
			Header: &tmproto.Header{
				ChainID:        chainID,
				Height:         int64(height),
				LastCommitHash: GenRandomByteArray(32),
			},
		},
	}
}

func HeaderToHeaderInfo(header *ibctmtypes.Header) *extendedkeeper.HeaderInfo {
	return &extendedkeeper.HeaderInfo{
		Hash:     header.Header.LastCommitHash,
		ChaindId: header.Header.ChainID,
		Height:   uint64(header.Header.Height),
	}
}
