# Notiflex 여정 기록

이 파일은 실제로 진행한 내용을 기록한다. 각 챕터 완료 시 업데이트한다.

> 이 실습은 책 기본 흐름과 두 가지가 다르다:
> - 작업 저장소: `notiflex-platform` → **`notiflex-platform-test`** (public, `jinnypark9393` 계정)
> - 인프라 프로비저닝: gcloud CLI → **Terraform (IaC, GCS backend)**

## 진행 현황

| 챕터 | 서브챕터 | 상태 | 완료일 | 비고 |
|------|---------|------|--------|------|
| ch2 | 2.2 설치 확인 | ✅ | 2026-07-05 | gcloud/kubectl/terraform/tfenv/direnv/gh 기설치 확인 |
| ch2 | 2.3 gcloud 설정 | ✅ | 2026-07-05 | 개인 계정 분리, 실습 프로젝트/리전 설정, 필요 API 활성화 |
| ch2 | 2.4 GitHub 저장소 | ✅ | 2026-07-05 | notiflex-platform-test public repo + CLAUDE.md + 구조 |
| ch2 | 2.5 GKE 클러스터 | ✅ | 2026-07-05 | **Terraform으로** 생성 (Zonal, Spot, Gateway API) |
| ch2 | 2.6 빌드/배포 | ✅ | 2026-07-05 | Go 앱, Cloud Build, AR(Terraform), K8s 배포(Pod 2개 Running) |
| ch2 | 2.7 첫 커밋 | ✅ | 2026-07-05 | 초기 구성~빌드·배포 커밋 3건 origin/main 푸시 완료 |
| ch3 | 3.2 GitOps 도구 | ✅ | 2026-07-12 | ArgoCD v3.4.5 설치, notiflex-smb App Synced/Healthy (auto-sync+prune+selfHeal) |
| ch3 | 3.3 기능 추가 | ✅ | 2026-07-12 | /version 추가, v0.1.1 롤링 업데이트 + git revert 롤백 테스트 완료 |
| ch3 | 3.4 CI | ✅ | 2026-07-12 | GitHub Actions + **WIF(OIDC) keyless** 인증, sha 태그 AR 푸시 검증 |
| ch3 | 3.5 CI-CD 연결 | ✅ | 2026-07-12 | CI가 매니페스트 sha 태그 자동 커밋 → ArgoCD 자동 배포, e2e 검증(v0.1.2) |
| ch3 | 3.6 CLAUDE.md 규칙 | ✅ | 2026-07-12 | 규칙 추가→delete 시나리오→selfHeal 17초 복원 확인→규칙 되돌림. 런타임 분류기도 delete 차단(3중 방어 확인) |
| ch4 | 4.2 메트릭 모니터링 | ✅ | 2026-07-12 | kube-prometheus-stack(Helm), Notiflex 대시보드 ConfigMap 등록, 타겟 16개 수집 |
| ch4 | 4.3 로그 수집 | ✅ | 2026-07-13 | Loki(SingleBinary) + Fluent Bit(DaemonSet×2), Grafana 데이터소스 등록, notiflex 로그 조회 확인 |
| ch4 | 4.4 알림 | ✅ | 2026-07-13 | PodRestartTooMany PrometheusRule 로드 확인. 단, 가이드의 테스트(파드 삭제)로는 발화 불가 — 트러블슈팅 참조 |
| ch5 | 5.2 트래픽 관리 | ✅ | 2026-07-19 | Gateway API(regional external) + HTTPRoute + HealthCheckPolicy, 외부 IP 35.216.101.141에서 /health·/id 검증 |
| ch5 | 5.3 무중단 배포 | ✅ | 2026-07-19 | Argo Rollouts v1.9.1, Deployment→Rollout(B/G) 전환, v0.2.0 배포로 preview→30초 auto-promote e2e 검증 |
| ch5 | 5.4 ADR | ✅ | 2026-07-19 | docs/architecture-decisions.md 신설, 도구 선택 기록을 ADR-001~012로 변환 |
| ch6 | 6.1 캐시 | ⬜ | | |
| ch6 | 6.2 시크릿 관리 | ⬜ | | |
| ch6 | 6.3 Canary 전환 | ⬜ | | |
| ch7 | 7.2 멀티 노드풀 | ⬜ | | |
| ch7 | 7.3 App of Apps | ⬜ | | |
| ch7 | 7.4 멀티테넌시 | ⬜ | | |
| ch8 | 8.1 메시징 | ⬜ | | |
| ch8 | 8.2 트레이싱 | ⬜ | | |
| ch8 | 8.3 CronJob | ⬜ | | |
| ch9 | 9.1 저장소 분석 | ⬜ | | |
| ch9 | 9.2 회고 | ⬜ | | |
| ch9 | 9.3 온보딩 문서 | ⬜ | | |
| ch9 | 9.4 GitAIOps 분석 | ⬜ | | |
| ch9 | 9.5 마무리 | ⬜ | | |

## 도구 선택 기록

3-프롬프트 패턴(탐색→비교→실행)에서 실제로 선택한 도구와 이유를 기록한다.

| 영역 | 선택 | 검토한 대안 | 선택 이유 |
|------|------|-----------|----------|
| 인프라 프로비저닝 (ch2.5) | Terraform | gcloud CLI | IaC 버전관리, 재현성, plan 검토, 삭제 용이 |
| Terraform state (ch2.5) | GCS backend | local state, bootstrap 모듈 | 버전관리 버킷, 팀 공유 표준. 버킷은 gcloud로 부트스트랩 |
| Terraform 버전 관리 (ch2.5) | tfenv + `.terraform-version` | 시스템 전역 설치 | 프로젝트별 버전 고정, 팀원 일치 |
| Terraform 변수 주입 (ch2.5) | direnv `.envrc` (`TF_VAR_*`) | tfvars 파일 | 환경변수 주입, 로컬 값 분리 |
| 이미지 빌드 (ch2.6) | Cloud Build (`gcloud builds submit`) | 로컬 Docker buildx | 원격 빌드, 로컬 Docker 불필요, 크로스컴파일 불필요 |
| 이미지 저장소 (ch2.6) | Artifact Registry (Terraform) | Docker Hub, GCR | GKE 네이티브, IaC 관리, 리전 로컬 |
| GitOps 도구 (ch3.2) | ArgoCD | Flux | 책 기본 흐름. 선언적 GitOps + UI 제공, App of Apps(7장) 확장 대비 |
| CI GCP 인증 (ch3.4) | WIF (GitHub OIDC, keyless) | SA 키 JSON (책 방식 A) | 장기 크레덴셜 미보관, 로테이션 불필요. AWS의 IAM Role+OIDC와 동일 패턴. Terraform으로 pool/provider/binding IaC 관리 |
| 메트릭 (ch4.2) | kube-prometheus-stack (Helm) | Datadog 등 SaaS | 책 기본 흐름. 50+ 리소스를 차트 하나로, ServiceMonitor/Rule 자동 연결. requests는 values로 축소 (ch6 전 재축소 예정) |
| 로그 (ch4.3) | Loki + Fluent Bit | ELK | 책 기본 흐름. 라벨 인덱싱으로 경량, Grafana 통합. 최신 차트의 캐시/카나리/게이트웨이는 리소스 예산 때문에 비활성화 |
| 외부 트래픽 (ch5.2) | Gateway API (GKE managed) | Ingress(NGINX 등) | 책 기본 흐름. GKE 네이티브 L7 LB, 역할 분리된 표준 리소스, 5.3 Argo Rollouts 트래픽 제어 확장 대비 |
| 배포 전략 도구 (ch5.3) | Argo Rollouts | Flagger, Istio | 책 기본 흐름. ArgoCD와 같은 Argo 생태계, Rollout CRD로 B/G→Canary 전환 용이 |

## Terraform 인프라 (IaC)

| 항목 | 값 |
|------|-----|
| 코드 위치 | `terraform/gcp/gke/` (클러스터), `terraform/gcp/apps/` (앱 리소스) |
| State | GCS backend (private 버킷, versioning 활성화). prefix 폴더별 분리 |
| 리소스 정의 방식 | `03-locals.tf`의 map + `for_each` (gke: `gke_definitions`, apps: `app_definitions`) |
| 공통 라벨 | `project=notiflex`, `managed-by=terraform` (전 폴더 통일) |
| 관리 리소스 | `google_container_cluster`, `google_container_node_pool`, `google_artifact_registry_repository` |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Terraform | 1.15.7 | tfenv 고정 |
| google provider | 7.39.0 | static 고정 |
| GKE (master) | 1.35.5-gke.1241004 | |
| Go | 1.25 | go.mod + golang:1.25-alpine |
| Notiflex 이미지 | sha-75efccb (app v0.2.0) | 3.5부터 CI가 git SHA 태그 자동 부여. 앱 내부 version 상수는 v0.2.0 |
| ArgoCD | v3.4.5 | stable manifest 설치 (2026-07-12) |
| Argo Rollouts | v1.9.1 | latest manifest 설치 (2026-07-19) |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | e2-medium (Spot) | 2 | notiflex-api (smb, replicas 2) |

- 클러스터: `notiflex-cluster` (asia-northeast3-a, Zonal, Public)
- Gateway API: `CHANNEL_STANDARD` 활성화
- kubectl 컨텍스트: `gke-sysnet4admin_book_gitaiops`

## 트러블슈팅 이력

겪은 문제와 해결 방법을 기록한다. 같은 문제를 다시 겪지 않도록 한다.

| 챕터 | 문제 | 해결 |
|------|------|------|
| 2.5 | terraform init/apply 시 GCS backend 403 (회사 SA로 접근) | `~/.zshrc`의 `GOOGLE_APPLICATION_CREDENTIALS`(회사 SA 키)가 ADC보다 우선. terraform 명령 앞에 `unset GOOGLE_APPLICATION_CREDENTIALS &&` 필수 |
| 2.5 | ADC가 회사 계정 기반이라 개인 프로젝트 접근 불가 | 개인 계정으로 `gcloud auth application-default login` 수행 (quota project 자동 정렬) |
| 2.5 | Gateway API를 처음 locals에서 누락 | apply 전 `gateway_api_channel = "CHANNEL_STANDARD"` 추가하여 처음부터 켜진 채로 생성 |
| 2.5 | kubectl `gke-gcloud-auth-plugin not found` | `gcloud components install gke-gcloud-auth-plugin`로 설치 |
| 2.6 | `gcloud builds submit` 403 (compute SA가 소스 버킷 접근 불가) | 신규 프로젝트는 기본 compute SA에 권한 없음. `roles/cloudbuild.builds.builder` 부여 |
| 2.6 | gke에 공통 라벨 추가 후 apply 시 노드풀 재생성 | GCE 인스턴스 라벨은 노드 재생성 필요. 워크로드 배포 전이라 무해 (Spot이라 원래 교체 가능) |
| 4.4 | 가이드의 알림 테스트 `kubectl delete pod -l app=notiflex-api`로는 PodRestartTooMany가 발화하지 않음 (실측: 삭제 → 새 파드 RESTARTS 0, `kube_pod_container_status_restarts_total` 미증가, 90초 후에도 inactive) | 파드 삭제는 '재생성'이지 '컨테이너 재시작'이 아님. 룰을 발화시키려면 컨테이너 크래시(liveness 실패, 프로세스 종료)가 필요. 룰 자체는 정상 로드 확인 |
| 3.5 | 가이드는 "manifest push 403 방지에 repo 레벨 Workflow permissions도 write 필수"라 하나, 실측 결과 **repo 기본값 read 유지 + ci.yaml `permissions: contents: write` 명시만으로 push 성공** | 워크플로우 레벨 permissions가 repo 기본값을 덮어씀 (GitHub 문서와 일치). repo 설정 변경 불필요 — 최소권한 유지 |
| 5.2 | 가이드는 "proxy-only 서브넷을 GKE가 자동 생성"이라 하나, 실측 결과 자동 생성되지 않고 Gateway SYNC 이벤트에 `An active proxy-only subnetwork is required` 에러 발생 | `gcloud compute networks subnets create proxy-only-subnet --purpose=REGIONAL_MANAGED_PROXY --role=ACTIVE --range=172.16.0.0/23`으로 수동 생성 후 1~2분 내 IP 할당·Programmed=True |
| 3.3 | ArgoCD가 새 커밋을 수 분간 감지 못함 (`sync.revision`이 이전 커밋에 고정, 폴링 3분 경과 후에도 미갱신) | 가이드의 트러블슈팅은 NetworkPolicy egress 차단을 지목하나, 실제 repo-server NP는 **Ingress 전용**이라 무관. `kubectl annotate application notiflex-smb -n argocd argocd.argoproj.io/refresh=hard --overwrite`로 즉시 refresh하면 해결 (NP 삭제 불필요) |
