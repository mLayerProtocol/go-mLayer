package entities

import (
	"math/big"

	"github.com/mlayerprotocol/go-mlayer/internal/crypto/schnorr"
)

// "math"

type ClaimData struct {
	Version float32 `json:"_v"`
	// Cycle uint64
	// Signature [32]byte
	// Commitment []byte
	// PubKeys []*btcec.PublicKey
	// ApplicationRewardCount []ApplicationCount
	Validator  []byte
	ClaimData  []ApplicationCount
	Cycle      *big.Int
	Index      *big.Int
	Signers    []schnorr.Point
	Commitment []byte
	Signature  []byte
	TotalCost  *big.Int
}

