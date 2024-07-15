package appmessage

import (
	"github.com/harbi-network/harbid/domain/consensus/model/externalapi"
)

// MsgRequestIBDBlocks implements the Message interface and represents a harbi
// RequestIBDBlocks message. It is used to request blocks as part of the IBD
// protocol.
type MsgRequestIBDBlocks struct {
	baseMessage
	Hashes []*externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestIBDBlocks) Command() MessageCommand {
	return CmdRequestIBDBlocks
}

// NewMsgRequestIBDBlocks returns a new MsgRequestIBDBlocks.
func NewMsgRequestIBDBlocks(hashes []*externalapi.DomainHash) *MsgRequestIBDBlocks {
	return &MsgRequestIBDBlocks{
		Hashes: hashes,
	}
}
