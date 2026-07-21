# 프로젝트 번호 조회 — WI principal 식별자에 project_number가 필요하다(하드코딩 회피).
data "google_project" "this" {
  project_id = var.project_id
}
