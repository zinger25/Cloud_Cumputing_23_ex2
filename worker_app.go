package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type EndpointsData struct {
	IP string `json:"ip"`
}

var Endpoints []EndpointsData

type RequestBody struct {
	Iterations int    `json:"iterations"`
	Buffer     string `json:"buffer"`
}

type WorkToDo struct {
	ID         uuid.UUID
	Iterations int
	Buffer     string
	T          time.Time
}

type CompWork struct {
	ID   uuid.UUID
	Hash string
}

func readEndpointsData(file_path string) {
	fileContent, _ := ioutil.ReadFile(file_path)
	json.Unmarshal(fileContent, &Endpoints)
}

func getWork() WorkToDo {
	rand.Seed(time.Now().UnixNano())
	hostname := Endpoints[rand.Intn(len(Endpoints))]

	//response, err := http.Get(fmt.Sprintf("http://localhost:8080/get_work")) //, hostname.IP))
	response, err := http.Get(fmt.Sprintf("http://%s:8080/get_work", hostname.IP))
	if err != nil {
		return WorkToDo{}
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return WorkToDo{}
		}
		defer response.Body.Close()
		var work WorkToDo
		err = json.Unmarshal(body, &work)
		return work
	}
	return WorkToDo{}
}

func computeHash(buffer []byte, iterations int) []byte {
	output := sha512.Sum512(buffer)
	for i := 1; i < iterations; i++ {
		output = sha512.Sum512(output[:])
	}
	return output[:]
}

func processWork(w WorkToDo) {
	//var requestBody RequestBody
	//err := json.NewDecoder(r.Body).Decode(&requestBody)
	//if err != nil {
	//	http.Error(w, "Failed to read request body", http.StatusBadRequest)
	//	return
	//}

	output := computeHash([]byte(w.Buffer), w.Iterations)
	cw := CompWork{
		ID:   w.ID,
		Hash: string(output),
	}

	//url := "http://localhost:8080/add_to_completed_queue"
	//jsonBody, err := json.Marshal(cw)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//_, err = http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	//if err != nil {
	//
	//}

	// send data to all endpoints
	for _, ep := range Endpoints {
		url := "http://" + ep.IP + ":8080/add_to_completed_queue"
		jsonBody, err := json.Marshal(cw)
		if err != nil {
			log.Fatal(err)
		}
		_, err = http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	}

	//w.Header().Set("Content-Type", "text/plain")
	//w.WriteHeader(http.StatusOK)
	//w.Write(output)
}

func loop() {
	for {
		work := getWork()
		if work.Buffer != "" {
			processWork(work)
		}
		time.Sleep(2 * time.Second)
	}
}

func main() {
	//readEndpointsData("./endpoints.json")
	readEndpointsData("/home/ubuntu/endpoints.json")
	loop()
}
