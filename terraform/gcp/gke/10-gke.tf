# gke_definitions map을 순회하여 클러스터를 생성한다.
resource "google_container_cluster" "this" {
  for_each = local.gke_definitions

  name     = each.key
  location = each.value.az[0] # 단일 존 = Zonal 클러스터

  # 공통 라벨 (apps 폴더와 통일).
  resource_labels = local.common_labels

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

  # Workload Identity (6.2) — locals의 workload_identity가 true일 때만.
  dynamic "workload_identity_config" {
    for_each = try(each.value.workload_identity, false) ? [1] : []
    content {
      workload_pool = "${var.project_id}.svc.id.goog"
    }
  }

  # Google Secret Manager CSI addon (6.2) — locals의 secret_manager가 true일 때만.
  dynamic "secret_manager_config" {
    for_each = try(each.value.secret_manager, false) ? [1] : []
    content {
      enabled = true
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
    # disk_type은 optional. 미지정이면 GKE 기본값(pd-balanced). 7.2 신규 풀은 SSD 쿼터
    # 회피를 위해 pd-standard(HDD)를 명시한다.
    disk_type = try(each.value.pool.disk_type, null)
    spot      = each.value.pool.spot

    # 노드(GCE 인스턴스) 공통 라벨 (apps 폴더와 통일).
    resource_labels = local.common_labels

    # Workload Identity 메타데이터 서버 (6.2) — pool.workload_metadata가 있을 때만.
    dynamic "workload_metadata_config" {
      for_each = try(each.value.pool.workload_metadata, null) != null ? [each.value.pool.workload_metadata] : []
      content {
        mode = workload_metadata_config.value
      }
    }
  }
}
