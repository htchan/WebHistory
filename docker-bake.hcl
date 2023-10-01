variable "DATE" {
  default = "${formatdate("YYYY.MM.DD",timestamp())}"
}

variable "HASH" { default = "TESTING" }

variable "BAKE_CI" { default="false" }

group "default" {
  targets = ["backend-api","backend-worker"]
}

target "backend" {
  name = "backend-${tgt}"
  context = "./backend"
  dockerfile = "./build/Dockerfile.${tgt}"

  labels = {
    "org.opencontainers.image.source" = "https://github.com/htchan/WebHistory"
  }

  tags = [
    "ghcr.io/htchan/web-history:${tgt}-v${DATE}-${HASH}",
    "ghcr.io/htchan/web-history:${tgt}-latest"
  ]
  platforms = equal(BAKE_CI, "true") ? ["linux/amd64","linux/arm64","linux/arm/v7"] : []
  output     = [equal(BAKE_CI, "true") ? "type=registry": "type=docker"]

  matrix = { tgt = ["api","worker"] }
}