package handler

import (
	"gotest/core/app"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	engine := app.SetupEngine(nil)
	engine.ServeHTTP(w, r)
}
