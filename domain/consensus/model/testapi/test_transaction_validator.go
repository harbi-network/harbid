package testapi

import (
	"github.com/harbi-network/harbid/domain/consensus/model"
	"github.com/harbi-network/harbid/domain/consensus/utils/txscript"
)

// TestTransactionValidator adds to the main TransactionValidator methods required by tests
type TestTransactionValidator interface {
	model.TransactionValidator
	SigCache() *txscript.SigCache
	SetSigCache(sigCache *txscript.SigCache)
}
