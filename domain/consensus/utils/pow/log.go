package pow

import (
	"github.com/harbi-network/harbid/infrastructure/logger"
	"github.com/harbi-network/harbid/util/panics"
)

var log = logger.RegisterSubSystem("POW")
var spawn = panics.GoroutineWrapperFunc(log)
