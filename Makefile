.PHONY: test release version

VERSION := $(shell go run *.go -v | awk '{print $$2}')

test:
	go test -v ./...

release: version
	docker run --rm --privileged \
		-v $$PWD:/go/src/github.com/progrhyme/shelp \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/github.com/progrhyme/shelp \
		-e GITHUB_TOKEN \
		goreleaser/goreleaser release --rm-dist

version:
	git commit -m $(VERSION)
	git tag -a v$(VERSION) -m $(VERSION)
	git push origin v$(VERSION)
	git push origin master
