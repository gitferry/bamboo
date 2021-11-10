package replica

import (
	"container/list"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/message"
	"time"
)

type Estimator struct {
	estimateWindow    int
	mbStableCaculator *StableCalculator
	pStableCaculator  *StableCalculator
	mbAckMap          map[crypto.Identifier]*AckMgr
	pAckMap           map[crypto.Identifier]*AckMgr
}

type AckMgr struct {
	estimateNum         int
	ackCounts           int
	aveDuration         float64
	accumulatedDuration time.Duration
}

type StableCalculator struct {
	stableTimeList *list.List
	aveStableTime  float64
	estimateWindow int
}

func NewStableCalculator() *StableCalculator {
	sc := new(StableCalculator)
	sc.stableTimeList = list.New()
	return sc
}

func (sc *StableCalculator) AddStableDuration(duration float64) {
	sc.stableTimeList.PushBack(duration)
	if sc.stableTimeList.Len() > sc.estimateWindow {
		last := sc.stableTimeList.Back()
		front := sc.stableTimeList.Front()
		sc.aveStableTime = (sc.aveStableTime*float64(sc.estimateWindow) + last.Value.(float64) - front.Value.(float64)) / float64(sc.estimateWindow)
		sc.stableTimeList.Remove(front)
	}
}

func (sc *StableCalculator) PredictedStableTime() float64 {
	return sc.aveStableTime
}

func NewAckMgr() *AckMgr {
	return &AckMgr{
		estimateNum: config.Configuration.EstimateNum,
	}
}

func NewEstimator() *Estimator {
	return &Estimator{
		estimateWindow:    config.Configuration.EstimateWindow,
		mbStableCaculator: NewStableCalculator(),
		pStableCaculator:  NewStableCalculator(),
		mbAckMap:          make(map[crypto.Identifier]*AckMgr),
		pAckMap:           make(map[crypto.Identifier]*AckMgr),
	}
}

func (am *AckMgr) addAck(ack *message.Ack) (bool, float64) {
	if am.ackCounts >= am.estimateNum {
		return false, am.aveDuration
	}
	am.ackCounts++
	am.accumulatedDuration += ack.AckTime.Sub(ack.SentTime)
	if am.ackCounts == am.estimateNum {
		am.aveDuration = am.accumulatedDuration.Seconds() / float64(am.ackCounts)
		return true, am.aveDuration
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
		isEnough, aveDur := mbAckMgr.addAck(ack)
		if isEnough {
			et.mbStableCaculator.AddStableDuration(aveDur)
		}
	} else if ack.Type == "p" {
		pAckMgr, exists := et.pAckMap[ack.ID]
		if !exists {
			pAckMgr = NewAckMgr()
			et.mbAckMap[ack.ID] = pAckMgr
		}
		isEnough, aveDur := pAckMgr.addAck(ack)
		if isEnough {
			et.pStableCaculator.AddStableDuration(aveDur)
		}
	} else {
		log.Errorf("ack type not specified!")
	}
}

func (et *Estimator) PredictStableTime(t string) time.Duration {
	var d time.Duration
	return d
}
