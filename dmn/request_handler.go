package dmn

// RequestHandler handles requests and needs to have access to a secret and history
type RequestHandler struct {
	CommandScheduler Scheduler
}
