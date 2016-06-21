package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/assumednormal/go-mm1"
)

func main() {
	var lambda, mu float64
	flag.Float64Var(&lambda, "arrival.rate", 0, "arrival rate, lambda")
	flag.Float64Var(&mu, "service.rate", 0, "service rate, mu")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	queue, err := mm1.New(lambda, mu)
	if err != nil {
		panic(err)
	}
	queue.Start()
	defer queue.Stop()

	enc := json.NewEncoder(os.Stdout)

	go func() {
		for {
			j := <-queue.OutChan
			enc.Encode(j)
		}
	}()

	sig := <-sigs
	fmt.Printf("\nQuitting on signal: %v\n", sig)
}
