package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var Total_reqs = 0
var reqs_interval = 0

type Work struct {
	ID         uuid.UUID
	Iterations int
	Buffer     string
	T          time.Time
}

type CompletedWork struct {
	ID   uuid.UUID
	Hash string
}

type WorkQueue struct {
	queue     []Work
	completed []CompletedWork
	//currentID int
}

//type RequestBody struct {
//	Iterations int    `json:"iterations"`
//	Buffer     string `json:"buffer"`
//}

var numOfInstances int

//var currentID = 1

func NewWorkQueue() *WorkQueue {
	return &WorkQueue{
		queue:     []Work{},
		completed: []CompletedWork{},
	}
}

func (wq *WorkQueue) EnqueueWork(itrs int, value string) uuid.UUID {
	id, _ := uuid.NewUUID()
	work := Work{
		ID:         id,
		Iterations: itrs,
		Buffer:     value,
		T:          time.Now(),
	}
	wq.queue = append(wq.queue, work)
	//currentID++

	return work.ID
}

func (wq *WorkQueue) DequeueWork() (uuid.UUID, int, string) {
	work := wq.queue[len(wq.queue)-1]
	newQueue := wq.queue[:len(wq.queue)-1]
	wq.queue = newQueue
	return work.ID, work.Iterations, work.Buffer
}

func (wq *WorkQueue) PullCompletedWorks(num int) []CompletedWork {
	var works []CompletedWork
	if num > len(wq.completed) {
		works = wq.completed
		return works
	}
	works = wq.completed[len(wq.completed)-num:]
	return works
}

func (wq *WorkQueue) ProcessWork() {
	for {
		time.Sleep(20 * time.Second)
		if numOfInstances == 0 {
			createNewWorker()
		} else if len(wq.queue) > 0 && numOfInstances < 5 && time.Now().Sub(wq.queue[len(wq.queue)-1].T).Seconds() > 10 {
			createNewWorker()
		} else if numOfInstances > 1 && len(wq.queue) == 0 {
			removeWorker()
		}
		Total_reqs += reqs_interval
		reqs_interval = 0

		//	// add len(wq.queue) % 5
		//	instanceID := WorkersIPs[0]
		//	id, itrs, buffer := wq.DequeueWork()
		//
		//	url := "http://" + instanceID + ":8080/compute_hash"
		//	requestBody := RequestBody{
		//		Iterations: itrs,
		//		Buffer:     buffer,
		//	}
		//	jsonBody, err := json.Marshal(requestBody)
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//	defer resp.Body.Close()
		//	responseBody, err := ioutil.ReadAll(resp.Body)
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//	cw := CompletedWork{
		//		ID:   id,
		//		Hash: string(responseBody),
		//	}
		//	wq.completed = append(wq.completed, cw)
		//}
	}
}

func createNewWorker() {
	numOfInstances++
	cmd := exec.Command("bash", "worker_setup.sh")
	output, err := cmd.Output()
	lines := strings.Split(string(output), "\n")
	if err != nil {
		log.Fatal(err)
	}
	Workers = append(Workers, Worker{ID: lines[len(lines)-2],
		IP: lines[len(lines)-3]})
}

func removeWorker() {
	rand.Seed(time.Now().UnixNano())
	if numOfInstances == 1 {
		return
	}
	chosenIndex := rand.Intn(numOfInstances)
	instanceID := Workers[chosenIndex].ID

	cmd := exec.Command("bash", "remove_worker.sh", instanceID)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	numOfInstances--
}

func stopAllWorkers() {
	for i := 0; i < numOfInstances; i++ {
		id := Workers[i].ID
		cmd := exec.Command("bash", "remove_worker.sh", id)
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
	}
}

type data struct {
	Buff string `json:"buffer"`
}

func handleEnqueue(w http.ResponseWriter, r *http.Request) {
	itrsString := r.URL.Query().Get("iterations")
	itrs, err := strconv.Atoi(itrsString)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var d data
	err = json.Unmarshal(body, &d)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	workID := WQ.EnqueueWork(itrs, d.Buff)

	type Response struct {
		ID uuid.UUID `json:"ID"`
	}

	response := Response{ID: workID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePullCompleted(w http.ResponseWriter, r *http.Request) {
	numParam := r.URL.Query().Get("top")
	num, err := strconv.Atoi(numParam)
	if err != nil {
		http.Error(w, "Invalid 'top' parameter", http.StatusBadRequest)
		return
	}

	works := WQ.PullCompletedWorks(num)
	type Response struct {
		Works []CompletedWork
	}
	response := Response{Works: works}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func handleGetWork(w http.ResponseWriter, r *http.Request) {
	if len(WQ.queue) == 0 {
		http.NotFound(w, r)
		return
	}
	latest := WQ.queue[0]
	WQ.queue = WQ.queue[1:]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latest)
}

func handleAddToCompQueue(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var cw CompletedWork
	err = json.Unmarshal(body, &cw)
	WQ.completed = append(WQ.completed, cw)
}

var WQ *WorkQueue
var Workers []Worker

type Worker struct {
	ID string
	IP string
}

func awsConfigure() {
	cmd := exec.Command("bash", "aws_configure_script.sh")
	output, err := cmd.Output()
	log.Printf(string(output))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	awsConfigure()
	WQ = NewWorkQueue()
	go WQ.ProcessWork()
	http.HandleFunc("/enqueue", handleEnqueue)
	http.HandleFunc("/pull_completed", handlePullCompleted)
	http.HandleFunc("/get_work", handleGetWork)
	http.HandleFunc("/add_to_completed_queue", handleAddToCompQueue)
	http.ListenAndServe(":8080", nil)
}
