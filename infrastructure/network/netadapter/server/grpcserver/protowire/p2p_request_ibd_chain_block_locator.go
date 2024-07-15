package protowire

import (
	"github.com/harbi-network/harbid/app/appmessage"
	"github.com/harbi-network/harbid/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

func (x *HarbidMessage_RequestIBDChainBlockLocator) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "HarbidMessage_RequestIBDChainBlockLocator is nil")
	}
	return x.RequestIBDChainBlockLocator.toAppMessage()
}

func (x *RequestIBDChainBlockLocatorMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RequestIBDChainBlockLocatorMessage is nil")
	}
	var err error
	var highHash, lowHash *externalapi.DomainHash
	if x.HighHash != nil {
		highHash, err = x.HighHash.toDomain()
		if err != nil {
			return nil, err
		}
	}
	if x.LowHash != nil {
		lowHash, err = x.LowHash.toDomain()
		if err != nil {
			return nil, err
		}
	}
	return &appmessage.MsgRequestIBDChainBlockLocator{
		HighHash: highHash,
		LowHash:  lowHash,
	}, nil

}

func (x *HarbidMessage_RequestIBDChainBlockLocator) fromAppMessage(msgGetBlockLocator *appmessage.MsgRequestIBDChainBlockLocator) error {
	var highHash, lowHash *Hash
	if msgGetBlockLocator.HighHash != nil {
		highHash = domainHashToProto(msgGetBlockLocator.HighHash)
	}
	if msgGetBlockLocator.LowHash != nil {
		lowHash = domainHashToProto(msgGetBlockLocator.LowHash)
	}
	x.RequestIBDChainBlockLocator = &RequestIBDChainBlockLocatorMessage{
		HighHash: highHash,
		LowHash:  lowHash,
	}

	return nil
}
