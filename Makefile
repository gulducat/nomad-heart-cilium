plugin: # build and install this custom plugin
	cd plugin && $(MAKE)

cilium: # at least need cilium-cni binary on nomad client machine
	./install-cilium.sh

config: # and this config to run that binary, and our extra custom one
	cp cni.conflist /opt/cni/config/nomad-heart-cilium.conflist

all: config plugin cilium

.PHONY: plugin cilium config all
