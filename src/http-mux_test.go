package src

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/sync/semaphore"
)

func prepareDatabase() *gorm.DB {
	db := NewDBWithString("sqlite3://test.db")
	err := db.AutoMigrate(&Message{}).Error
	if err != nil {
		panic(err)
	}
	db.Delete(&Message{})
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
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("Unexpected status code %d, message: %s", resp.StatusCode, string(body))
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

func TestMux_invalid_message_creation(t *testing.T) {
	db := prepareDatabase()
	mux := GetHTTPServeMux(db)

	r, err := http.NewRequest("POST", "/message", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Error(err)
		return
	}
	r.Header.Set("content-type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("Unexpected status code %d, message: %s", resp.StatusCode, string(body))
		return
	}
}

func TestMux_message_creation_parallel(t *testing.T) {
	db := prepareDatabase()
	mux := GetHTTPServeMux(db)

	numClients := 4
	sem := semaphore.NewWeighted(int64(numClients))

	running := true
	numberOfMessages := 0
	var clientError error
	ctx := context.Background()

	for i := 0; i < numClients; i++ {
		go (func(clientIndex int) {
			for running {
				sem.Acquire(ctx, 1)
				resp, err := sendCreateMessageRequest(mux, fmt.Sprintf("test message from client %d", clientIndex), time.Now().Unix())
				if err != nil {
					clientError = err
				}
				if resp.Status != "ok" {
					clientError = fmt.Errorf("Status should be ok and is %s", resp.Status)
				}

				if clientError != nil {
					return
				}

				numberOfMessages++
				sem.Release(1)
				time.Sleep(time.Millisecond * 50)
			}
		})(i)
	}

	time.Sleep(time.Millisecond * 100)

	for numberOfMessages > 0 && numberOfMessages < 200 {
		sem.Acquire(ctx, int64(numClients))
		if clientError != nil {
			t.Errorf("Failed with client error %s", clientError.Error())
			return
		}
		messages, err := sendListMessageRequest(mux)
		if err != nil {
			t.Error(err)
		}
		expectedNumberOfMessages := int(math.Min(float64(numberOfMessages), 100.0))
		if expectedNumberOfMessages != len(messages) {
			t.Errorf("Number of fetched messages should be %d and is %d", expectedNumberOfMessages, len(messages))
		}
		for i := 0; i < len(messages)-2; i++ {
			if messages[i].Timestamp < messages[i+1].Timestamp {
				t.Errorf("Timestamp should be ordered by timestamp %d > %d", messages[i].Timestamp, messages[i+1].Timestamp)
			}
		}
		sem.Release(int64(numClients))
	}

	running = false
	// wait for clients finish
	sem.Acquire(ctx, int64(numClients))
	sem.Release(int64(numClients))
}
