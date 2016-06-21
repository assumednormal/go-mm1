package mm1

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

// ErrLambdaNonPositive occurs when the arrival rate (lambda) is not positive.
var ErrLambdaNonPositive = errors.New("arrival rate (lambda) must be positive")

// ErrMuNotGreaterThanLambda occurs when the service rate (mu) is not greater
// than the arrival rate (lambda).
var ErrMuNotGreaterThanLambda = errors.New(
	"service rate (mu) must be greater than arrival rate (lambda)")

// Job contains data on an individual job flowing through the queue.
type Job struct {
	EnterQueue   time.Time
	StartService time.Time
	EndService   time.Time
}

// MM1 represents an M/M/1 queue with arrival rate lambda and service rate mu.
type MM1 struct {
	lambda   float64    // arrival rate
	mu       float64    // service rate
	jobChan  chan *Job  // channel for jobs
	DoneChan chan *Job  // channel for completed jobs
	rng      *rand.Rand // random number generator
}

// New returns an M/M/1 queue with arrival rate lambda and service rate mu.
func New(lambda, mu float64) (*MM1, error) {
	// check that lambda is positive
	if lambda <= 0 {
		return nil, ErrLambdaNonPositive
	}

	// check that mu is greater than lambda
	if mu <= lambda {
		return nil, ErrMuNotGreaterThanLambda
	}

	// set length of channel to 1 + expected length + 10 * sd of length
	rho := lambda / mu
	expectedLength := rho / (1 - rho)
	sdLength := math.Sqrt(rho / math.Pow(1-rho, 2))
	chanLength := int64(1 + expectedLength + 10*sdLength)

	return &MM1{
		lambda:   lambda,
		mu:       mu,
		jobChan:  make(chan *Job, chanLength),
		DoneChan: make(chan *Job, chanLength),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Start starts the arrival and servicing processes in separate goroutines.
func (q *MM1) Start() time.Time {
	go q.arrivalProcess()
	go q.serviceProcess()
	return time.Now()
}

func (q *MM1) arrivalProcess() {
	for {
		dur := time.Duration(int64(q.rng.ExpFloat64() / q.lambda * float64(1e9)))
		timer := time.NewTimer(dur)
		<-timer.C
		q.jobChan <- &Job{
			EnterQueue: time.Now(),
		}
	}
}

func (q *MM1) serviceProcess() {
	for {
		j := <-q.jobChan
		j.StartService = time.Now()
		dur := time.Duration(int64(q.rng.ExpFloat64() / q.mu * float64(1e9)))
		timer := time.NewTimer(dur)
		<-timer.C
		j.EndService = time.Now()
		q.DoneChan <- j
	}
}
