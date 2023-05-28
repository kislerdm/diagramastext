resource "google_kms_key_ring" "this" {
  name     = "crypto"
  location = "global"
}

locals {
  key_rotation_d   = 14
  key_rotation_sec = local.key_rotation_d * 86400
}

resource "google_kms_crypto_key" "ciam" {
  name            = "ciam"
  key_ring        = google_kms_key_ring.this.id
  rotation_period = "${local.key_rotation_sec}s"

  version_template {
    algorithm = "GOOGLE_SYMMETRIC_ENCRYPTION"
  }

  purpose = "ENCRYPT_DECRYPT"

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_kms_crypto_key_iam_member" "this" {
  crypto_key_id = google_kms_crypto_key.ciam.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:${google_service_account.this.email}"
}
