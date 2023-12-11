# Nomad Heart Cilium

Super early WIP, so good luck!

## Messy setup:

- Run Consul `consul agent -dev` for Cilium state backend
- `make all` to install Cilium, CNI config, and this plugin
- Build and run Nomad branch: [db-cni-experiments](https://github.com/hashicorp/nomad/compare/db-cni-experiments)
  so Nomad sends `CNI_ARGS` for the plugin to read.
- Nomad run `cilium.nomad.hcl` (with Nomad client config to succeed at the junk in there),
  or run cilium-agent some other way that still puts `/var/run/cilium/cilium.sock` on the host
- `cilium status` & `cilium endpoint list` <- does this work?
- Nomad run `web.nomad.hcl` which uses our "nomad-heart-cilium" CNI plugin config
- `cilium endpoint list` again, should have an endpoint with a couple nomad labels on it

That's as far as I've gotten so far!
