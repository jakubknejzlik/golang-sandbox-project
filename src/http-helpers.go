package src

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// CreateMessageResponse basic status response for /message route
type CreateMessageResponse struct {
	Status string
}

func sendStatus(r http.ResponseWriter, isOk bool, statusCode int) {
	status := "ok"
	if !isOk {
		status = "nok"
	}
	response := CreateMessageResponse{Status: status}

	sendAsJSON(r, response, statusCode)
}

func sendAsJSON(r http.ResponseWriter, object interface{}, statusCode int) {
	body, err := json.Marshal(object)
	if err != nil {
		sendUnexpectedError(r, err)
		return
	}

	r.Header().Set("content-type", "application/json")
	r.WriteHeader(statusCode)
	r.Write(body)
}

func sendUnexpectedError(r http.ResponseWriter, err error) {
	r.WriteHeader(http.StatusInternalServerError)
	r.Write([]byte(err.Error()))
}

func parseRequestBody(req *http.Request, m interface{}) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &m)
}
