variable "project_id" {
  description = "GCP 프로젝트 ID"
  type        = string
}

variable "region" {
  description = "기본 리전"
  type        = string
  default     = "asia-northeast3"
}

variable "zone" {
  description = "provider 기본 존 (Zonal 클러스터 배치 기준)"
  type        = string
  default     = "asia-northeast3-a"
}
