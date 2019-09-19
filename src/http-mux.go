package src

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

// GetHTTPServeMux Get basic HTTP mux, for larger API it's bettero to use some mux library (gin / gorilla mux / chi)
func GetHTTPServeMux(db *gorm.DB) *http.ServeMux {
	ms := NewMessageService(db)
	mux := http.NewServeMux()

	mux.HandleFunc("/message", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			http.NotFound(res, req)
			return
		}

		m := Message{}
		err := parseRequestBody(req, &m)
		if err != nil {
			sendStatus(res, false, http.StatusUnprocessableEntity)
			return
		}

		err = ms.Create(m)
		if err != nil {
			sendStatus(res, false, http.StatusBadRequest)
			return
		}

		sendStatus(res, true, http.StatusCreated)
	})

	mux.HandleFunc("/messages", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.NotFound(res, req)
			return
		}
		messages, err := ms.GetList()
		if err != nil {
			sendUnexpectedError(res, err)
			return
		}

		sendAsJSON(res, messages, 200)
	})

	return mux
}
