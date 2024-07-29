package http_server

// пакет http_server используется для формирования ошибки в формате json по заданной структуре TaskResponseError
import (
	"encoding/json"
	"net/http"
)

type TaskResponseError struct {
	Message string `json:"error"`
}

// JsonErrorMarshal generates an error in json format. The input is to specify the error message in the TaskResponseError structure, whether the error is a server error or a bad request.
// JsonErrorMarshal формирует ошибку в формате json. На входе нужно указать сообщение ошибки в структуре TaskResponseError, является ли ошибка ошибкой сервера или плохим запросом.
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
