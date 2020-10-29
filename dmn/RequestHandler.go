package dmn

import "log"

// RequestHandler handles requests and needs to have access to a secret and history
type RequestHandler struct {
	Secret           Secret
	History          HistoryFile
	CommandScheduler Scheduler
	Log              *log.Logger
}

// Set sets global some variables
func (handler *RequestHandler) Set(secret Secret, history HistoryFile) {
	handler.Secret = secret
	handler.History = history
}
