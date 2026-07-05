# Notiflex Platform

**Notiflex** — B2B 알림 SaaS 플랫폼. 기업 고객에게 다양한 채널로 알림을 발송하는 서비스입니다.

이 저장소는 GKE 위에서 Notiflex를 처음부터 구축하는 인프라·애플리케이션 코드를 담고 있습니다.

## 기술 스택

| 영역 | 사용 기술 |
|------|-----------|
| 언어 | Go (표준 라이브러리) |
| 컨테이너 | scratch 베이스 이미지 |
| 인프라 | GKE Standard (Zonal), Spot VM |
| IaC | Terraform (GCS backend, tfenv·direnv) |
| CI | GitHub Actions |
| GitOps | ArgoCD *(예정)* |
| 관측 가능성 | Prometheus, Grafana, Loki, Fluent Bit, Tempo *(예정)* |
| 배포 전략 | Rolling → Blue/Green → Canary *(점진 진화 예정)* |

## 디렉터리 구조

```
notiflex-platform-test/
├── CLAUDE.md          # 프로젝트 컨텍스트 + 행동 규칙
├── app/               # Go 애플리케이션
├── k8s/
│   └── smb/           # Kubernetes 매니페스트
├── terraform/
│   └── gcp/
│       └── gke/       # GKE 클러스터 (IaC)
└── .github/
    └── workflows/     # CI 파이프라인 (GitHub Actions)
```

## 인프라 (Terraform)

GKE 클러스터는 `terraform/gcp/gke/`에서 Terraform으로 관리합니다.

- **State**: GCS backend에 저장 (버전관리 활성화)
- **버전 고정**: `.terraform-version`(tfenv) + `01-provider.tf`에 Terraform·provider 버전 고정
- **변수 주입**: `.envrc`(direnv)의 `TF_VAR_*` 환경변수
- **리소스 정의**: `03-locals.tf`의 `gke_definitions` map을 `for_each`로 순회

```bash
# 최초 1회: direnv 승인
cd terraform/gcp/gke && direnv allow

# 실행 (.envrc 활성화와 한 줄로 이어서 실행)
cd terraform/gcp/gke && source ./.envrc && terraform init
cd terraform/gcp/gke && source ./.envrc && terraform plan
```

> ℹ️ GCS backend 접근에는 Application Default Credentials(ADC)가 필요합니다.
> `gcloud auth application-default login`으로 실습 계정의 ADC를 설정하세요.

## 시작하기

인프라와 애플리케이션은 챕터를 진행하면서 단계적으로 구축됩니다.

- **CI/CD**: 코드 push 시 GitHub Actions가 빌드·테스트·이미지 푸시를 수행합니다.
- **배포**: ArgoCD가 GitOps 방식으로 클러스터에 반영합니다. *(예정)*

> ⚠️ **보안**: 토큰·키·비밀번호 등 크레덴셜은 코드나 매니페스트에 하드코딩하지 않습니다. Secret Manager / Kubernetes Secret / GitHub Secrets로만 관리합니다.
