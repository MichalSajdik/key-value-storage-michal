package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type server struct{}

type DB []createRequest

var database DB

type keySctruct struct {
	Key string `json: "key"`
}

type duplicateKeyStruct struct {
	Key         string
	OriginalKey string
	Timestamp   time.Time
	CreateTime  time.Time
}

type listOfDuplicateKeysType []duplicateKeyStruct

var listOfDuplicateKeys listOfDuplicateKeysType

type createRequest struct {
	Key        string
	Data       []byte
	Timestamp  time.Time
	CreateTime time.Time
}

type createRequestRaw struct {
	Key       string `json: "key"`
	Data      []byte `json: "data"`
	Timestamp string `json: "expiration_date"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", get).Methods(http.MethodGet)
	r.HandleFunc("/", post).Methods(http.MethodPost)
	r.HandleFunc("/", notFound)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	var rawKey keySctruct

	err = json.Unmarshal(body, &rawKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	cR := findCRInDatabase(rawKey.Key)

	j, err := json.Marshal(cR)
	if err != nil {
		log.Fatal(err.Error())
	}
	w.Write(j)
}

func findCr(key string) createRequest {
	cR := findCRInDatabase(key)

	for _, e := range listOfDuplicateKeys {
		if e.Key == key {
			if time.Now().Sub(e.Timestamp) > 0 {
				continue
			}

			tmpCR := findCRInDatabase(e.OriginalKey)
			if cR.CreateTime.Sub(e.CreateTime) < 0 {
				cR = tmpCR
			}
		}
	}

	return cR
}

func findCRInDatabase(key string) createRequest {
	cR := new(createRequest)
	for _, e := range database {
		if e.Key == key {
			if time.Now().Sub(e.Timestamp) > 0 {
				continue
			}

			if cR == nil {
				cR = &e
			}

			//9. Priority data: when multiple values under the same key, we return the last saved data
			if cR.CreateTime.Sub(e.CreateTime) < 0 {
				cR = &e
			}
		}
	}

	return *cR
}

func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer r.Body.Close()

	var rawJsonRequest createRequestRaw
	err = json.Unmarshal(body, &rawJsonRequest)
	if err != nil {
		log.Fatal(err.Error())
	}

	t, err := time.Parse(time.RFC3339, rawJsonRequest.Timestamp)
	if err != nil {
		log.Fatal(err.Error())
	}

	duplicateKey := findDataInDatabase([]byte(rawJsonRequest.Data))

	if duplicateKey != "" {
		addToListOfDuplicateKeys(rawJsonRequest, t, duplicateKey)
	} else {
		addToDatabase(rawJsonRequest, t)
	}
}

func addToListOfDuplicateKeys(rawJsonRequest createRequestRaw, t time.Time, duplicateKey string) {
	var dK duplicateKeyStruct
	dK.Key = rawJsonRequest.Key
	dK.OriginalKey = duplicateKey
	dK.Timestamp = t
	dK.CreateTime = time.Now()

	listOfDuplicateKeys = append(listOfDuplicateKeys, dK)
}

func addToDatabase(rawJsonRequest createRequestRaw, t time.Time) {
	var cR createRequest
	cR.Key = rawJsonRequest.Key
	cR.Data = []byte(rawJsonRequest.Data)
	cR.Timestamp = t
	cR.CreateTime = time.Now()

	database = append(database, cR)
}

func findDataInDatabase(data []byte) string {
	for _, e := range database {
		if bytes.Compare(e.Data, data) == 0 {
			return e.Key
		}
	}
	return ""
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "not found"}`))
}
