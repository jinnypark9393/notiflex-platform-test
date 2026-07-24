# Architecture Decision Records

한 결정에 한 항목, 시간 순서대로 누적한다. 결정의 "왜"를 남겨 다른 머신·다른 시점에서도 같은 근거로 판단할 수 있게 한다.

## ADR-001: 인프라 프로비저닝은 Terraform (2장)
**시점**: 2026-07 / **결정**: GKE 클러스터·노드풀·Artifact Registry를 gcloud CLI가 아닌 Terraform(IaC)으로 생성한다
**이유**:
- 인프라 구성이 Git으로 버전 관리되어 재현 가능
- `plan`으로 변경 영향을 적용 전에 검토 가능
- 실습 종료 시 `destroy`로 리소스 정리가 용이
- 책 기본 흐름(gcloud)과 달리 실무 표준 워크플로우를 그대로 연습

## ADR-002: Terraform state는 GCS backend (2장)
**시점**: 2026-07 / **결정**: local state 대신 versioning이 켜진 private GCS 버킷을 backend로 사용하고, 폴더별 prefix로 state를 분리한다
**이유**:
- state 유실 방지 (버킷 versioning으로 복구 가능)
- 팀 공유를 전제로 한 표준 구성과 동일
- 폴더(gke/apps)별 prefix 분리로 blast radius 축소
- 버킷 자체는 gcloud로 부트스트랩 (닭-달걀 문제 회피)

## ADR-003: Terraform 버전은 tfenv + .terraform-version 고정 (2장)
**시점**: 2026-07 / **결정**: 시스템 전역 설치 대신 tfenv로 프로젝트별 Terraform 버전을 고정한다
**이유**:
- 프로젝트별 버전 고정으로 동작 재현성 확보
- 팀원 간 버전 불일치로 인한 state 오염 방지
- provider 버전도 01-provider.tf에 static 고정하여 이중 방어

## ADR-004: Terraform 변수 주입은 direnv .envrc (2장)
**시점**: 2026-07 / **결정**: tfvars 파일 대신 direnv `.envrc`의 `TF_VAR_*` 환경변수로 로컬 값을 주입한다
**이유**:
- 프로젝트 ID 등 로컬 값을 코드와 분리
- 디렉터리 진입만으로 환경 자동 구성
- 커밋 대상이 아니므로 민감 값이 저장소에 남지 않음

## ADR-005: 이미지 빌드는 Cloud Build (2장)
**시점**: 2026-07 / **결정**: 로컬 Docker buildx 대신 `gcloud builds submit` 원격 빌드를 사용한다
**이유**:
- 로컬 Docker 데몬 불필요, 환경 독립적
- amd64 노드 대상 크로스컴파일 문제 원천 회피
- GCP 네이티브라 Artifact Registry 푸시 권한 연동이 단순

## ADR-006: 이미지 저장소는 Artifact Registry (2장)
**시점**: 2026-07 / **결정**: Docker Hub·GCR 대신 Artifact Registry를 Terraform으로 생성해 사용한다
**이유**:
- GKE와 같은 프로젝트/리전이라 pull 지연·비용 최소화
- IAM 기반 접근 제어를 GCP에서 일원화
- IaC(Terraform)로 저장소 수명주기 관리

## ADR-007: GitOps 도구는 ArgoCD (3장)
**시점**: 2026-07 / **결정**: Flux 대신 ArgoCD를 채택하여 Git(k8s/smb)을 단일 진실 원천으로 자동 배포한다
**이유**:
- Web UI로 배포 상태를 실시간 확인 → "지금 무슨 일이 일어나는지" 눈으로 볼 수 있다
- Application CRD로 "어떤 Git 경로 → 어떤 네임스페이스" 선언적 관리
- selfHeal: 누군가 kubectl로 직접 수정해도 Git 상태로 되돌린다
- GKE Standard와 네이티브 호환, e2-medium 노드에서 구동 가능 (~500MB 메모리)

## ADR-008: CI는 GitHub Actions + WIF keyless 인증 (3장)
**시점**: 2026-07 / **결정**: Jenkins 등 별도 CI 서버 대신 GitHub Actions를 쓰고, GCP 인증은 SA 키 JSON 대신 Workload Identity Federation(GitHub OIDC)으로 한다
**이유**:
- GitHub 네이티브: 코드 저장소와 CI가 같은 플랫폼 → 별도 서버 설치/관리 불필요
- YAML 선언적: ci.yaml 한 파일로 빌드→푸시→매니페스트 갱신 파이프라인 정의
- 퍼블릭 저장소 무료, 프라이빗도 월 2,000분 무료
- WIF keyless: 장기 크레덴셜을 저장하지 않아 유출·로테이션 부담 제거 (책 방식 A인 SA 키 대신 채택, pool/provider/binding은 Terraform 관리)

## ADR-009: 메트릭은 kube-prometheus-stack (4장)
**시점**: 2026-07 / **결정**: Datadog 등 SaaS 대신 kube-prometheus-stack Helm 차트로 자체 호스팅 모니터링을 구성한다
**이유**:
- 오픈소스 표준: Kubernetes 모니터링의 사실상 표준 (CNCF Graduated)
- SaaS 구독료 없이 자체 호스팅
- Helm 번들: 6개 컴포넌트를 검증된 버전 조합으로 한 번에 설치
- Grafana 대시보드로 시각화, Loki/Tempo와 같은 UI에서 통합 조회

## ADR-010: 로그는 Loki + Fluent Bit (4장)
**시점**: 2026-07 / **결정**: ELK 대신 Loki(SingleBinary) + Fluent Bit(DaemonSet) 조합으로 로그를 수집한다
**이유**:
- 경량: Loki 128Mi, Fluent Bit 64Mi — e2-medium에서 ELK(2Gi+)는 불가능
- Grafana 통합: 메트릭(Prometheus)과 같은 UI에서 로그 조회
- 라벨 기반 인덱싱: 풀텍스트 인덱싱 대비 저장 비용이 낮다
- 리소스 예산상 최신 차트의 캐시/카나리/게이트웨이 컴포넌트는 비활성화

## ADR-011: 외부 트래픽은 Gateway API (5장)
**시점**: 2026-07 / **결정**: Ingress(NGINX Ingress Controller 등) 대신 GKE managed Gateway API(`gke-l7-regional-external-managed`)로 외부 진입점을 만든다
**이유**:
- K8s 공식 표준: Ingress를 대체하는 차세대 API (GA since K8s 1.27)
- GKE 네이티브: 별도 Ingress Controller 설치 없이 GKE가 자동으로 처리
- 역할 분리: Gateway(인프라팀) / HTTPRoute(앱팀)로 관심사 분리
- 5.3 Blue/Green 연동: HTTPRoute의 backendRefs로 트래픽 분배 가능

## ADR-012: 배포 전략 도구는 Argo Rollouts (5장)
**시점**: 2026-07 / **결정**: Flagger·Istio 대신 Argo Rollouts를 설치하고 Deployment를 Rollout(Blue/Green, 30초 auto-promote)으로 전환한다
**이유**:
- ArgoCD 통합: 같은 Argo 생태계, ArgoCD UI에서 Rollout 상태 확인 가능
- CRD 기반: YAML 선언으로 배포 전략 정의, GitOps 호환
- 점진적 진화: 5장 Blue/Green → 6장 Canary 전환 시 Rollout CRD만 수정
- kubectl 플러그인으로 배포 진행 상태 실시간 모니터링

## ADR-013: 캐시는 Valkey (6장)
**시점**: 2026-07 / **결정**: Redis·Memcached·DragonflyDB 대신 Valkey로 Pod 간 ID 카운터를 공유한다
**이유**:
- BSD 라이선스: Redis의 SSPL 라이선스 제약(상용 제한) 없이 사용
- Redis 프로토콜 호환: 기존 Redis 클라이언트·명령어(INCR 등) 그대로 사용
- INCR + 영속성: ID 생성에 INCR이 필요하고 Pod 재시작 후 카운터 유지 필요 — Memcached는 영속성 없어 부적합
- 경량: standalone 모드로 CPU 50m, Memory 64Mi

## ADR-014: 시크릿은 Google Secret Manager + CSI (6장)
**시점**: 2026-07 / **결정**: K8s Secret·Sealed Secrets·External Secrets Operator 대신 Google Secret Manager + GKE managed CSI Driver + Workload Identity로 시크릿을 관리한다
**이유**:
- GKE 네이티브: Workload Identity가 GKE와 GCP IAM을 직접 연결, SA 키 JSON 불필요
- 단일 진실 원천: Secret Manager가 시크릿의 유일한 저장소, 앱과 Valkey가 같은 값을 CSI 파일로 읽음
- addon 활성화: GKE managed CSI는 `--enable-secret-manager` 한 줄, 오픈소스 helm 설치 불필요
- keyless: 장기 크레덴셜 미보관 (WIF와 동일 철학)

## ADR-015: 배포 전략을 Canary로 전환 (6장)
**시점**: 2026-07 / **결정**: Blue/Green에서 Canary(20→50→80→100%)로 전환한다. 도구는 Argo Rollouts 그대로, strategy 블록만 교체
**이유**:
- 위험도 최소화: 새 버전에 트래픽을 점진적으로 늘려 문제를 조기 발견
- 빠른 abort: 이상 시 즉시 중단 가능, 롤백이 Blue/Green보다 세밀
- 리소스 효율: Canary 1.2x vs Blue/Green 2x (전환 순간 파드 부담 완화)
- 점진적 고도화: 도구 교체 없이 같은 Rollout CRD의 strategy만 진화

## ADR-016: Valkey는 helm 차트 대신 순수 매니페스트 (6장)
**시점**: 2026-07 / **결정**: bitnami/공식 valkey helm 차트를 버리고 valkey/valkey 순수 매니페스트로 배포한다. 비번은 GSM 파일을 `--requirepass`로 직접 읽는다
**이유**:
- 비번 단일 원천 실현: 앱과 Valkey 서버가 같은 GSM 파일을 CSI로 읽어 불일치(WRONGPASS) 원천 제거
- helm 차트 한계 회피: 차트는 비번을 K8s Secret으로만 참조 → CSI secretObjects 합성과 existingSecret 볼륨이 순환
- 랜덤 비번 방지: helm이 재배포마다 비번을 새로 생성하던 문제 제거
- 단순성: StatefulSet + Service + SecretProviderClass만으로 충분, bitnami의 configmap 3개·PDB·NetworkPolicy 등 불필요

## ADR-017: 워크로드 노드 배치는 멀티 노드풀 + nodeSelector (7장)
**시점**: 2026-07 / **결정**: Taint/Toleration·Node Affinity 대신 역할별 노드풀(api/worker/ops) + nodeSelector(GKE 자동 라벨)로 워크로드를 격리한다
**이유**:
- 관심사 분리: API(api-pool)/메시징(worker-pool)/운영도구(ops-pool)를 물리적으로 분리해 상호 영향 최소화
- 단순성: nodeSelector 한 줄이면 배치 지정 완료, Taint보다 설정이 간단
- 키 일관성: cloud.google.com/gke-nodepool 자동 라벨만 사용(커스텀 키는 라벨 부재로 Pending 유발)
- IaC 관리: Terraform node_pools 맵 확장으로 노드풀 수명주기 관리

## ADR-018: 다수 앱 관리는 App of Apps + Sync Wave (7장)
**시점**: 2026-07 / **결정**: 단일 Application·ApplicationSet 대신 App of Apps(root-app이 각 앱 application.yaml 수집) + sync-wave로 설치 순서를 제어한다
**이유**:
- 앱 단위 폴더 구조와 자연 결합: k8s/<app>/application.yaml을 부모가 include
- 설치 순서 보장: CRD(argo-rollouts, wave 0) → 플랫폼(관측, wave 1) → 앱(wave 2)로 의존성 역전 방지
- 선언형: 새 앱은 폴더 추가로 편입, 명령형 설치 불필요
- 자기참조 회피: 리소스는 manifests/·Helm, application.yaml만 수집

## ADR-019: 멀티테넌시는 Namespace 분리 + per-tenant Rollout (7장)
**시점**: 2026-07 / **결정**: 단일 namespace 라벨 격리·vCluster 대신 테넌트별 Namespace 분리와 독립 Rollout으로 멀티테넌시를 구현한다
**이유**:
- 강한 격리: 테넌트별 Namespace로 RBAC·리소스·네트워크 경계 확보
- 독립 배포: 테넌트마다 Rollout/전략을 독립적으로 운영(enterprise는 별도 canary)
- App of Apps와 결합: 테넌트 앱을 root-app이 자동 관리, CreateNamespace로 ns 자동 생성
- 공유 자원 재사용: Valkey는 cross-namespace FQDN으로 공유, 비번은 GSM 단일 원천이라 테넌트 간 불일치 없음

## ADR-020: 메시징은 Kafka(Strimzi, KRaft) (8장)
**시점**: 2026-07 / **결정**: RabbitMQ·NATS·Redis Streams 대신 Kafka(Strimzi operator, KRaft 모드)로 이벤트 드리븐을 구현한다
**이유**:
- 업계 표준: 이벤트 드리븐 아키텍처의 사실상 표준, 학습 가치가 높다
- GitOps 호환: Strimzi가 Kafka를 CRD(Kafka/KafkaNodePool/KafkaTopic)로 선언, ArgoCD로 관리
- KRaft 모드: ZooKeeper 없이 단일 노드로 운영해 리소스 절약
- 메시지 영속성: 디스크에 저장, Consumer가 죽어도 유실 없음. Redis Streams는 기능 제한, NATS는 채택률 낮음

## ADR-021: 분산 트레이싱은 Grafana Tempo (8장)
**시점**: 2026-07 / **결정**: Jaeger·Zipkin 대신 Grafana Tempo로 분산 트레이싱을 구현하고, 앱에 OpenTelemetry SDK를 붙인다
**이유**:
- Grafana 통합: 4장에서 운영 중인 Grafana에서 바로 트레이스를 조회, 별도 UI 불필요
- 3축 통합: Prometheus(메트릭)+Loki(로그)+Tempo(트레이스)를 한 대시보드에서 연결
- 경량: 단일 바이너리 모드, ops-pool에 25m로 배치
- OTLP 네이티브: OpenTelemetry 프로토콜 기본 지원, 벤더 중립

## ADR-022: 배치 자동화는 K8s CronJob (8장)
**시점**: 2026-07 / **결정**: 외부 cron·Argo Workflows 대신 K8s CronJob으로 주기 작업을 실행한다
**이유**:
- 쿠버네티스 네이티브: 별도 스케줄러 없이 클러스터가 관리
- GitOps 관리: 매니페스트로 선언, ArgoCD App of Apps에 편입
- 노드 격리: ops-pool 배치로 운영 도구를 앱과 분리
- 이력 관리: successfulJobsHistoryLimit으로 최근 실행만 유지

## ADR-023: 위험 작업은 command-guardrails/ 절차서 (8장)
**시점**: 2026-07 / **결정**: 위험 작업(Kafka Topic 삭제, CronJob 수동 실행, 테넌트 Namespace 삭제)은 command-guardrails/의 절차서(사전 확인→실행→사후 검증)를 따른다
**이유**:
- 강제(harness)와 절차(문서)의 분리: settings.local.json이 실행 차단이라면, command-guardrails는 "어떻게 안전하게 하는가"의 누적 지식
- GitOps 일관성: 절차서가 kubectl 직접 삭제 대신 Git 경유를 명시해 selfHeal 충돌 방지
- 영구 자산: 체험 후 삭제하는 settings.local.json과 달리 git에 누적, 새 위험 작업마다 추가
