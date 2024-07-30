job "web" {
  group "web" {
    count = 1

    network {
      mode = "cni/nomad-heart-cilium"
      cni {
        args = {
          NomadNamespace = "${NOMAD_NAMESPACE}"
          NomadTaskGroup = "${NOMAD_GROUP_NAME}"
        }
      }

      port "http" {
        static = 8080
        to     = 8080
      }
    }

    task "python" {
      driver = "docker"

      config {
        image   = "python:slim"
        command = "python3"
        args    = ["-m", "http.server", "${NOMAD_PORT_http}"]
        ports   = ["http"]

        labels {
          hiiii = "sure ok" # todo...?
        }
      }
    }
  }
}
