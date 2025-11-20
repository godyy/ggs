package agent

import (
	"github.com/godyy/ggs/app/agent/internal/log"
	"github.com/godyy/ggs/internal/libs/logger"
	"go.uber.org/zap"
)

var (
	loggerInst        logger.Logger
	loggerInsideAgent logger.Logger
)

func init() {
	logger.RegisterAfterInitFunc(func(logger logger.Logger) {
		loggerInst = logger.Named("agent")
		loggerInsideAgent = logger.AddCallerSkip(1)
	})
}

func (a *Agent) getLogFields(fields ...zap.Field) []zap.Field {
	var baseFields []zap.Field
	if a.isConnected() {
		baseFields = []zap.Field{log.FldPlayerId(a.playerId), log.FldSessionId(a.sessionId)}
	} else {
		baseFields = []zap.Field{log.FldRemoteAddr(a.conn.RemoteAddr())}
	}
	return append(baseFields, fields...)
}

func (a *Agent) DebugFields(f string, fields ...zap.Field) {
	loggerInsideAgent.DebugFields(f, a.getLogFields(fields...)...)
}

func (a *Agent) InfoFields(f string, fields ...zap.Field) {
	loggerInsideAgent.InfoFields(f, a.getLogFields(fields...)...)
}

func (a *Agent) errorFields(f string, fields ...zap.Field) {
	loggerInsideAgent.ErrorFields(f, a.getLogFields(fields...)...)
}
