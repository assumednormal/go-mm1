# go-mm1

Simulate an M/M/1 queue in Golang.

```
# get help
go run cmd/main.go -h

# simulate M/M/1 queue with arrival rate of 1 and service rate of 2 for 1 minute
go run cmd/main.go -arrival.rate 1 -service.rate 2 -run.duration 1m0s
```
