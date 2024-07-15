package transactionid

import (
	"github.com/harbi-network/harbid/domain/consensus/model/externalapi"
)

// FromString creates a new DomainTransactionID from the given string
func FromString(str string) (*externalapi.DomainTransactionID, error) {
	hash, err := externalapi.NewDomainHashFromString(str)
	return (*externalapi.DomainTransactionID)(hash), err
}
