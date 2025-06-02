package main

import (
	"net/http"
)

// HealthCheckHandler godoc
//
//	@Summary		Health check
//	@Description	Health check endpoint
//	@Tags			ops
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		500	{object}	error
//	@Router			/health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
	}

	if err := writeJSON(w, http.StatusOK, data); err != nil {
		app.errorJSON(w, http.StatusInternalServerError, "internal server error")
	}
}
