# app_definitions에서 create_ci_sa = true인 앱만 골라 CI용 서비스 계정을 생성한다.
# GitHub Actions가 WIF(OIDC)로 이 SA를 impersonate하여 Artifact Registry에 이미지를 푸시한다.
# (책 3.4 방식 A의 SA 키 대신 keyless 인증 — 21-workload-identity.tf 참조)
resource "google_service_account" "ci" {
  for_each = {
    for app_name, app in local.app_definitions :
    app_name => app
    if try(app.create_ci_sa, false)
  }

  account_id   = "${each.key}-ci"
  display_name = "${each.key} CI (GitHub Actions)"
  description  = "GitHub Actions에서 Artifact Registry 푸시용"
}

# 방식 A(docker build + push)는 artifactregistry.writer만으로 충분하다.
resource "google_project_iam_member" "ci_ar_writer" {
  for_each = google_service_account.ci

  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${each.value.email}"
}
