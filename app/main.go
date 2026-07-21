package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IBM/sarama"
	valkey "github.com/valkey-io/valkey-go"
)

// version은 이 바이너리의 릴리스 버전이다.
// 이미지 태그는 CI가 git SHA(sha-<7자리>) 기반으로 자동 부여한다 (3.5부터).
const version = "v0.4.1"

// idKey는 /id 카운터를 저장하는 Valkey 키이다. 모든 Pod이 같은 키를 INCR하므로
// 인메모리 카운터와 달리 Pod 수와 무관하게 전역 순차 ID가 보장된다.
const idKey = "notiflex:id"

// notificationsTopic은 /id 생성 이벤트를 발행하는 Kafka 토픽이다 (8.1).
const notificationsTopic = "notifications"

// valkeyClient는 모든 핸들러가 공유하는 Valkey 커넥션이다.
var valkeyClient valkey.Client

// kafkaProducer는 notifications 토픽에 이벤트를 발행한다. Kafka 미설정 시 nil이다.
var kafkaProducer sarama.SyncProducer

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
	// 8.1: ID 생성 이벤트를 Kafka에 발행한다(비동기 후처리 파이프라인의 진입점).
	// Kafka가 없거나 발행에 실패해도 /id 응답 자체는 성공으로 처리한다(캐시가 진실 원천).
	publishNotification(id)
	writeJSON(w, http.StatusOK, map[string]any{
		"id":           id,
		"generated_by": podName,
	})
}

// publishNotification은 생성된 ID를 notifications 토픽에 발행한다.
func publishNotification(id int64) {
	if kafkaProducer == nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{"id": id, "generated_by": podName})
	msg := &sarama.ProducerMessage{
		Topic: notificationsTopic,
		Value: sarama.ByteEncoder(payload),
	}
	if _, _, err := kafkaProducer.SendMessage(msg); err != nil {
		log.Printf("kafka publish failed: %v", err)
	}
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

// newKafkaConfig는 Kafka 4.x 브로커에 맞춘 sarama 설정을 만든다.
func newKafkaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_1_0_0
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Return.Successes = true
	return cfg
}

// startConsumer는 백그라운드에서 notifications 토픽의 모든 파티션을 구독해 수신
// 메시지를 로그로 출력한다. 토픽이 다중 파티션(3)이라 파티션 0만 구독하면 다른
// 파티션에 발행된 메시지를 놓치므로, 파티션 목록을 받아 각각 구독한다.
func startConsumer(broker string) {
	consumer, err := sarama.NewConsumer([]string{broker}, newKafkaConfig())
	if err != nil {
		log.Printf("kafka consumer 생성 실패: %v", err)
		return
	}
	partitions, err := consumer.Partitions(notificationsTopic)
	if err != nil {
		log.Printf("kafka partition 목록 조회 실패: %v", err)
		return
	}
	log.Printf("kafka consumer 시작: topic=%s partitions=%v", notificationsTopic, partitions)
	for _, p := range partitions {
		pc, err := consumer.ConsumePartition(notificationsTopic, p, sarama.OffsetNewest)
		if err != nil {
			log.Printf("kafka partition %d 구독 실패: %v", p, err)
			continue
		}
		go func(pc sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				log.Printf("kafka 수신: %s", string(msg.Value))
			}
		}(pc)
	}
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

	// 8.1: Kafka는 선택적. KAFKA_BROKER가 있으면 Producer/Consumer를 기동한다.
	if broker := os.Getenv("KAFKA_BROKER"); broker != "" {
		producer, err := sarama.NewSyncProducer([]string{broker}, newKafkaConfig())
		if err != nil {
			log.Printf("kafka producer 생성 실패(발행 없이 계속): %v", err)
		} else {
			kafkaProducer = producer
			defer producer.Close()
			go startConsumer(broker)
			log.Printf("kafka 연결됨: broker=%s", broker)
		}
	}

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
