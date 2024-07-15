package main

import (
	"github.com/harbi-network/harbid/infrastructure/logger"
	"github.com/harbi-network/harbid/util/panics"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("RORG")
	spawn      = panics.GoroutineWrapperFunc(log)
)
