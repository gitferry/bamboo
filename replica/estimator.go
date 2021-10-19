package replica

import (
	"github.com/gitferry/bamboo/message"
	"time"
)

type Estimator struct {
}

func NewEstimator() *Estimator {
	return &Estimator{}
}

func (et *Estimator) AddAck(ack *message.Ack) {

}

func (et *Estimator) PredictStableTime(t string) time.Time {
	return time.Now()
}
