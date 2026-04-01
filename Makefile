.PHONY: build run clean
build:
	CGO_ENABLED=0 go build -o ponyexpress ./cmd/ponyexpress/
run: build
	./ponyexpress
clean:
	rm -f ponyexpress
