package consensus

import (
	"github.com/harbi-network/harbid/infrastructure/logger"
	"github.com/harbi-network/harbid/util/panics"
)

var log = logger.RegisterSubSystem("BDAG")
var spawn = panics.GoroutineWrapperFunc(log)
