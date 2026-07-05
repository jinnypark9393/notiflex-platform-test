terraform {
  backend "gcs" {
    bucket = "notiflex-tfstate-454892209447"
    prefix = "gcp/apps"
  }
}
