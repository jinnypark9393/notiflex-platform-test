# CLAUDE.md — Notiflex Platform

이 파일은 Claude Code에게 이 저장소의 컨텍스트와 행동 규칙을 제공한다.

## 프로젝트 개요

**Notiflex** — B2B 알림 SaaS 플랫폼. 기업 고객에게 다양한 채널로 알림을 발송하는 서비스이다.

## 기술 스택

- **언어**: Go (표준 라이브러리만 사용, 외부 웹 프레임워크 없음)
- **컨테이너**: scratch 베이스 이미지 (최소 크기, 공격 표면 최소화)
- **인프라**: GKE Standard (Zonal), Spot VM
- **CI**: GitHub Actions (`.github/workflows/`) — 빌드·테스트·이미지 푸시를 GitHub Actions로 실행한다
- **GitOps**: ArgoCD (App of Apps — `k8s/` 아래 앱별 Application, root-app이 관리)
- **캐시**: Valkey (Pod 간 ID 카운터 공유, ch6.1)
- **시크릿**: Google Secret Manager + GKE managed CSI, Workload Identity (ch6.2)
- **관측 가능성**: Prometheus, Grafana, Loki, Fluent Bit (Tempo는 8장 예정)
- **배포 전략**: Rolling → Blue/Green(현재) → Canary(ch6.3 예정)

## GCP 설정

| 항목 | 값 |
|------|-----|
| 프로젝트 ID | `project-fea698e1-5762-48a2-918` |
| 리전 | `asia-northeast3` (서울) |
| 존 | `asia-northeast3-a` |
| 계정 | `jinnypark9393cc@gmail.com` (실습 전용 개인 계정) |

### Artifact Registry

```
asia-northeast3-docker.pkg.dev/project-fea698e1-5762-48a2-918/notiflex
```

## 행동 규칙

1. **리소스 생성·삭제 시 사용자 확인 필수**: 클러스터, 노드풀, 네임스페이스, Deployment, LoadBalancer, 저장소 등 **모든 리소스의 생성·삭제 작업은 실행 전 반드시 사용자에게 확인을 받는다.** 확인 없이 생성하거나 삭제하지 않는다.
2. **크레덴셜 하드코딩·출력 금지**: 토큰, API 키, 비밀번호, 인증서 등 크레덴셜 정보는 **절대 코드·매니페스트·문서에 하드코딩하지 않으며, 프롬프트(응답)에도 출력하지 않는다.** 시크릿은 Secret Manager / K8s Secret / GitHub Secrets 등 전용 저장소로만 관리하고, 매니페스트에는 참조(reference)만 남긴다.
3. **항상 확인 후 실행**: 인프라 변경(클러스터, 배포, 삭제)은 실행 전 현재 상태를 먼저 확인한다.
4. **변경 전 현재 상태 확인**: `kubectl get`, `gcloud ... describe` 등으로 대상의 현재 상태를 파악한 뒤 변경한다.
5. **kubectl 컨텍스트 명시**: 잘못된 클러스터를 건드리지 않도록 kubectl 명령에는 실습 클러스터 컨텍스트를 명시한다.
6. **비용 인지**: GKE 노드, LoadBalancer 등 비용이 발생하는 리소스는 생성 전 안내한다.
7. **Terraform 실행 시 .envrc 활성화 필수**: 셸 상태가 명령 간 유지되지 않으므로, `.envrc`(direnv) 활성화와 terraform 명령을 **반드시 `&&`로 한 줄에** 이어서 실행한다. 별도 줄로 source하면 `TF_VAR_*`가 적용되지 않는다.
   ```bash
   # ✅ 올바른 사용 (한 줄)
   cd terraform/gcp/gke && source ./.envrc && terraform plan
   # ❌ 금지 — .envrc가 다음 명령에 적용 안 됨
   source ./.envrc
   terraform plan
   ```
8. **GOOGLE_APPLICATION_CREDENTIALS 무력화**: `~/.zshrc`가 회사 SA 키(`gcp-devops.json`)를 `GOOGLE_APPLICATION_CREDENTIALS`로 export한다. 이 값이 ADC보다 우선하므로, terraform 등 GCP 접근 명령은 **`unset GOOGLE_APPLICATION_CREDENTIALS &&`를 맨 앞에 붙여** 실습 계정 ADC를 쓰게 한다. (`~/.zshrc`는 회사 설정이므로 수정하지 않는다)
   ```bash
   cd terraform/gcp/gke && unset GOOGLE_APPLICATION_CREDENTIALS && source ./.envrc && terraform plan
   ```

## 디렉터리 구조

```
notiflex-platform-test/
├── CLAUDE.md          # 이 파일 — 프로젝트 컨텍스트
├── app/               # Go 애플리케이션
├── k8s/               # GitOps 리소스 (App of Apps — 아래 규칙 참조)
│   ├── root-app.yaml  # 부모 App of Apps (각 앱 application.yaml만 include)
│   ├── smb/           # notiflex-api (application.yaml + manifests/)
│   ├── valkey/        # Valkey 캐시 (application.yaml + values.yaml, Helm source)
│   ├── argo-rollouts/ # Argo Rollouts 컨트롤러 (application.yaml + values.yaml)
│   ├── kube-prometheus/ # Prometheus 스택 + 대시보드/알림 CR (multi-source)
│   ├── loki/          # Loki (application.yaml + values.yaml)
│   └── fluent-bit/    # Fluent Bit (application.yaml + values.yaml)
├── terraform/
│   └── gcp/
│       ├── gke/       # GKE 클러스터 (IaC)
│       └── apps/      # 앱 배포용 GCP 리소스 (Artifact Registry 등)
└── .github/
    └── workflows/     # CI 파이프라인 (GitHub Actions)
```

## GitOps 구조 규칙 (k8s/) — 필수 준수

이 저장소의 모든 Kubernetes 리소스는 **App of Apps 패턴**으로 ArgoCD가 관리한다. 명령형 설치(`kubectl apply`, `helm install`)로 클러스터를 직접 바꾸지 않는다.

1. **앱 단위 폴더**: `k8s/` 하위는 앱마다 폴더 하나(`k8s/<app>/`). 여러 앱을 한 폴더에 섞지 않는다.
2. **각 앱 폴더에 `application.yaml`**: 그 앱의 ArgoCD Application CR을 앱 폴더 안에 둔다. Application 이름은 `notiflex-<app>`.
3. **source 종류에 따라**:
   - **순수 매니페스트 앱**(smb 등): 리소스를 `k8s/<app>/manifests/`에 두고, `application.yaml`의 `source.path`는 `k8s/<app>/manifests`를 가리킨다. (application.yaml과 리소스를 분리해 자기참조 방지)
   - **Helm 차트 앱**(valkey, argo-rollouts, loki, fluent-bit, kube-prometheus): `k8s/<app>/values.yaml`을 두고, `application.yaml`은 multi-source로 Git repo(ref: values) + 원격 Helm 차트를 조합해 `$values/k8s/<app>/values.yaml`을 참조한다.
4. **부모 App of Apps** (`k8s/root-app.yaml`): `source.path: k8s` + `directory.recurse: true` + `include: '{앱들}/application.yaml'`로 **각 앱의 application.yaml만** 수집한다. manifests/·values.yaml·root-app.yaml 자신은 include하지 않는다 (자기참조 방지).
5. **namespace는 매니페스트/Application에서 지정**: 앱마다 대상 namespace가 다르므로(smb→notiflex, monitoring 계열→monitoring, argo-rollouts→argo-rollouts) 각 application.yaml의 `destination.namespace`에 명시하고 필요 시 `CreateNamespace=true`.
6. **Helm 릴리스 adopt**: 기존 수동 helm 릴리스를 ArgoCD로 이관할 때는 `application.yaml`의 `helm.releaseName`을 기존 릴리스명과 일치시켜 리소스명 충돌 없이 adopt한다.
7. **CI 경로**: `.github/workflows/ci.yaml`의 이미지 태그 sed 대상은 `k8s/smb/manifests/rollout.yaml`. rollout 위치가 바뀌면 CI 참조도 함께 갱신한다.

## Terraform (terraform/gcp/)

- **폴더 분리**: GCP 리소스는 `terraform/gcp/` 아래 서비스 단위 폴더로 분리한다 (`gke/`, `apps/` 등). 각 폴더는 독립 state(backend prefix로 구분).
- **State**: GCS backend (private 버킷). prefix는 폴더별로 분리 (`gcp/gke/...`, `gcp/apps`).
- **버전 고정**: `.terraform-version`(tfenv) + `01-provider.tf`에 Terraform·provider 버전 static 고정.
- **변수 주입**: `.envrc`(direnv)의 `TF_VAR_*` 환경변수 — 실행 시 `&& source ./.envrc &&`로 한 줄에 이어서.
- **리소스 정의**: 리소스 레벨 값은 `03-locals.tf`의 map으로 정의하고 `for_each`로 순회한다 (gke는 `gke_definitions`, apps는 `app_definitions`). 리소스 이름은 map의 key를 사용한다.
- **공통 라벨**: 모든 GCP 리소스에 `03-locals.tf`의 `common_labels`(`project = "notiflex"`, `managed-by = "terraform"`)를 붙인다. 폴더 간 동일하게 유지한다. (provider가 자동으로 붙이는 `goog-terraform-provisioned=true`는 별개)
- **작업 습관**: 코드 작성 후 `terraform fmt`, `apply` 전 `terraform validate` + `plan` 검토.
