package tendermint

import (
	"bytes"
	"errors"

	lerr "github.com/tendermint/tendermint/lite/errors"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

var _ exported.ConsensusState = ConsensusState{}

// ConsensusState defines a Tendermint consensus state
type ConsensusState struct {
	ChainID          string
	Height           uint64
	Root             ics23.Root
	NextValidatorSet *tmtypes.ValidatorSet
}

func (ConsensusState) Kind() exported.Kind {
	return exported.Tendermint
}

func (cs ConsensusState) GetHeight() uint64 {
	return cs.Height
}

func (cs ConsensusState) GetRoot() ics23.Root {
	return cs.Root
}

func (cs ConsensusState) update(header Header) ConsensusState {
	return ConsensusState{
		ChainID:          cs.ChainID,
		Height:           uint64(header.Height),
		Root:             merkle.NewRoot(header.AppHash),
		NextValidatorSet: header.NextValidatorSet,
	}
}

// CheckValidityAndUpdateState
func (cs ConsensusState) CheckValidityAndUpdateState(header exported.Header) (exported.ConsensusState, error) {
	tmHeader, ok := header.(Header)
	if !ok {
		return nil, errors.New("header is not from a tendermint consensus") // TODO: create concrete error
	}

	if cs.Height == uint64(tmHeader.Height-1) {
		nexthash := cs.NextValidatorSet.Hash()
		if !bytes.Equal(tmHeader.ValidatorsHash, nexthash) {
			return nil, lerr.ErrUnexpectedValidators(tmHeader.ValidatorsHash, nexthash)
		}
	}

	if !bytes.Equal(tmHeader.NextValidatorsHash, tmHeader.NextValidatorSet.Hash()) {
		return nil, lerr.ErrUnexpectedValidators(tmHeader.NextValidatorsHash, tmHeader.NextValidatorSet.Hash())
	}

	err := tmHeader.ValidateBasic(cs.ChainID)
	if err != nil {
		return nil, err
	}

	err = cs.NextValidatorSet.VerifyFutureCommit(tmHeader.ValidatorSet, cs.ChainID, tmHeader.Commit.BlockID, tmHeader.Height, tmHeader.Commit)
	if err != nil {
		return nil, err
	}

	return cs.update(tmHeader), nil
}

// CheckMisbehaviourAndUpdateState - not implemented
func (cs ConsensusState) CheckMisbehaviourAndUpdateState(mb exported.Misbehaviour) bool {
	// TODO: implement
	return false
}

var _ exported.Header = Header{}

// Header defines the Tendermint consensus Header
type Header struct {
	// TODO: define Tendermint header type manually, don't use tmtypes
	tmtypes.SignedHeader
	ValidatorSet     *tmtypes.ValidatorSet
	NextValidatorSet *tmtypes.ValidatorSet
}

// Kind defines that the Header is a Tendermint consensus algorithm
func (header Header) Kind() exported.Kind {
	return exported.Tendermint
}

// GetHeight returns the current height
func (header Header) GetHeight() uint64 {
	return uint64(header.Height)
}