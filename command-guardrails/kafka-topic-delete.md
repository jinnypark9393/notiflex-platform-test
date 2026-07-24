# Kafka Topic 삭제

## 사전 확인
1. Topic에 미처리 메시지가 있는지 확인
   - `kubectl exec -n kafka notiflex-kafka-single-0 -c kafka -- bin/kafka-get-offsets.sh --bootstrap-server localhost:9092 --topic <topic>`
2. Consumer(notiflex-api)가 모두 처리를 완료했는지 로그로 확인
3. 이 Topic을 사용하는 Producer 목록 파악 (현재는 notiflex-api의 `/id` 발행)

## 실행
1. 관련 Producer를 먼저 중지 (메시지 유입 차단) — 앱 배포에서 KAFKA_BROKER env 제거 or 발행 비활성
2. Consumer가 잔여 메시지를 모두 처리할 때까지 대기
3. KafkaTopic 리소스 삭제 — GitOps 경유: `k8s/kafka/manifests/topic.yaml`에서 제거 후 git push (kubectl delete 직접 사용 금지, ArgoCD가 selfHeal로 되돌림)

## 사후 검증
1. `kubectl get kafkatopic -n kafka`로 Topic이 삭제되었는지 확인
2. notiflex-api 로그에 발행/구독 에러가 없는지 확인
