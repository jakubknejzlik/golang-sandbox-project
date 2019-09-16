package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

func prepareDatabase() *gorm.DB {
	db := NewDBWithString("sqlite3://:memory:")
	db.AutoMigrate(&Message{})
	return db
}

func sendCreateMessageRequest(mux *http.ServeMux, message string, timestamp int64) (response CreateMessageResponse, err error) {
	m := Message{Message: message, Timestamp: timestamp}
	requestBody, err := json.Marshal(m)
	if err != nil {
		return
	}

	r, err := http.NewRequest("POST", "/message", bytes.NewBuffer(requestBody))
	if err != nil {
		return
	}
	r.Header.Set("content-type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf("Unexpected status code %d", resp.StatusCode)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &response)
	return response, err
}

func sendListMessageRequest(mux *http.ServeMux) ([]Message, error) {
	r, err := http.NewRequest("GET", "/messages", nil)
	if err != nil {
		return nil, err
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code %d", resp.StatusCode)
	}

	messages := []Message{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return messages, err
	}

	err = json.Unmarshal(body, &messages)
	return messages, err
}

func TestMux_message_empty_list(t *testing.T) {
	db := prepareDatabase()
	mux := GetHTTPServeMux(db)

	_, err := sendListMessageRequest(mux)
	if err != nil {
		t.Error(err)
	}
}

func TestMux_message_creation(t *testing.T) {
	db := prepareDatabase()
	mux := GetHTTPServeMux(db)

	resp, err := sendCreateMessageRequest(mux, "test message", time.Now().Unix())
	if err != nil {
		t.Error(err)
	}
	if resp.Status != "ok" {
		t.Errorf("Status should be ok and is %s", resp.Status)
	}

	messages, err := sendListMessageRequest(mux)
	if err != nil {
		t.Error(err)
	}
	if len(messages) != 1 {
		t.Errorf("Number of messages should be 1 and is %d", len(messages))
	}
}
