package rpcclient

import (
	"github.com/harbi-network/harbid/infrastructure/logger"
	"github.com/harbi-network/harbid/util/panics"
)

var log = logger.RegisterSubSystem("RPCC")
var spawn = panics.GoroutineWrapperFunc(log)
