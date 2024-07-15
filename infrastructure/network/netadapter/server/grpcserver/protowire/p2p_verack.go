package protowire

import (
	"github.com/harbi-network/harbid/app/appmessage"
	"github.com/pkg/errors"
)

func (x *HarbidMessage_Verack) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "HarbidMessage_Verack is nil")
	}
	return &appmessage.MsgVerAck{}, nil
}

func (x *HarbidMessage_Verack) fromAppMessage(_ *appmessage.MsgVerAck) error {
	return nil
}
