output "artifact_registry_repos" {
  description = "생성된 Artifact Registry 저장소 (앱 이름 => 저장소 경로)"
  value = {
    for app_name, repo in google_artifact_registry_repository.this :
    app_name => "${repo.location}-docker.pkg.dev/${var.project_id}/${repo.repository_id}"
  }
}
