package mm1

import (
	"errors"
	"math"
	"math/rand"
	"runtime"
	"time"
)

// ErrLambdaNonPositive occurs when the arrival rate (lambda) is not positive.
var ErrLambdaNonPositive = errors.New("arrival rate (lambda) must be positive")

// ErrMuNotGreaterThanLambda occurs when the service rate (mu) is not greater
// than the arrival rate (lambda).
var ErrMuNotGreaterThanLambda = errors.New("service rate (mu) must be greater than arrival rate (lambda)")

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
	stopChan chan bool  // tell arrival and service goroutines to stop
}

// New returns an M/M/1 queue with arrival rate lambda and service rate mu.
// Completed jobs are json encoded on w.
func New(lambda, mu float64) (*MM1, error) {
	// check that lambda is positive
	if lambda <= 0 {
		return nil, ErrLambdaNonPositive
	}

	// check that mu is greater than lambda
	if mu <= lambda {
		return nil, ErrMuNotGreaterThanLambda
	}

	// set length of channel to expected queue length + 10 * sd of queue length, to
	// be somewhat safe
	rho := lambda / mu
	expectedLength := rho / (1 - rho)
	sdLength := math.Sqrt(rho / math.Pow(1-rho, 2))
	chanLength := int64(expectedLength + 10*sdLength)

	return &MM1{
		lambda:   lambda,
		mu:       mu,
		jobChan:  make(chan *Job, chanLength),
		DoneChan: make(chan *Job, chanLength),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		stopChan: make(chan bool, 2),
	}, nil
}

// Start starts the arrival process and servicing in separate goroutines.
func (q *MM1) Start() time.Time {
	go q.arrivalProcess()
	go q.serviceProcess()
	return time.Now()
}

// Stop closes job and done channels.
func (q *MM1) Stop() {
	// send two messages on stopChan to signal goroutines
	q.stopChan <- true
	q.stopChan <- true

	// close channels in q
	close(q.jobChan)
	close(q.DoneChan)
	close(q.stopChan)
}

func (q *MM1) arrivalProcess() {
	for {
		select {
		case <-q.stopChan:
			runtime.Goexit()
		default:
			dur := time.Duration(int64(q.rng.ExpFloat64() / q.lambda * float64(1e9)))
			timer := time.NewTimer(dur)
			<-timer.C
			q.jobChan <- &Job{
				EnterQueue: time.Now(),
			}
		}
	}
}

func (q *MM1) serviceProcess() {
	for {
		select {
		case <-q.stopChan:
			runtime.Goexit()
		case j := <-q.jobChan:
			j.StartService = time.Now()
			dur := time.Duration(int64(q.rng.ExpFloat64() / q.mu * float64(1e9)))
			timer := time.NewTimer(dur)
			<-timer.C
			j.EndService = time.Now()
			q.DoneChan <- j
		}
	}
}
