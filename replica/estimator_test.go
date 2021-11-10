package replica

import (
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/utils"
	"github.com/stretchr/testify/require"
	"math/rand"
	"sort"
	"testing"
	"time"
)

// estimate number is set to 3
// estimate window is set to 3

// insert one stable time into the calculator, which should return the one
func TestStableCalculator_PredictedStableTime1(t *testing.T) {
	config.Configuration.EstimateWindow = 3
	sc := NewStableCalculator()
	sd1 := 120 * time.Millisecond
	sc.AddStableDuration(sd1)
	require.Equal(t, sd1, sc.PredictedStableTime())
}

// insert three stable time into the calculator, which should return the average
func TestStableCalculator_PredictedStableTime2(t *testing.T) {
	config.Configuration.EstimateWindow = 3
	sc := NewStableCalculator()
	sd1 := 120 * time.Millisecond
	sd2 := 320 * time.Millisecond
	sd3 := 220 * time.Millisecond
	sc.AddStableDuration(sd1)
	sc.AddStableDuration(sd2)
	sc.AddStableDuration(sd3)
	require.Equal(t, (sd1+sd2+sd3)/3, sc.PredictedStableTime())
}

// insert four stable time into the calculator, which should return the average of the last three
func TestStableCalculator_PredictedStableTime3(t *testing.T) {
	config.Configuration.EstimateWindow = 3
	sc := NewStableCalculator()
	sd1 := 120 * time.Millisecond
	sd2 := 320 * time.Millisecond
	sd3 := 220 * time.Millisecond
	sd4 := 150 * time.Millisecond
	sc.AddStableDuration(sd1)
	sc.AddStableDuration(sd2)
	sc.AddStableDuration(sd3)
	sc.AddStableDuration(sd4)
	require.Equal(t, (sd2+sd3+sd4)/3, sc.PredictedStableTime())
}

// add two acks, should return false
func TestAckMgr_AddAck1(t *testing.T) {
	config.Configuration.EstimateNum = 3
	ackMgr := NewAckMgr()
	ack1, _ := NewMockAckWithDelay("mb")
	ack2, _ := NewMockAckWithDelay("mb")
	isEnough, _ := ackMgr.AddAck(ack1)
	require.False(t, isEnough)
	isEnough, _ = ackMgr.AddAck(ack2)
	require.False(t, isEnough)
}

// add three acks, should return true and average
func TestAckMgr_AddAck2(t *testing.T) {
	config.Configuration.EstimateNum = 3
	ackMgr := NewAckMgr()
	ack1, d1 := NewMockAckWithDelay("mb")
	ack2, d2 := NewMockAckWithDelay("mb")
	ack3, d3 := NewMockAckWithDelay("mb")
	delays := []int{int(d1), int(d2), int(d3)}
	sort.Ints(delays)
	ackMgr.AddAck(ack1)
	ackMgr.AddAck(ack2)
	isEnough, duration := ackMgr.AddAck(ack3)
	require.True(t, isEnough)
	require.Equal(t, time.Duration(delays[2]), duration)
}

// add four acks, should return false
func TestAckMgr_AddAck3(t *testing.T) {
	config.Configuration.EstimateNum = 3
	ackMgr := NewAckMgr()
	ack1, _ := NewMockAckWithDelay("mb")
	ack2, _ := NewMockAckWithDelay("mb")
	ack3, _ := NewMockAckWithDelay("mb")
	ack4, _ := NewMockAckWithDelay("mb")
	ackMgr.AddAck(ack1)
	ackMgr.AddAck(ack2)
	ackMgr.AddAck(ack3)
	isEnough, _ := ackMgr.AddAck(ack4)
	require.False(t, isEnough)
}

func NewMockAckWithDelay(t string) (*message.Ack, time.Duration) {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	ack := new(message.Ack)
	ack.SentTime = time.Now().Add(time.Duration(r.Intn(100) * int(time.Millisecond)))
	ack.AckTime = ack.SentTime.Add(time.Duration(r.Intn(100) * int(time.Millisecond)))
	ack.ID = utils.IdentifierFixture()
	ack.Type = t
	return ack, ack.AckTime.Sub(ack.SentTime)
}
