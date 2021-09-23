package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type server struct{}

type DB struct {
	listOfCreateRequests []createRequest
	mux                  sync.Mutex
}

var database DB

func (db *DB) Append(createRequestInstance createRequest) {
	db.mux.Lock()
	db.listOfCreateRequests = append(db.listOfCreateRequests, createRequestInstance)
	db.mux.Unlock()
}

type keySctruct struct {
	Key string `json: "key"`
}

type duplicateKeyStruct struct {
	Key         string
	OriginalKey string
	Timestamp   time.Time
	CreateTime  time.Time
}

type listOfDuplicateKeysType struct {
	listOfDuplicateKeysType []duplicateKeyStruct
	mux                     sync.Mutex
}

func (l *listOfDuplicateKeysType) Append(duplicateKeyStructInstance duplicateKeyStruct) {
	l.mux.Lock()
	l.listOfDuplicateKeysType = append(l.listOfDuplicateKeysType, duplicateKeyStructInstance)
	l.mux.Unlock()
}

var duplicateKeys listOfDuplicateKeysType

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
	log.Println("Server started on port :8080")
	r.HandleFunc("/", get).Methods(http.MethodGet)
	r.HandleFunc("/", post).Methods(http.MethodPost)
	r.HandleFunc("/", notFound)
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Handle get request for getting data
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

	cR := findCr(rawKey.Key)

	j, err := json.Marshal(cR)
	if err != nil {
		log.Fatal(err.Error())
	}
	w.Write(j)
}

// Find data in database or in duplicate keys
func findCr(key string) createRequest {
	oldData := false
	cR := findCRInDatabase(key)
	if isOld(cR.Timestamp) {
		oldData = true
		cR = *(new(createRequest))
	}

	duplicateKeys.mux.Lock()
	defer duplicateKeys.mux.Unlock()
	for _, e := range duplicateKeys.listOfDuplicateKeysType {
		if e.Key == key {
			oldData = isOld(e.Timestamp)
			if oldData {
				continue
			}

			tmpCR := findCRInDatabase(e.OriginalKey)
			if cR.CreateTime.Sub(e.CreateTime) < 0 {
				tmpCR.Key = e.Key
				tmpCR.CreateTime = e.CreateTime
				tmpCR.Timestamp = e.Timestamp
				cR = tmpCR
			}
		}
	}
	if oldData {
		go cleanData()
	}

	return cR
}

// Check if timestamp is expired
func isOld(t time.Time) bool {
	return time.Now().Sub(t) > 0
}

// Find data in database
func findCRInDatabase(key string) createRequest {
	cR := new(createRequest)

	database.mux.Lock()
	defer database.mux.Unlock()
	for _, e := range database.listOfCreateRequests {
		if e.Key == key {
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

// Handle post request for adding data or duplicate keys
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

// Clean all old data
func cleanData() {
	//Clean data
	database.mux.Lock()
	var newListOfCreateRequests []createRequest
	for i := 0; i < len(database.listOfCreateRequests); i++ {
		e := database.listOfCreateRequests[i]
		if isOld(e.Timestamp) {
			// Replace data key by duplicate key and clean duplicate keys
			duplicateKeys.mux.Lock()
			for i := 0; i < len(duplicateKeys.listOfDuplicateKeysType); i++ {
				dK := duplicateKeys.listOfDuplicateKeysType[i]
				if isOld(dK.Timestamp) {
					continue
				}
				// Replace
				if e.Key == dK.OriginalKey {
					e.Key = dK.Key
					e.Timestamp = dK.Timestamp
					e.CreateTime = dK.CreateTime
					newListOfCreateRequests = append(newListOfCreateRequests, e)
					continue
				}
			}
			duplicateKeys.mux.Unlock()
		} else {
			newListOfCreateRequests = append(newListOfCreateRequests, e)
		}
	}

	database.listOfCreateRequests = newListOfCreateRequests
	database.mux.Unlock()

	cleanDuplicateKeys()
	runtime.GC()
}

//Clean duplicate keys
func cleanDuplicateKeys() {
	duplicateKeys.mux.Lock()
	var newListOfDuplicateKeysType []duplicateKeyStruct
	for i := 0; i < len(duplicateKeys.listOfDuplicateKeysType)-1; i++ {
		e := duplicateKeys.listOfDuplicateKeysType[i]
		if isOld(e.Timestamp) {
			continue
		}
		newListOfDuplicateKeysType = append(newListOfDuplicateKeysType, e)
	}
	duplicateKeys.listOfDuplicateKeysType = newListOfDuplicateKeysType
	duplicateKeys.mux.Unlock()
}

// Add key to list of duplicates
func addToListOfDuplicateKeys(rawJsonRequest createRequestRaw, t time.Time, duplicateKey string) {
	var dK duplicateKeyStruct
	dK.Key = rawJsonRequest.Key
	dK.OriginalKey = duplicateKey
	dK.Timestamp = t
	dK.CreateTime = time.Now()

	duplicateKeys.Append(dK)
}

// Add data to database
func addToDatabase(rawJsonRequest createRequestRaw, t time.Time) {
	var cR createRequest
	cR.Key = rawJsonRequest.Key
	cR.Data = []byte(rawJsonRequest.Data)
	cR.Timestamp = t
	cR.CreateTime = time.Now()

	database.Append(cR)
}

// Find data in database
func findDataInDatabase(data []byte) string {
	database.mux.Lock()
	defer database.mux.Unlock()

	for _, e := range database.listOfCreateRequests {
		if bytes.Compare(e.Data, data) == 0 {
			return e.Key
		}
	}
	return ""
}

// Handle invalid request type
func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "not found"}`))
}
