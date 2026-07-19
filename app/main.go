package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	valkey "github.com/valkey-io/valkey-go"
)

// version은 이 바이너리의 릴리스 버전이다.
// 이미지 태그는 CI가 git SHA(sha-<7자리>) 기반으로 자동 부여한다 (3.5부터).
const version = "v0.3.1"

// idKey는 /id 카운터를 저장하는 Valkey 키이다. 모든 Pod이 같은 키를 INCR하므로
// 인메모리 카운터와 달리 Pod 수와 무관하게 전역 순차 ID가 보장된다.
const idKey = "notiflex:id"

// valkeyClient는 모든 핸들러가 공유하는 Valkey 커넥션이다.
var valkeyClient valkey.Client

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
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	id, err := valkeyClient.Do(ctx, valkeyClient.B().Incr().Key(idKey).Build()).AsInt64()
	if err != nil {
		log.Printf("valkey INCR failed: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "id backend unavailable"})
		return
	}
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

// connectValkey는 Valkey에 연결한다. Pod 기동 시 DNS 전파나 Valkey 기동이
// 늦어질 수 있으므로 3초 간격으로 최대 10회 재시도한다. 재시도 없이 즉시
// 종료하면 CrashLoopBackOff에 빠진다.
func connectValkey(addr, password string) (valkey.Client, error) {
	var client valkey.Client
	var err error
	for i := 0; i < 10; i++ {
		client, err = valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{addr},
			Password:    password,
		})
		if err == nil {
			return client, nil
		}
		log.Printf("Valkey 연결 재시도 %d/10: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	return nil, err
}

func main() {
	addr := os.Getenv("VALKEY_ADDR")
	if addr == "" {
		log.Fatal("VALKEY_ADDR is required")
	}
	// ch6.2: 비밀번호는 Secret Manager CSI가 마운트한 파일을 우선 사용한다.
	// 파일 경로가 없으면 환경변수(VALKEY_PASSWORD)로 폴백한다.
	password := os.Getenv("VALKEY_PASSWORD")
	if pwFile := os.Getenv("VALKEY_PASSWORD_FILE"); pwFile != "" {
		data, err := os.ReadFile(pwFile)
		if err != nil {
			log.Fatalf("failed to read password file %s: %v", pwFile, err)
		}
		password = string(data)
	}

	client, err := connectValkey(addr, password)
	if err != nil {
		log.Fatalf("valkey connect failed: %v", err)
	}
	defer client.Close()
	valkeyClient = client

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /version", versionHandler)
	mux.HandleFunc("GET /id", idHandler)

	listenAddr := ":8080"
	log.Printf("notiflex-api listening on %s (pod=%s, valkey=%s)", listenAddr, podName, addr)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
