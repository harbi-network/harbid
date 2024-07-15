package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/harbi-network/harbid/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var commandTypes = []reflect.Type{
	reflect.TypeOf(protowire.HarbidMessage_AddPeerRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetConnectedPeerInfoRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetPeerAddressesRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetCurrentNetworkRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetInfoRequest{}),

	reflect.TypeOf(protowire.HarbidMessage_GetBlockRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetBlocksRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetHeadersRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetBlockCountRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetBlockDagInfoRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetSelectedTipHashRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetVirtualSelectedParentBlueScoreRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetVirtualSelectedParentChainFromBlockRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_ResolveFinalityConflictRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_EstimateNetworkHashesPerSecondRequest{}),

	reflect.TypeOf(protowire.HarbidMessage_GetBlockTemplateRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_SubmitBlockRequest{}),

	reflect.TypeOf(protowire.HarbidMessage_GetMempoolEntryRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetMempoolEntriesRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetMempoolEntriesByAddressesRequest{}),

	reflect.TypeOf(protowire.HarbidMessage_SubmitTransactionRequest{}),

	reflect.TypeOf(protowire.HarbidMessage_GetUtxosByAddressesRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetBalanceByAddressRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_GetCoinSupplyRequest{}),

	reflect.TypeOf(protowire.HarbidMessage_BanRequest{}),
	reflect.TypeOf(protowire.HarbidMessage_UnbanRequest{}),
}

type commandDescription struct {
	name       string
	parameters []*parameterDescription
	typeof     reflect.Type
}

type parameterDescription struct {
	name   string
	typeof reflect.Type
}

func commandDescriptions() []*commandDescription {
	commandDescriptions := make([]*commandDescription, len(commandTypes))

	for i, commandTypeWrapped := range commandTypes {
		commandType := unwrapCommandType(commandTypeWrapped)

		name := strings.TrimSuffix(commandType.Name(), "RequestMessage")
		numFields := commandType.NumField()

		var parameters []*parameterDescription
		for i := 0; i < numFields; i++ {
			field := commandType.Field(i)

			if !isFieldExported(field) {
				continue
			}

			parameters = append(parameters, &parameterDescription{
				name:   field.Name,
				typeof: field.Type,
			})
		}
		commandDescriptions[i] = &commandDescription{
			name:       name,
			parameters: parameters,
			typeof:     commandTypeWrapped,
		}
	}

	return commandDescriptions
}

func (cd *commandDescription) help() string {
	sb := &strings.Builder{}
	sb.WriteString(cd.name)
	for _, parameter := range cd.parameters {
		_, _ = fmt.Fprintf(sb, " [%s]", parameter.name)
	}
	return sb.String()
}
