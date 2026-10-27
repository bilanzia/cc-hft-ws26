terraform {
  required_providers {
    local = {
      source  = "hashicorp/local" # https://registry.terraform.io/providers/hashicorp/local/latest/docs
      version = "~> 2.9"
    }
  }
}

resource "local_file" "hello" {
  filename = "${path.module}/cloud_intro.txt"
  content  = "Infrastruktur als Code funktioniert lokal und in der Cloud!"
}
