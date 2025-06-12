package merrors

import "errors"

var (
	ErrIncorrectCounterValue = errors.New("counter value must be int64")
	ErrIncorrectGaugeValue   = errors.New("counter value must be float64")
)

var (
	ErrIncorrectMetricType     = errors.New("incorrect type. type must be a counter or a gauge")
	ErrMissingMetricValue      = errors.New("missing metric value")
	ErrCannotGetNewMetricValue = errors.New("can'not get new value")
	ErrNotFoundMetric          = errors.New("not found metric")
)

var (
	ErrMocked = errors.New("mocked error")
)
