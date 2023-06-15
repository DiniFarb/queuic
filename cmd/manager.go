package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dinifarb/mlog"
	"github.com/dinifarb/queuic/pkg/proto"
)

type Manager struct {
	http.ServeMux
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Start() error {
	m.HandleFunc("/stats", m.statsHandler)
	m.HandleFunc("/createQueue", m.createQueueHandler)
	//m.HandleFunc("/deleteQueue", m.deleteQueueHandler)
	m.HandleFunc("/enqueue", m.enqueueHandler)
	mlog.Info("starting http interface on port 8080")
	return http.ListenAndServe(":8080", m)
}

func (m *Manager) statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(srv.GetStats())
}

type CreateQueueRequest struct {
	QueueName string `json:"queueName"`
}

func (m *Manager) createQueueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode("method not allowed")
		return
	}
	var body CreateQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("bad request")
		return
	}
	name := proto.QueueName{}
	name.ParseFromString(body.QueueName)
	if err := srv.CreateQueue(name); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("internal server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("queue created")
}

type EnqueueRequest struct {
	QueueName string `json:"queueName"`
	Message   string `json:"message"`
}

func (m *Manager) enqueueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode("method not allowed")
		return
	}
	var body EnqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("bad request")
		return
	}
	name := proto.QueueName{}
	name.ParseFromString(body.QueueName)
	if err := srv.Enqueue(name, []byte(body.Message)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf(`{"error": "internal server error: %s"}`, err.Error()))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("message enqueued")
}
