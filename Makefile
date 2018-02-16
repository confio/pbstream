.PHONY: test benchmark

test:
	go test . -v

benchmark:
	cd benchmarks && go test -bench=.
