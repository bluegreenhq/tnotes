.PHONY: init init-ci init-demo build run dev lint lint-fix lint-workflow lint-release test-unit test-e2e test demo-gif demo release clean

init:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	brew install watchexec actionlint goreleaser
	go mod download

init-ci:
	go mod download

init-demo:
	brew install asciinema agg

build:
	go build -o tnotes .

run:
	go run . $(ARGS)

dev:
	watchexec -r -e go -- go run .

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

lint-workflow:
	actionlint

lint-release:
	goreleaser check

test-unit:
	go test -v -run 'Test[^E]' ./...

test-e2e:
	go test -v -run TestE2E ./...

test: lint test-unit test-e2e

demo-gif:
	bash demo.sh

demo: demo-gif

release:
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=0.1.0)
endif
	@if git log -1 --format='%s' | grep -qi '\[ci skip\]\|\[skip ci\]'; then \
		echo "Error: HEAD commit contains [ci skip]. Release workflow will not trigger." >&2; \
		exit 1; \
	fi
	git tag v$(VERSION)
	git push origin v$(VERSION)

clean:
	rm -f tnotes
