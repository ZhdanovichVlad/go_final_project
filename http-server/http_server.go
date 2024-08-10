package http_server

// the http_server package is used to correct http errors for the user.
import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type TaskResponseError struct {
	Message string `json:"error"`
}

// ResponseJson processes, serializes, and writes to Write the error. On the input you need to specify the error message, httpStatus, error, and ResponseWriter where to write the response.
func ResponseJson(message string, httpStatus int, err error, w http.ResponseWriter) {
	var responseMessage string
	if err == nil {
		responseMessage = message
	} else {
		responseMessage = fmt.Sprintf("%s : %w", message, err)
	}

	responseStruct := TaskResponseError{responseMessage}
	jsonMsg, err := json.Marshal(responseStruct)
	if err != nil {
		http.Error(w, fmt.Sprintf("error in serializing the response when an error occurs : %w", err), http.StatusInternalServerError)
	}

	w.WriteHeader(httpStatus)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(jsonMsg)
	if err != nil {
		log.Printf("error while writing response: %v", err)
	}
}
