package templatemanager

import (
	"sync"

	"github.com/harbi-network/harbid/app/appmessage"
	"github.com/harbi-network/harbid/domain/consensus/model/externalapi"
	"github.com/harbi-network/harbid/domain/consensus/utils/pow"
	"github.com/harbi-network/harbid/infrastructure/logger"
)

var currentTemplate *externalapi.DomainBlock
var currentState *pow.State
var isSynced bool
var lock = &sync.Mutex{}

// Get returns the template to work on
func Get() (*externalapi.DomainBlock, *pow.State, bool) {
	lock.Lock()
	defer lock.Unlock()
	// Shallow copy the block so when the user replaces the header it won't affect the template here.
	if currentTemplate == nil {
		return nil, nil, false
	}
	block := *currentTemplate
	state := *currentState
	return &block, &state, isSynced
}

// Set sets the current template to work on
func Set(template *appmessage.GetBlockTemplateResponseMessage, backendLog *logger.Backend) error {
	block, err := appmessage.RPCBlockToDomainBlock(template.Block)
	if err != nil {
		return err
	}
	lock.Lock()
	defer lock.Unlock()
	currentTemplate = block
	pow.SetLogger(backendLog, logger.LevelTrace)
	currentState = pow.NewState(block.Header.ToMutable(), true)
	isSynced = template.IsSynced
	return nil
}
