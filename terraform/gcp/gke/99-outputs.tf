output "cluster_names" {
  description = "생성된 GKE 클러스터 이름 목록"
  value       = [for c in google_container_cluster.this : c.name]
}

output "cluster_locations" {
  description = "클러스터별 배치 위치(존)"
  value       = { for k, c in google_container_cluster.this : k => c.location }
}

output "cluster_endpoints" {
  description = "클러스터 API 엔드포인트 (민감정보)"
  value       = { for k, c in google_container_cluster.this : k => c.endpoint }
  sensitive   = true
}
