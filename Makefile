VERSION ?= latest
ARCH ?= amd64
ARCH_FLAG := --arch $(ARCH)
IMAGE_NAME = gobench
CBIN = podman
BUILD := $(CBIN) build . --format docker $(ARCH_FLAG)
MANIFEST := $(CBIN) manifest
RUN := $(CBIN) run
NETWORK := $(CBIN) network
RM := $(CBIN) rm
BASE_IMAGE = base
IMAGES = uperf

.ONESHELL:

define build =
export IMAGE_MANIFEST=$(IMAGE_NAME):$@-$(VERSION)
export IMAGE_ARCH=$(IMAGE_NAME):$@-$(VERSION)-$(ARCH)
export  SHA=`$(CBIN) images $$IMAGE_ARCH --format={{.Digest}}`
$(MANIFEST) remove $$IMAGE_MANIFEST $$SHA || true
$(MANIFEST) create $$IMAGE_MANIFEST || true
$(BUILD) -f Containerfile.$@ -t localhost/$$IMAGE_ARCH --manifest $$IMAGE_MANIFEST
endef

.PHONY: $(BASE_IMAGE) $(IMAGES)

$(BASE_IMAGE):
	$(build)

$(IMAGES): $(BASE_IMAGE)
	$(build)

local-eskb-net:
	$(NETWORK) create elastic || true

local-es: local-eskb-net
	$(RM) es01-gobench || true
	$(RUN) --name es01-gobench --net elastic -p 127.0.0.1:9200:9200 -p 127.0.0.1:9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.17.0

local-kb: local-eskb-net
	$(RM) kib01-gobench || true
	$(RUN) --name kib01-gobench --net elastic -p 127.0.0.1:5601:5601 -e "ELASTICSEARCH_HOSTS=http://es01-gobench:9200" docker.elastic.co/kibana/kibana:7.17.0

local-cleanup:
	$(NETWORK) rm elastic
	$(RM) kit01-gobench es01-gobench
