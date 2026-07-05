# app_definitions에서 create_registry = true인 앱만 골라 Docker 저장소를 생성한다.
# 저장소 ID는 map의 key(앱 이름)를 그대로 사용한다.
resource "google_artifact_registry_repository" "this" {
  for_each = {
    for app_name, app in local.app_definitions :
    app_name => app
    if try(app.create_registry, false)
  }

  repository_id = each.key
  location      = var.region
  format        = "DOCKER"

  labels = local.common_labels
}
