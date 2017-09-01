package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

const (
	Namespace  = "namespace"
	ObjectType = "objecttype"
	ObjectName = "objectname"
)

func main() {

	format := new(log.JSONFormatter)
	format.TimestampFormat = "2006-01-02 15:04:05"

	log.SetFormatter(format)
	log.SetOutput(os.Stdout)

	r := CreateRouter(log.StandardLogger())

	http.Handle("/", r)
	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		fmt.Println("unable to start server", "localhost", err)
	}

}

// CreateRouter sets up the main mux
func CreateRouter(logger *log.Logger) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc(fmt.Sprintf("/log/{%s}/{%s}/{%s}", Namespace, ObjectType, ObjectName), Logger(logger)).Methods("POST")
	r.HandleFunc("/status", func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(http.StatusOK)
	})
	return r
}

// Logger logs incomming request bodies as raw or json to stdout in json format
func Logger(logger *log.Logger) func(http.ResponseWriter, *http.Request) {

	return func(response http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)

		metadata := map[string]interface{}{
			Namespace:  vars[Namespace],
			ObjectType: vars[ObjectType],
			ObjectName: vars[ObjectName],
		}
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte(err.Error()))
			return
		}

		if request.Header.Get("Content-Type") == "application/json" {
			var jsonBody map[string]interface{}
			err := json.Unmarshal(body, &jsonBody)
			if err != nil {
				LogRaw(logger, metadata, string(body))
				response.WriteHeader(http.StatusBadRequest)
				response.Write([]byte(err.Error()))
				return
			}
			LogJSON(logger, metadata, jsonBody)
		} else {
			LogRaw(logger, metadata, string(body))
		}
		response.WriteHeader(http.StatusOK)
	}

}

// LogRaw logs all non application/json types with request body as msg
func LogRaw(logger *log.Logger, fields map[string]interface{}, msg string) {
	entry := log.NewEntry(logger)
	entry.WithFields(fields).Info(msg)
}

// LogJSON logs all application/json types with request body as fields in the field 'json'
func LogJSON(logger *log.Logger, fields map[string]interface{}, json map[string]interface{}) {
	entry := log.NewEntry(logger)
	entry.WithFields(fields).WithField("json", json).Info("")
}
