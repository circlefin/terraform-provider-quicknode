resource "quicknode_endpoint" "main" {
  network    = "mainnet"
  chain      = "eth"
  label      = "multichain-created-by-terraform"
  multichain = true
}
