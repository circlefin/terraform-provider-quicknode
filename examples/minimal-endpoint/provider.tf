terraform {
  required_providers {
    quicknode = {
      source = "registry.terraform.io/hashicorp/quicknode"
    }
  }
}

provider "quicknode" {
  // Also set via QUICKNODE_APIKEY
  // apikey = ""
}
