package app

import (
	"net/http"

	"github.com/nahue/setlist_manager/templates"
)

// serveWelcome handles GET /
func (app *Application) serveWelcome(w http.ResponseWriter, r *http.Request) {
	component := templates.IndexPage()
	component.Render(r.Context(), w)
}
