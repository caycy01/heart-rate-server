package handlers

import (
	"heart-rate-server/internal/utils"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	utils.SendResponse(w, http.StatusOK, "OK", nil)
}
