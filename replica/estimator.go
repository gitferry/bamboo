package replica

import (
	"container/list"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/message"
	"sort"
	"time"
)

type Estimator struct {
	mbStableCaculator *StableCalculator
	pStableCaculator  *StableCalculator
	mbAckMap          map[crypto.Identifier]*AckMgr
	pAckMap           map[crypto.Identifier]*AckMgr
}

type AckMgr struct {
	estimateNum    int
	delays         []int
	stableDuration time.Duration
}

type StableCalculator struct {
	stableTimeList  *list.List
	aveStableTime   time.Duration
	totalStableTime time.Duration
	estimateWindow  int
}

func NewStableCalculator() *StableCalculator {
	sc := new(StableCalculator)
	sc.stableTimeList = list.New()
	sc.estimateWindow = config.Configuration.EstimateWindow
	return sc
}

func (sc *StableCalculator) AddStableDuration(duration time.Duration) {
	sc.stableTimeList.PushBack(duration)
	front := sc.stableTimeList.Front()
	sc.totalStableTime += duration
	if sc.stableTimeList.Len() > sc.estimateWindow {
		sc.totalStableTime -= front.Value.(time.Duration)
		sc.stableTimeList.Remove(front)
	}
	sc.aveStableTime = time.Duration(float64(sc.totalStableTime.Nanoseconds()) / float64(sc.stableTimeList.Len()))
}

func (sc *StableCalculator) PredictedStableTime() time.Duration {
	return sc.aveStableTime
}

func NewAckMgr() *AckMgr {
	return &AckMgr{
		estimateNum: config.Configuration.EstimateNum,
	}
}

func NewEstimator() *Estimator {
	return &Estimator{
		mbStableCaculator: NewStableCalculator(),
		pStableCaculator:  NewStableCalculator(),
		mbAckMap:          make(map[crypto.Identifier]*AckMgr),
		pAckMap:           make(map[crypto.Identifier]*AckMgr),
	}
}

func (am *AckMgr) AddAck(ack *message.Ack) (bool, time.Duration) {
	duration := ack.AckTime.Sub(ack.SentTime)
	am.delays = append(am.delays, int(duration))
	if len(am.delays) == am.estimateNum {
		sort.Ints(am.delays)
		return true, time.Duration(am.delays[len(am.delays)-1])
	}
	return false, 0
}

func (et *Estimator) AddAck(ack *message.Ack) {
	if ack.Type == "mb" {
		mbAckMgr, exists := et.mbAckMap[ack.ID]
		if !exists {
			mbAckMgr = NewAckMgr()
			et.mbAckMap[ack.ID] = mbAckMgr
		}
		isEnough, aveDur := mbAckMgr.AddAck(ack)
		if isEnough {
			et.mbStableCaculator.AddStableDuration(aveDur)
		}
	} else if ack.Type == "p" {
		pAckMgr, exists := et.pAckMap[ack.ID]
		if !exists {
			pAckMgr = NewAckMgr()
			et.pAckMap[ack.ID] = pAckMgr
		}
		isEnough, aveDur := pAckMgr.AddAck(ack)
		if isEnough {
			et.pStableCaculator.AddStableDuration(aveDur)
		}
	} else {
		log.Errorf("ack type not specified!")
	}
}

func (et *Estimator) PredictStableTime(t string) time.Duration {
	var prediction time.Duration
	if t == "mb" {
		prediction = et.mbStableCaculator.PredictedStableTime()
	} else {
		prediction = et.pStableCaculator.PredictedStableTime()
	}
	if prediction == 0 {
		log.Debugf("Prediction is boosting")
		prediction = time.Duration(config.Configuration.DefaultDelay * int(time.Millisecond))
	}
	return prediction
}
