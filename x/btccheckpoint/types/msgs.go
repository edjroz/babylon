package types

import (
	"errors"
	fmt "fmt"
	"math/big"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// Ensure that MsgInsertBTCSpvProof implements all functions of the Msg interface
	_ sdk.Msg = (*MsgInsertBTCSpvProof)(nil)
)

// Parse and Validate transactions which should contain OP_RETURN data.
// OP_RETURN bytes are not validated in any way. It is up to the caller attach
// semantic meaning and validity to those bytes.
// Returned ParsedProofs are in same order as raw proofs
func ParseTwoProofs(
	submitter sdk.AccAddress,
	proofs []*BTCSpvProof,
	powLimit *big.Int,
	expectedTag txformat.BabylonTag) (*RawCheckpointSubmission, error) {
	// Expecting as many proofs as many parts our checkpoint is composed of
	if len(proofs) != txformat.NumberOfParts {
		return nil, fmt.Errorf("expected at exactly valid op return transactions")
	}

	var parsedProofs []*ParsedProof

	for _, proof := range proofs {
		parsedProof, e :=
			ParseProof(
				proof.BtcTransaction,
				proof.BtcTransactionIndex,
				proof.MerkleNodes,
				proof.ConfirmingBtcHeader,
				powLimit,
			)

		if e != nil {
			return nil, e
		}

		parsedProofs = append(parsedProofs, parsedProof)
	}

	var checkpointData [][]byte

	for i, proof := range parsedProofs {
		data, err := txformat.GetCheckpointData(
			expectedTag,
			txformat.CurrentVersion,
			uint8(i),
			proof.OpReturnData,
		)

		if err != nil {
			return nil, err
		}
		checkpointData = append(checkpointData, data)
	}

	// at this point we know we have two correctly formated babylon op return transacitons
	// we need to check if parts match
	rawCkptData, err := txformat.ConnectParts(txformat.CurrentVersion, checkpointData[0], checkpointData[1])

	if err != nil {
		return nil, err
	}

	rawCheckpoint, err := txformat.DecodeRawCheckpoint(txformat.CurrentVersion, rawCkptData)

	if err != nil {
		return nil, err
	}

	sub := NewRawCheckpointSubmission(submitter, *parsedProofs[0], *parsedProofs[1], *rawCheckpoint)

	return &sub, nil
}

func ParseSubmission(
	m *MsgInsertBTCSpvProof,
	powLimit *big.Int,
	expectedTag txformat.BabylonTag) (*RawCheckpointSubmission, error) {
	if m == nil {
		return nil, errors.New("msgInsertBTCSpvProof can't nil")
	}

	address, err := sdk.AccAddressFromBech32(m.Submitter)

	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid submitter address: %s", err)
	}

	sub, err := ParseTwoProofs(address, m.Proofs, powLimit, expectedTag)

	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (m *MsgInsertBTCSpvProof) ValidateBasic() error {
	// m.Proofs are validated in ante-handler
	_, err := sdk.AccAddressFromBech32(m.Submitter)

	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid submitter address: %s", err)
	}

	return nil
}

func (m *MsgInsertBTCSpvProof) GetSigners() []sdk.AccAddress {
	// cosmos-sdk modules usually ignore possible error here, we panic for the sake
	// of informing something terrible had happend

	submitter, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		// Panic, since the GetSigners method is called after ValidateBasic
		// which performs the same check.
		panic(err)
	}

	return []sdk.AccAddress{submitter}
}
