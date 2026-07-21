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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// version은 이 바이너리의 릴리스 버전이다.
// 이미지 태그는 CI가 git SHA(sha-<7자리>) 기반으로 자동 부여한다 (3.5부터).
const version = "v0.5.0"

// tracer는 각 핸들러에서 span을 만들 때 사용한다. OTel 미설정 시 no-op tracer이다.
var tracer = otel.Tracer("notiflex-api")

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
	// 8.2: /id 요청 전체를 span으로 감싼다. Valkey INCR·Kafka 발행이 하위 span으로 이어진다.
	ctx, span := tracer.Start(r.Context(), "idHandler")
	defer span.End()

	incrCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	incrCtx, incrSpan := tracer.Start(incrCtx, "valkey.incr")
	id, err := valkeyClient.Do(incrCtx, valkeyClient.B().Incr().Key(idKey).Build()).AsInt64()
	incrSpan.End()
	if err != nil {
		log.Printf("valkey INCR failed: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "id backend unavailable"})
		return
	}
	// 8.1: ID 생성 이벤트를 Kafka에 발행한다(비동기 후처리 파이프라인의 진입점).
	// Kafka가 없거나 발행에 실패해도 /id 응답 자체는 성공으로 처리한다(캐시가 진실 원천).
	publishNotification(ctx, id)
	writeJSON(w, http.StatusOK, map[string]any{
		"id":           id,
		"generated_by": podName,
	})
}

// publishNotification은 생성된 ID를 notifications 토픽에 발행한다.
func publishNotification(ctx context.Context, id int64) {
	if kafkaProducer == nil {
		return
	}
	_, span := tracer.Start(ctx, "kafka.publish")
	defer span.End()
	payload, _ := json.Marshal(map[string]any{"id": id, "generated_by": podName})
	msg := &sarama.ProducerMessage{
		Topic: notificationsTopic,
		Value: sarama.ByteEncoder(payload),
	}
	if _, _, err := kafkaProducer.SendMessage(msg); err != nil {
		log.Printf("kafka publish failed: %v", err)
	}
}

// initTracer는 OTLP gRPC exporter로 트레이스를 Tempo에 보내는 TracerProvider를 설정한다.
// OTEL_EXPORTER_OTLP_ENDPOINT가 없으면 트레이싱을 비활성화한다(no-op tracer 유지).
// 반환된 shutdown 함수는 종료 시 남은 span을 flush한다.
func initTracer(ctx context.Context) (func(context.Context) error, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName("notiflex-api")),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("notiflex-api")
	log.Printf("otel tracing 활성화: endpoint=%s", endpoint)
	return tp.Shutdown, nil
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

	// 8.2: OTel 트레이싱 초기화(OTEL_EXPORTER_OTLP_ENDPOINT 있을 때만 활성).
	shutdownTracer, err := initTracer(context.Background())
	if err != nil {
		log.Printf("otel tracer 초기화 실패(트레이싱 없이 계속): %v", err)
	} else {
		defer func() { _ = shutdownTracer(context.Background()) }()
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
