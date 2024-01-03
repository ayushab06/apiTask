package utility

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool
	Message string
}

func Respond(statusCode int, message string, w *http.ResponseWriter, success bool) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(statusCode)
	res := Response{Success: success, Message: message}
	data, _ := json.Marshal(res)
	(*w).Write(data)
}

func RespondStruct(data []byte, w *http.ResponseWriter, success bool) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusAccepted)
	(*w).Write(data)
}
