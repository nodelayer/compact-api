bin/main: $(shell find . -type f -name '*.go')
	mkdir -p bin
	go build -o ./bin/main main.go

.PHONY: clean
clean:
	rm -rf bin/
