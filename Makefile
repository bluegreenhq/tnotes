.PHONY: init init-ci init-demo build run dev lint lint-fix lint-workflow test-unit test-e2e test demo-gif demo-mp4 demo clean

init:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	brew install watchexec actionlint
	go mod download

init-ci:
	go mod download

init-demo:
	brew install charmbracelet/tap/vhs ttyd ffmpeg libfaketime

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

test-unit:
	go test -v -run 'Test[^E]' ./...

test-e2e:
	go test -v -run TestE2E ./...

test: lint test-unit test-e2e

demo-gif:
	vhs -o demo.gif demo.tape

demo-mp4:
	vhs -o demo.mp4 demo.tape

demo: demo-gif

clean:
	rm -f tnotes
