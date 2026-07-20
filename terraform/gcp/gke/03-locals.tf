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

      # Workload Identity. 6.2 Secret Manager CSI가 GCP SA 없이 KSA로 시크릿에
      # 접근하기 위해 필요하다. 클러스터+노드풀 양쪽 활성화가 전제.
      workload_identity = true

      # Google Secret Manager CSI addon (6.2). Valkey 비밀번호를 파일로 마운트한다.
      secret_manager = true

      node_pools = {
        default-pool = {
          machine_type = "e2-medium"
          node_count   = 3 # 6.2: WI/CSI DaemonSet(노드당 ~220m) 추가로 B/G 파드 스케줄 여유 확보 위해 3개 (2026-07-20)
          disk_size_gb = 30
          spot         = true
          # 6.2: WI를 노드에서 쓰려면 메타데이터 서버 모드를 GKE_METADATA로.
          workload_metadata = "GKE_METADATA"
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
