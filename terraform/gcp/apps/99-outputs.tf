output "artifact_registry_repos" {
  description = "생성된 Artifact Registry 저장소 (앱 이름 => 저장소 경로)"
  value = {
    for app_name, repo in google_artifact_registry_repository.this :
    app_name => "${repo.location}-docker.pkg.dev/${var.project_id}/${repo.repository_id}"
  }
}

output "ci_service_accounts" {
  description = "CI용 서비스 계정 이메일 (앱 이름 => SA email)"
  value = {
    for app_name, sa in google_service_account.ci :
    app_name => sa.email
  }
}

output "wif_provider" {
  description = "GitHub Actions ci.yaml의 workload_identity_provider에 넣을 전체 리소스 이름"
  value       = google_iam_workload_identity_pool_provider.github.name
}
