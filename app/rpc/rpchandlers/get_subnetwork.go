package rpchandlers

import (
	"github.com/harbi-network/harbid/app/appmessage"
	"github.com/harbi-network/harbid/app/rpc/rpccontext"
	"github.com/harbi-network/harbid/infrastructure/network/netadapter/router"
)

// HandleGetSubnetwork handles the respectively named RPC command
func HandleGetSubnetwork(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	response := &appmessage.GetSubnetworkResponseMessage{}
	response.Error = appmessage.RPCErrorf("not implemented")
	return response, nil
}
