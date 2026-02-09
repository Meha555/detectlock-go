test:
	GO111MODULE=on GOPROXY=https://goproxy.cn,direct go test ./...

bench:
	GO111MODULE=on GOPROXY=https://goproxy.cn,direct go test -run=none -count=1 -benchtime=10000x -benchmem -bench=. ./... | grep Benchmark

vet:
	go vet ./...

fmt:
	go fmt ./...

ifeq ($(OS),Windows_NT)
    EXE_SUFFIX := .exe
else
    EXE_SUFFIX :=
endif

examples:
	go build -o waiting_on_eachother$(EXE_SUFFIX) ./examples/waiting_on_eachother/
	go build -o reentry$(EXE_SUFFIX) ./examples/reentry/
	go build -o rwlock$(EXE_SUFFIX) ./examples/rwlock/

clean:
	rm -f mutex01$(EXE_SUFFIX)
	rm -f mutex02$(EXE_SUFFIX)
	rm -f mutex03$(EXE_SUFFIX)

.PHONY: test bench vet fmt examples clean
