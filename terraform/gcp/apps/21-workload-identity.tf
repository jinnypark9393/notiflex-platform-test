# GitHub Actions keyless 인증을 위한 Workload Identity Federation.
# GitHub이 발급한 OIDC 토큰을 GCP STS가 검증·교환하여 CI SA를 impersonate한다.
# SA 키(장기 크레덴셜)를 GitHub Secrets에 저장하지 않는다.

# WIF 토큰 교환에 필요한 STS API (iam/iamcredentials는 ch2.3에서 활성화됨).
resource "google_project_service" "sts" {
  service            = "sts.googleapis.com"
  disable_on_destroy = false
}

# 풀/프로바이더는 프로젝트당 하나만 두고, 저장소 제한은 SA 바인딩(attribute.repository)에서 건다.
resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github-pool"
  display_name              = "GitHub Actions"
  description               = "GitHub Actions OIDC 토큰 교환용 풀"

  depends_on = [google_project_service.sts]
}

resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"
  display_name                       = "GitHub OIDC"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }

  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.repository" = "assertion.repository"
  }

  # 소유 계정 외 저장소의 토큰은 풀 단계에서 차단한다.
  attribute_condition = "assertion.repository_owner == \"jinnypark9393\""
}

# 앱별 CI SA impersonation 허용 — 해당 앱의 GitHub 저장소(ci_github_repo)에서 온 토큰만.
resource "google_service_account_iam_member" "ci_wif" {
  for_each = {
    for app_name, app in local.app_definitions :
    app_name => app
    if try(app.create_ci_sa, false)
  }

  service_account_id = google_service_account.ci[each.key].name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${each.value.ci_github_repo}"
}
