variable "DATE" {
  default = "${formatdate("YYYY.MM.DD",timestamp())}"
}

variable "HASH" { default = "TESTING" }

variable "BAKE_CI" { default = "false" }

variable "BRANCH" { default = "" }
variable "IMAGE_TAG" { default = "${equal(BRANCH,"master") ? "latest" : "beta"}" }

group "default" {
  targets = ["backend-api","backend-worker","backend-scheduler"]
}

target "backend" {
  name = "backend-${service}"
  context = "./backend"
  dockerfile = "./build/Dockerfile"
  cache-from = [ "type=gha" ]
  cache-to = [ "type=gha,mode=max" ]

  args = {
    SERVICE = "${service}"
  }

  labels = {
    "org.opencontainers.image.source" = "https://github.com/htchan/WebHistory"
  }
  attest = [
    "type=provenance,disabled=true"
  ]

  tags = [
    "ghcr.io/htchan/web-history:${service}-v${DATE}-${substr(HASH,0,7)}",
    "ghcr.io/htchan/web-history:${service}-${IMAGE_TAG}"
  ]
  platforms = equal(BAKE_CI, "true") ? ["linux/amd64","linux/arm64"] : []
  output     = [equal(BAKE_CI, "true") ? "type=registry": "type=docker"]

  matrix = { service = ["api","worker","scheduler"] }
}
