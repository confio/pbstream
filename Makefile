.PHONY: test benchmark gen

test:
	go test . -v

gen:
	cd ./_gen && make gen

benchmark:
	cd benchmarks && go test -bench=.
