# Notiflex 아키텍처 스냅샷

> 현재 시점(2026-07-20)의 아키텍처 한눈 보기. AI가 매 대화에서 전체 그림을 빠르게 잡기 위한 단일 페이지 요약이다.

## 3층 지식 구조

이 저장소는 AI 컨텍스트를 세 층으로 분리해 관리한다. 세 층이 섞이지 않아야 메타데이터·현재 그림·과거 결정이 혼동되지 않는다.

- **CLAUDE.md** — 프로젝트 메타데이터와 행동 규칙(GitOps 구조 규칙, 크레덴셜 금지 등). 매 대화 자동 로드.
- **claude-context/** (이 문서) — 지금 어떻게 동작하는가의 아키텍처 스냅샷. 자동 참조용 현재 상태 요약.
- **docs/architecture-decisions.md** (ADR) — 왜 이 결정을 내렸는가의 누적 기록. 사람과 AI가 함께 검토.

## 클러스터 토폴로지

| 항목 | 값 |
|------|-----|
| 클러스터 | `notiflex-cluster` (GKE Standard, Zonal) |
| 리전/존 | `asia-northeast3` / `asia-northeast3-a` |
| 노드풀 | `default-pool` — e2-medium (Spot) × 3 |
| K8s 버전 | v1.35.5-gke.1241004 |
| 활성 기능 | Gateway API(CHANNEL_STANDARD), Workload Identity, Secret Manager CSI addon |
| IaC | Terraform (`terraform/gcp/gke`), GCS backend, node_count·WI·CSI 모두 코드 반영 |

## 컴포넌트 다이어그램

```
[인터넷]
   │
   ▼
[Gateway API]  notiflex-gateway (gke-l7-regional-external-managed, 35.216.101.141)
   │  HTTPRoute: / → notiflex-api
   ▼
[Service] notiflex-api (stable) / notiflex-api-preview (canary)   ← Argo Rollouts가 트래픽 대상 전환
   │
   ▼
[Rollout] notiflex-api (Canary 전략, 20→50→80→100%)
   │
   ▼
[Pod] notiflex-api (Go, scratch 이미지)
   │           │
   │           └──(CSI 파일 /mnt/secrets/valkey-password)──▶ [Google Secret Manager] valkey-password
   ▼
[Valkey] valkey-primary (순수 매니페스트, GSM 파일로 --requirepass)
              └──(CSI 파일 /mnt/gsm-secrets/valkey-password)──▶ 같은 GSM secret

시크릿 단일 원천: notiflex-api와 valkey 모두 같은 GSM valkey-password를 CSI 파일로 읽는다.
```

## 배포 파이프라인

```
개발자 git push (app/**)
   │
   ▼
GitHub Actions CI  (.github/workflows/ci.yaml)
   │  WIF(OIDC) keyless 인증 → 이미지 빌드 → Artifact Registry 푸시(sha-<7> 태그)
   │  → k8s/smb/manifests/rollout.yaml 이미지 태그 자동 커밋 [skip ci]
   ▼
Artifact Registry  asia-northeast3-docker.pkg.dev/.../notiflex/api
   │
   ▼
ArgoCD (App of Apps)  root-app → 자식 6개 Application
   │  notiflex-smb가 rollout.yaml 변경 감지 → sync
   ▼
Argo Rollouts  Canary 20→50→80→100% 점진 배포
```

## GitOps 구조 (App of Apps)

| Application | source | 대상 ns | 상태 |
|------------|--------|---------|------|
| notiflex-root | k8s (application.yaml만 include) | argocd | Synced/Healthy |
| notiflex-smb | k8s/smb/manifests (순수 매니페스트) | notiflex | Synced/Healthy |
| notiflex-valkey | k8s/valkey/manifests (순수 매니페스트) | notiflex | Synced/Healthy |
| notiflex-argo-rollouts | argo/argo-rollouts Helm 2.41.1 | argo-rollouts | Synced/Healthy |
| notiflex-kube-prometheus | kube-prometheus-stack Helm 87.15.1 + CR | monitoring | Synced/Healthy |
| notiflex-loki | grafana/loki Helm 7.0.0 | monitoring | Synced/Healthy |
| notiflex-fluent-bit | grafana/fluent-bit Helm 2.6.0 | monitoring | Synced/Healthy |

## 관측 가능성

| 도구 | 역할 |
|------|------|
| Prometheus (kube-prometheus-stack) | 메트릭 수집·저장, PrometheusRule 알림 룰(PodRestartTooMany) |
| Grafana | 메트릭/로그 통합 대시보드 (Notiflex Overview) |
| Loki | 로그 저장 (라벨 인덱싱, SingleBinary) |
| Fluent Bit | 노드별 컨테이너 로그 수집 DaemonSet → Loki |
| Tempo | 트레이싱 (8장 예정, 미도입) |

## 주요 네임스페이스

| 네임스페이스 | 주요 워크로드 |
|-------------|-------------|
| notiflex | notiflex-api (Rollout, Canary), valkey-primary (StatefulSet), Gateway/HTTPRoute, SecretProviderClass |
| monitoring | Prometheus, Grafana, Alertmanager, Loki, Fluent Bit |
| argo-rollouts | Argo Rollouts 컨트롤러 |
| argocd | ArgoCD (root-app + 자식 6개 Application) |
