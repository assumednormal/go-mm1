package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/assumednormal/go-mm1"
)

type summary struct {
	n  int64    // number of jobs completed
	si float64  // sum of interarrival durations
	ss float64  // sum of service durations
	pj *mm1.Job // previous job
}

func (s *summary) update(j *mm1.Job) {
	s.n++
	s.si += j.EnterQueue.Sub(s.pj.EnterQueue).Seconds()
	s.ss += j.EndService.Sub(j.StartService).Seconds()
	s.pj = j
}

func (s *summary) String() string {
	out := fmt.Sprintf("Number of jobs completed: %d Jobs", s.n)
	out = fmt.Sprintf("%s\nMean Interarrival Duration: %0.5f Seconds", out,
		s.si/float64(s.n))
	out = fmt.Sprintf("%s\nMean Service Duration: %0.5f Seconds\n", out,
		s.ss/float64(s.n))
	return out
}

func run() int {
	var lambda, mu float64
	var runDur time.Duration
	flag.Float64Var(&lambda, "arrival.rate", 0, "arrival rate, lambda")
	flag.Float64Var(&mu, "service.rate", 0, "service rate, mu")
	flag.DurationVar(&runDur, "run.duration", time.Duration(0),
		"amount of time to run simulation")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	q, err := mm1.New(lambda, mu)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("%v\n", err))
		return 1
	}
	start := q.Start()

	end := time.NewTicker(runDur)

	s := &summary{
		n:  int64(0),
		si: float64(0),
		ss: float64(0),
		pj: &mm1.Job{EnterQueue: start},
	}

	for {
		select {
		case j := <-q.DoneChan:
			s.update(j)
		case sig := <-sigs:
			os.Stdout.WriteString(fmt.Sprintf("\nQuitting on signal: %v", sig))
			os.Stdout.WriteString(s.String())
			return 0
		case <-end.C:
			os.Stdout.WriteString(s.String())
			return 0
		}
	}
}

func main() {
	os.Exit(run())
}
