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
| ch6 | 6.1 캐시 | ✅ | 2026-07-19 | Valkey standalone(Helm, requests 최소화), 인메모리 카운터→INCR 전환(v0.3.0), replicas 1 축소 선행. 단일 replica라 Pod 간 공유는 6.3 canary(stable+canary 동시 기동)에서 검증 |
| ch6 | 6.2 시크릿 관리 | ✅ | 2026-07-20 | Workload Identity(클러스터+노드풀) + Secret Manager CSI addon 활성화, valkey 비번을 Google Secret Manager에 저장, SecretProviderClass(provider=gke) 파일 마운트로 v0.3.1 배포·검증. WI principal에 secretAccessor 직접 바인딩(GCP SA 미생성) |
| ch6 | 6.3 Canary 전환 | ✅ | 2026-07-20 | rollout.yaml B/G→canary(20/50/80%), git push→Rollout 삭제→ArgoCD 재적용 순서. v0.3.2로 step 1→3→5→6 점진 승격 e2e 검증 |
| ch6 | 6.4 아키텍처 스냅샷 | ✅ | 2026-07-20 | claude-context/architecture.md 신설(3층 지식구조·토폴로지·컴포넌트·파이프라인·GitOps·관측·ns 6+섹션) |
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
| 캐시 (ch6.1) | Valkey | Redis, Memcached | 책 기본 흐름. Redis 라이선스 변경 이후 오픈소스(BSD) 포크, Redis 프로토콜 호환, INCR로 전역 순차 ID |
| 시크릿 관리 (ch6.2) | Google Secret Manager + GKE managed CSI | K8s Secret, 오픈소스 CSI | 책 기본 흐름. 시크릿을 클러스터 밖 GSM에 저장, WI로 keyless 접근, CSI 파일 마운트. GKE managed는 addon 한 줄로 활성화(오픈소스 helm 설치 불필요) |
| GitOps 관리 구조 (재편) | App of Apps + 앱별 폴더 | 단일 Application, ApplicationSet | 사용자 규칙. k8s/&lt;app&gt;/에 application.yaml + (manifests/ 또는 Helm values.yaml), root-app이 각 application.yaml만 include. 명령형 설치(helm/kubectl) 전부 선언형(ArgoCD)으로 이관 |
| 배포 전략 전환 (ch6.3) | Argo Rollouts Canary | Blue/Green | 5장 B/G 경험 후 같은 Rollout CRD에서 strategy만 canary로 교체. 20→50→80% 점진 전환으로 위험도 최소화, 도구 변경 없이 전략만 진화 |
| Valkey 배포 방식 (ch6.3 재작성) | 순수 매니페스트 (valkey/valkey:9.1) | bitnami helm 차트 | helm 차트의 랜덤 비번 재생성·existingSecret 볼륨 순환 회피. GSM 파일을 --requirepass로 직접 읽어 앱·서버 비번 단일 원천(GSM) 통일 |

## Terraform 인프라 (IaC)

| 항목 | 값 |
|------|-----|
| 코드 위치 | `terraform/gcp/gke/` (클러스터), `terraform/gcp/apps/` (앱 리소스) |
| State | GCS backend (private 버킷, versioning 활성화). prefix 폴더별 분리 |
| 리소스 정의 방식 | `03-locals.tf`의 map + `for_each` (gke: `gke_definitions`, apps: `app_definitions`) |
| 공통 라벨 | `project=notiflex`, `managed-by=terraform` (전 폴더 통일) |
| 관리 리소스 | `google_container_cluster`, `google_container_node_pool`, `google_artifact_registry_repository` |
| ch6.2 반영 (2026-07-20) | gke 클러스터에 `workload_identity_config`, `secret_manager_config` + 노드풀 `workload_metadata_config(GKE_METADATA)`, `node_count 2→3`. gcloud 수동 변경분을 IaC로 정합화(`plan` no-change 확인). **6.2 클러스터 변경은 gcloud로 먼저 적용 후 코드 반영했으므로, 재현 시엔 코드→apply 순서 권장** |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Terraform | 1.15.7 | tfenv 고정 |
| google provider | 7.39.0 | static 고정 |
| GKE (master) | 1.35.5-gke.1241004 | |
| Go | 1.25 | go.mod + golang:1.25-alpine |
| Notiflex 이미지 | sha-dbc2ea9 (app v0.3.2) | 3.5부터 CI가 git SHA 태그 자동 부여. Canary로 승격된 최신 버전 |
| ArgoCD | v3.4.5 | stable manifest 설치 (2026-07-12). App of Apps 재편 후 root-app + 자식 6개 관리 |
| Argo Rollouts | v1.9.1 (chart 2.41.1) | 2026-07-20 App of Apps 재편으로 helm chart(argo/argo-rollouts) 기반 ArgoCD 관리로 전환 |
| Valkey | valkey/valkey:9.1-alpine | ch6.3에서 bitnami helm→순수 매니페스트 재작성. GSM 파일 --requirepass, ArgoCD 관리 |
| 관측 스택 | kube-prometheus-stack 87.15.1 / loki 7.0.0 / fluent-bit 2.6.0 | App of Apps 재편으로 ArgoCD Helm source 관리, values는 각 앱 폴더로 이동 |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | e2-medium (Spot) | 3 | notiflex-api (smb, replicas 1), valkey-primary-0, 관측 스택, CSI/WI DaemonSet |

- 클러스터: `notiflex-cluster` (asia-northeast3-a, Zonal, Public)
- Gateway API: `CHANNEL_STANDARD` 활성화 / Workload Identity·Secret Manager CSI 활성화(ch6.2)
- 노드 3대로 증설(2026-07-20): ch6.2에서 WI 메타데이터 서버(100m/노드)+CSI DaemonSet(120m/노드)이 추가되며 B/G 파드가 CPU 부족으로 Pending → 노드 1대 추가로 해소. Terraform node_count 3으로 정합화
- kubectl 컨텍스트: `gke-sysnet4admin_book_gitaiops`
- gcloud 실습 계정: named config `book-gitaiops`(account=jinnypark9393cc@gmail.com). 실습 명령은 `CLOUDSDK_ACTIVE_CONFIG_NAME=book-gitaiops` + `unset GOOGLE_APPLICATION_CREDENTIALS`

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
| 6.1 | Valkey 앱 기동 시 Valkey Service DNS/기동 지연으로 3~4회 `context deadline exceeded` | 앱에 연결 재시도 로직(10회/3초) 넣어 해결. 재시도 없으면 CrashLoopBackOff. 가드레일 예고와 일치 |
| 6.2 | WI/CSI 활성화 후 노드당 CPU 220m 추가 소모(WI 메타데이터 서버 100m + CSI DaemonSet 120m)로 B/G 새 파드가 13시간 Pending(`Insufficient cpu`) | 관측 스택 requests 5m/1m 축소만으론 부족. **노드 3대로 증설**(Terraform node_count 3 정합화)해 해소. 가드레일은 "Loki/FluentBit 임시 제거"를 제시하나 이미 1m라 효과 없음 → 실측상 노드 추가가 정답 |
| 6.2 | scratch 이미지라 `kubectl exec ... ls /mnt/secrets`가 `executable file not found` | 셸/바이너리 없는 scratch 특성. CSI 마운트 검증은 exec 대신 앱 동작(/id가 Valkey 연결 성공 = 비번 파일 정상 읽음)으로 확인 |
| GitOps 재편 | App of Apps 전환 시 valkey Application의 helm releaseName 기본값(`notiflex-valkey`)이 기존 수동 릴리스(`valkey`)와 달라 adopt 안 되고 중복 StatefulSet 생성 | Application에 `helm.releaseName: valkey` 명시해 기존 `valkey-primary` 리소스명과 일치시켜 adopt. hard refresh 후 중복 `notiflex-valkey-*` prune됨 |
| GitOps 재편 | App of Apps 전환 시 옛 `notiflex-smb` Application이 옛 경로(k8s/smb)를 가리켜 prune 위험 | `kubectl delete application notiflex-smb --cascade=orphan`으로 Application만 제거(리소스 보존) 후 root-app apply로 새 자식이 adopt |
| GitOps 재편 | rollout.yaml을 k8s/smb/manifests로 이동 후 CI가 옛 경로 `k8s/smb/rollout.yaml`을 sed하려다 실패 위험 | ci.yaml의 4개 참조를 `k8s/smb/manifests/rollout.yaml`로 전역 치환. workflow 파일이라 `workflow` 스코프 PAT(store credential)로 push |
| 6.3 | App of Apps adopt로 valkey helm이 비번을 랜덤 재생성 → GSM엔 옛 비번 남아 앱이 WRONGPASS로 CrashLoop | (1차) 현재 Valkey 실제 비번을 GSM 새 버전으로 동기화 + 파드 재시작으로 급한 불. (근본) bitnami helm을 버리고 순수 매니페스트로 재작성 — valkey가 GSM 파일을 --requirepass로 직접 읽어 앱과 단일 원천 공유 |
| 6.3 | bitnami/공식 valkey helm 차트 모두 existingSecret을 볼륨 직접마운트 → CSI secretObjects 합성 Secret과 순환(Secret 없어서 마운트 실패) | 차트로는 GSM 파일 직접 읽기 불가(차트가 command를 안 열어줌). 순수 매니페스트에서 StatefulSet command를 직접 짜 `--requirepass "$(cat /mnt/gsm-secrets/valkey-password)"`로 해결 |
| 6.3 | valkey 순수 매니페스트 apply 시 `StatefulSet spec Forbidden: updates to ... forbidden`(불변필드) | 옛 bitnami StatefulSet selector/volumeClaimTemplates가 불변이라 patch 불가. `kubectl delete statefulset valkey-primary`로 삭제 후 ArgoCD가 새 매니페스트로 재생성 |
| 6.3 | 새 valkey KSA(`valkey`)가 GSM 접근 시 `secretmanager.versions.access denied` | 6.2에서 부여한 대상은 `default`/`valkey-primary` KSA뿐. 새 StatefulSet의 `valkey` KSA에도 WI principal로 secretAccessor 바인딩 필요 |
