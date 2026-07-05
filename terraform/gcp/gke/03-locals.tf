locals {
  # GKE 클러스터 정의 map.
  # 클러스터 이름을 key로, 하위에 리소스 레벨 값(존/노드풀 등)을 중첩한다.
  # 클러스터를 추가하려면 이 map에 항목을 추가한다 (리소스는 for_each로 순회).
  gke_definitions = {
    notiflex-cluster = {
      # 배치할 존 목록. 단일 원소 = Zonal 클러스터.
      az = ["asia-northeast3-a"]

      # Gateway API 채널. 클러스터 생성 시점에 켜둔다 (5장에서 사용, 책 2.5장 지침).
      gateway_api_channel = "CHANNEL_STANDARD"

      node_pools = {
        default-pool = {
          machine_type = "e2-medium"
          node_count   = 0 # 실습 중단: 비용 절감을 위해 0으로 축소 (재개 시 2로 복원)
          disk_size_gb = 30
          spot         = true
        }
      }
    }
  }

  # 모든 리소스에 공통으로 붙이는 라벨 (apps 폴더와 통일).
  common_labels = {
    project    = "notiflex"
    managed-by = "terraform"
  }
}
