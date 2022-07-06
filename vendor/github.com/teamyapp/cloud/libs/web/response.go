package web

import (
	"net/http"
)

func WriteJSON(writer http.ResponseWriter, body []byte) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	writer.Write(body)
}
