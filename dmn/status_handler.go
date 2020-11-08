package dmn

import "net/http"

// HandleStatus handles the status command
func (a *App) HandleStatus(w http.ResponseWriter, r *http.Request) {

}

// StatusCmd indicates if the app is up or down
func (a *App) StatusCmd() (bool, error) {

	return true, nil
}
