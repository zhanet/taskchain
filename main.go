package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// TaskBlock 任务区块
type TaskBlock struct {
	Index       int
	Title       string
	Description string
	CreateStamp time.Time
	PrevHash    string
	Hash        string
}

// TaskMessage 任务消息
type TaskMessage struct {
	Title       string
	Description string
}

// TaskChain ...
type TaskChain []TaskBlock

var taskChain TaskChain

var mutex = &sync.Mutex{}

func (task TaskBlock) calculateHash() (string, error) {
	hashText := string(task.Index) + task.Title + task.Description + task.CreateStamp.String() + task.PrevHash
	h := sha256.New()
	_, err := h.Write([]byte(hashText))
	if err != nil {
		return "", err
	}
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed), nil
}

func appendBlock(tb TaskBlock) {
	mutex.Lock()
	taskChain = append(taskChain, tb)
	mutex.Unlock()
	spew.Dump(tb)
}

func genesisBlock() {
	var genesisBlock TaskBlock
	genesisBlock.Index = 0
	genesisBlock.Title = "Genesis Block"
	genesisBlock.Description = "This is Genesis Block, Copyright belong to ZHANET"
	genesisBlock.CreateStamp = time.Now()
	genesisBlock.PrevHash = ""
	hash, err := genesisBlock.calculateHash()
	if err != nil {
		log.Fatal(err)
	}
	genesisBlock.Hash = hash
	appendBlock(genesisBlock)
}

func generateBlock(oldTask TaskBlock, message TaskMessage) (TaskBlock, error) {
	var newTask TaskBlock
	newTask.Index = oldTask.Index + 1
	newTask.Title = message.Title
	newTask.Description = message.Description
	newTask.CreateStamp = time.Now()
	newTask.PrevHash = oldTask.Hash
	hash, err := newTask.calculateHash()
	if err != nil {
		return newTask, err
	}
	newTask.Hash = hash

	return newTask, nil
}

func isBlockValid(newTask, oldTask TaskBlock) bool {
	if oldTask.Index+1 != newTask.Index {
		return false
	}
	if oldTask.Hash != newTask.PrevHash {
		return false
	}

	hash, err := newTask.calculateHash()
	if err != nil || hash != newTask.Hash {
		return false
	}

	return true
}

func handleGetTaskChain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(taskChain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m TaskMessage

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newTask, err := generateBlock(taskChain[len(taskChain)-1], m)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}

	if isBlockValid(newTask, taskChain[len(taskChain)-1]) {
		appendBlock(newTask)
	}

	respondWithJSON(w, r, http.StatusCreated, newTask)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func makeRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", handleGetTaskChain).Methods("GET")
	r.HandleFunc("/", handleCreateTask).Methods("POST")
	return r
}

func server() error {
	httpPort := os.Getenv("PORT")
	_, err := strconv.Atoi(httpPort)
	if err != nil {
		log.Fatal(err)
	}

	myHandler := makeRouter()
	s := &http.Server{
		Addr:           ":" + httpPort,
		Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("TaskChain Listening on port:", httpPort)
	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// prepare()
	go genesisBlock()
	log.Fatal(server())
}
