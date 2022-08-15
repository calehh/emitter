package emitter

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func EventSignatureHash(funSignature string) ethcommon.Hash {
	return crypto.Keccak256Hash([]byte(funSignature))
}

