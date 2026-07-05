locals {
  # 앱 정의 map.
  # 앱 이름을 key로, 하위에 그 앱이 GCP에 필요로 하는 리소스 여부를 플래그로 둔다.
  # 앱을 추가하려면 이 map에 항목을 추가한다 (리소스는 for_each로 순회).
  # 리소스 이름은 이 map의 key(앱 이름)를 그대로 사용한다.
  #
  # 필드 설명:
  #   create_registry = Artifact Registry(Docker) 저장소 생성 여부
  app_definitions = {
    notiflex = {
      create_registry = true
    }
  }

  # 모든 리소스에 공통으로 붙이는 라벨 (gke 폴더와 통일).
  common_labels = {
    project    = "notiflex"
    managed-by = "terraform"
  }
}
