package mm1

import "testing"

func TestLowLambda(t *testing.T) {
	var lambda float64
	var mu float64
	if _, err := New(lambda, mu); err != ErrLambdaNonPositive {
		t.Error("expected ErrLambdaNonPositive error")
	}
}

func TestLowMu(t *testing.T) {
	var lambda float64 = 1
	var mu = lambda
	if _, err := New(lambda, mu); err != ErrMuNotGreaterThanLambda {
		t.Error("expected ErrMuNotGreaterThanLambda")
	}
}
