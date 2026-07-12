package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

// version은 이 바이너리의 릴리스 버전이다. 이미지 태그와 일치시킨다.
const version = "v0.1.1"

// counter는 /id 요청마다 순차적으로 증가하는 인메모리 카운터이다.
var counter atomic.Uint64

// podName은 이 Pod의 이름이다. Downward API로 주입된 POD_NAME 환경변수에서 읽는다.
var podName = func() string {
	if n := os.Getenv("POD_NAME"); n != "" {
		return n
	}
	return "unknown"
}()

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"version":   version,
		"served_by": podName,
	})
}

func idHandler(w http.ResponseWriter, r *http.Request) {
	id := counter.Add(1)
	writeJSON(w, http.StatusOK, map[string]any{
		"id":           id,
		"generated_by": podName,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /version", versionHandler)
	mux.HandleFunc("GET /id", idHandler)

	addr := ":8080"
	log.Printf("notiflex-api listening on %s (pod=%s)", addr, podName)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
