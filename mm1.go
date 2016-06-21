package mm1

import (
	"errors"
	"math"
	"math/rand"
	"runtime"
	"time"
)

var errLambdaNonPositive = errors.New("arrival rate (lambda) must be positive")
var errMuNotGreaterThanLambda = errors.New("service rate (mu) must be greate than arrival rate (lambda)")

// Job contains data on an individual job flowing through the queue.
type Job struct {
	EnterQueue      time.Time
	StartService    time.Time
	EndService      time.Time
	WaitDuration    float64
	ServiceDuration float64
}

// MM1 represents an M/M/1 queue with arrival rate lambda and service rate mu.
type MM1 struct {
	lambda   float64    // arrival rate
	mu       float64    // service rate
	jobChan  chan *Job  // channel for jobs
	doneChan chan *Job  // channel for completed jobs
	OutChan  chan *Job  // channel to send processed completed jobs to caller
	rng      *rand.Rand // random number generator
	stopChan chan bool  // tell arrival and service goroutines to stop
}

// New returns an M/M/1 queue with arrival rate lambda and service rate mu.
// Completed jobs are json encoded on w.
func New(lambda, mu float64) (*MM1, error) {
	// check that lambda is positive
	if lambda <= 0 {
		return nil, errLambdaNonPositive
	}

	// check that mu is greater than lambda
	if mu <= lambda {
		return nil, errMuNotGreaterThanLambda
	}

	// set length of channel to expected queue length + 3 * sd of queue length, to
	// be somewhat safe
	rho := lambda / mu
	expectedLength := rho / (1 - rho)
	sdLength := math.Sqrt(rho / math.Pow(1-rho, 2))
	chanLength := int64(expectedLength + 3*sdLength)

	return &MM1{
		lambda:   lambda,
		mu:       mu,
		jobChan:  make(chan *Job, chanLength),
		doneChan: make(chan *Job, chanLength),
		OutChan:  make(chan *Job, chanLength),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		stopChan: make(chan bool, 4),
	}, nil
}

// Start starts the arrival process and servicing in separate goroutines and
// json encodes completed jobs for the caller.
func (q *MM1) Start() {
	go q.arrivalProcessStart()
	go q.serviceStart()
	go q.processCompletedJobs()
}

// Stop closes job and done channels.
func (q *MM1) Stop() {
	// send four messages on stopChan to signal goroutines
	q.stopChan <- true
	q.stopChan <- true
	q.stopChan <- true
	q.stopChan <- true

	// close channels in q
	close(q.jobChan)
	close(q.doneChan)
	close(q.OutChan)
	close(q.stopChan)
}

func (q *MM1) arrivalProcessStart() {
	for {
		select {
		case <-q.stopChan:
			runtime.Goexit()
		default:
			timer := time.NewTimer(time.Duration(int64(q.rng.ExpFloat64() / q.lambda * float64(1e9))))
			<-timer.C
			q.jobChan <- &Job{
				EnterQueue: time.Now(),
			}
		}
	}
}

func (q *MM1) processCompletedJobs() {
	for {
		select {
		case j := <-q.doneChan:
			j.WaitDuration = j.StartService.Sub(j.EnterQueue).Seconds()
			j.ServiceDuration = j.EndService.Sub(j.StartService).Seconds()
			q.OutChan <- j
		case <-q.stopChan:
			runtime.Goexit()
		}
	}
}

func (q *MM1) serviceStart() {
	for {
		select {
		case <-q.stopChan:
			runtime.Goexit()
		default:
			j := <-q.jobChan
			j.StartService = time.Now()
			timer := time.NewTimer(time.Duration(int64(q.rng.ExpFloat64() / q.mu * float64(1e9))))
			<-timer.C
			j.EndService = time.Now()
			q.doneChan <- j
		}
	}
}
