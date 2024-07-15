package protowire

import (
	"github.com/harbi-network/harbid/app/appmessage"
	"github.com/pkg/errors"
)

func (x *HarbidMessage_Ready) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "HarbidMessage_Ready is nil")
	}
	return &appmessage.MsgReady{}, nil
}

func (x *HarbidMessage_Ready) fromAppMessage(_ *appmessage.MsgReady) error {
	return nil
}
