terraform {
  backend "gcs" {
    bucket = "notiflex-tfstate-454892209447"
    prefix = "gcp/gke/notiflex-cluster"
  }
}
