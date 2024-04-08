resource "quicknode_endpoint" "main" {
  network = "mainnet"
  chain   = "eth"
  label   = "test-created-by-terraform"
}
