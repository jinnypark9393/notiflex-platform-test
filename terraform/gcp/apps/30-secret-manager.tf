# Google Secret Manager 접근 권한 (6.2/6.3).
# secret 리소스와 값 자체는 Terraform에서 관리하지 않는다 — 값이 state에 평문으로
# 남는 것을 피하기 위해 secret은 gcloud로 생성한 상태를 유지하고, 여기서는 API 활성화와
# KSA→secret 접근 IAM 바인딩만 코드로 관리한다.

resource "google_project_service" "secretmanager" {
  service            = "secretmanager.googleapis.com"
  disable_on_destroy = false
}

# valkey-password secret에 대한 KSA별 secretAccessor 바인딩 (gsm_secret_accessors 맵 순회).
resource "google_secret_manager_secret_iam_member" "valkey_password" {
  for_each = local.gsm_secret_accessors

  secret_id = "valkey-password" # gcloud로 생성된 secret (Terraform 미관리)
  role      = "roles/secretmanager.secretAccessor"
  member    = "${local.wi_pool_prefix}/ns/${each.value.namespace}/sa/${each.value.ksa}"

  depends_on = [google_project_service.secretmanager]
}
