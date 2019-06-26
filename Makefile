.PHONY: fib
.DEFAULT_GOAL:= fib

fib:
	go build -o fib main.go 
