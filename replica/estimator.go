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
	estimateWindow      int
	mbStableTime        time.Duration
	pStableTime         time.Duration
	mbAckStableTimeList *list.List
	pAckStableTimeList  *list.List
	mbAckMap            map[crypto.Identifier]*AckMgr
	pAckMap             map[crypto.Identifier]*AckMgr
}

type AckMgr struct {
	estimateNum         int
	ackCounts           int
	aveDuration         float64
	accumulatedDuration time.Duration
}

func NewAckMgr() *AckMgr {
	return &AckMgr{
		estimateNum: config.Configuration.EstimateNum,
	}
}

func NewEstimator() *Estimator {
	return &Estimator{
		estimateWindow:      config.Configuration.EstimateWindow,
		mbAckStableTimeList: list.New(),
		pAckStableTimeList:  list.New(),
		mbAckMap:            make(map[crypto.Identifier]*AckMgr),
		pAckMap:             make(map[crypto.Identifier]*AckMgr),
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
			et.mbAckStableTimeList.PushBack(aveDur)
		}
	} else if ack.Type == "p" {
		pAckMgr, exists := et.pAckMap[ack.ID]
		if !exists {
			pAckMgr = NewAckMgr()
			et.mbAckMap[ack.ID] = pAckMgr
		}
		isEnough, aveDur := pAckMgr.addAck(ack)
		if isEnough {
			et.pAckStableTimeList.PushBack(aveDur)
		}
	} else {
		log.Errorf("ack type not specified!")
	}
}

func (et *Estimator) PredictStableTime(t string) time.Duration {
	var d time.Duration
	return d
}
