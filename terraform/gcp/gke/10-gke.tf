# gke_definitions map을 순회하여 클러스터를 생성한다.
resource "google_container_cluster" "this" {
  for_each = local.gke_definitions

  name     = each.key
  location = each.value.az[0] # 단일 존 = Zonal 클러스터

  # 기본 노드풀을 제거하고 아래 google_container_node_pool로 별도 관리한다 (모범 사례).
  remove_default_node_pool = true
  initial_node_count       = 1

  # 실습용 클러스터 — 삭제 보호 해제 (provider 6.x+ 기본값 true).
  deletion_protection = false

  # Gateway API — locals에 gateway_api_channel이 있을 때만 설정한다.
  dynamic "gateway_api_config" {
    for_each = try(each.value.gateway_api_channel, null) != null ? [each.value.gateway_api_channel] : []
    content {
      channel = gateway_api_config.value
    }
  }
}

# 클러스터 × 노드풀을 평탄화하여 노드풀 리소스를 생성한다.
resource "google_container_node_pool" "this" {
  for_each = {
    for np in flatten([
      for cluster_name, cluster in local.gke_definitions : [
        for pool_name, pool in cluster.node_pools : {
          key          = "${cluster_name}/${pool_name}"
          cluster_name = cluster_name
          pool_name    = pool_name
          pool         = pool
        }
      ]
    ]) : np.key => np
  }

  name       = each.value.pool_name
  cluster    = google_container_cluster.this[each.value.cluster_name].name
  location   = google_container_cluster.this[each.value.cluster_name].location
  node_count = each.value.pool.node_count

  node_config {
    machine_type = each.value.pool.machine_type
    disk_size_gb = each.value.pool.disk_size_gb
    spot         = each.value.pool.spot
  }
}
