state_store_required_provider {
  test = {
    source = "hashicorp/test"
    version = "1.2.3"
  }
}

from {
  state_store "test_store" {
    path = "terraform.tfstate"
  }
}
