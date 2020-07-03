# Use:
#
# `make build` will build a container -- you must supply a tag.
# `make push` will push the container -- you must supply a tag.

REPO = jpedro1992/
NAME = sfc-controller
TAG = 0.0.5

build:
	docker build -t $(REPO)$(NAME):$(TAG) .

push: build
	docker push $(REPO)$(NAME):$(TAG)
