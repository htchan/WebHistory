variable "DATE" {
  default = "${formatdate("YYYY.MM.DD",timestamp())}"
}

variable "HASH" { default = "TESTING" }

variable "CI" { default="false" }

group "default" {
  targets = ["backend-api","backend-worker"]
}

target "backend" {
  name = "backend-${tgt}"
  context = "./backend"
  dockerfile = "./build/Dockerfile.${tgt}"

  tags = [
    "ghcr.io/htchan/web-history:${tgt}-v${DATE}-${HASH}",
    "ghcr.io/htchan/web-history:${tgt}-latest"
  ]
  platforms = equal(CI, "true") ? ["linux/amd64","linux/arm64","linux/arm/v7"] : []
  output     = [equal(CI, "true") ? "type=registry": "type=docker"]

  matrix = { tgt = ["api","worker"] }
}