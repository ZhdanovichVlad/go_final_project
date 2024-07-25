package http_server

import (
	"encoding/json"
	"net/http"
)

type TaskResponseError struct {
	Message string `json:"error"`
}

func JsonErrorMarshal(message TaskResponseError, isBadRequest bool) ([]byte, int) {
	var returnStatus int
	if isBadRequest {
		returnStatus = http.StatusBadRequest
	} else {
		returnStatus = http.StatusInternalServerError
	}
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return []byte(err.Error()), http.StatusInternalServerError
	}
	return jsonMsg, returnStatus
}
