cross-build:
	@for os in darwin linux windows; do \
		for arch in amd64 386; do \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -a -o dist/$${os}_$${arch}/check-fireworq; \
		done \
	done

prepare-release: cross-build
	@for os in darwin linux windows; do \
		for arch in amd64 386; do \
			zip mackerel-check-fireworq_$${os}_$${arch}.zip dist/$${os}_$${arch}/check-fireworq; \
		done \
	done

.PHONY: cross-build prepare-release
