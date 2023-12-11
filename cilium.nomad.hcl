# based on https://github.com/cosmonic/netreap#cilium-agent

variable "cilium_tag" {
  default = "v1.14.4"
  #default = "local" # if you build cilium container yourself
}

variable "consul_token" {
  default = ""
}

job "cilium" {
  group "cilium" {
    task "agent" {
      driver = "docker"
      config {
        image = "cilium/cilium:${var.cilium_tag}"

        command = "cilium-agent"
        args = [
          "--kvstore=consul",
          "--kvstore-opt=consul.address=127.0.0.1:8500",
          #"--kvstore=nomad", # WIP
          "--enable-ipv6=false",
          "-t", "geneve",
          "--enable-l7-proxy=false",
          "--ipv4-range=172.16.0.0/16",
        ]

        network_mode = "host"
        privileged   = true

        volumes = [
          "/var/run/cilium:/var/run/cilium",
          "/sys/fs/bpf:/sys/fs/bpf",
        ]
      }
      env {
        CONSUL_HTTP_TOKEN = var.consul_token
      }
    }
  }
}
