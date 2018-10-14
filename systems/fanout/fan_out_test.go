package fanout

import (
	"testing"
	"time"

	"github.com/go-home-io/server/plugins/common"
	"github.com/stretchr/testify/assert"
)

// Tests devices updates channels.
func TestDeviceUpdates(t *testing.T) {
	fo := NewFanOut()
	idd1, d1 := fo.SubscribeDeviceUpdates()
	idd2, d2 := fo.SubscribeDeviceUpdates()
	var m1 *common.MsgDeviceUpdate
	var m2 *common.MsgDeviceUpdate
	d1Exited := false
	d2Exited := false

	go func() {
		for m := range d1 {
			m1 = m
		}

		d1Exited = true
	}()

	go func() {
		for m := range d2 {
			m2 = m
		}

		d2Exited = true
	}()

	fo.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{}
	time.Sleep(1 * time.Second)
	assert.NotNil(t, m1, "channel 1")
	assert.NotNil(t, m2, "channel 2")

	m1 = nil
	m2 = nil

	fo.UnSubscribeDeviceUpdates(idd1)
	fo.UnSubscribeTriggerUpdates(idd1)
	fo.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{}
	time.Sleep(1 * time.Second)

	assert.Nil(t, m1, "unsubscribe channel 1")
	assert.NotNil(t, m2, "unsubscribe channel 2")
	assert.True(t, d1Exited, "exit channel 1")

	m1 = nil
	m2 = nil

	fo.UnSubscribeDeviceUpdates(idd2)
	fo.UnSubscribeTriggerUpdates(idd2)
	fo.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{}
	time.Sleep(1 * time.Second)

	assert.Nil(t, m1, "full unsubscribe channel 1")
	assert.Nil(t, m2, "full unsubscribe channel 2")
	assert.True(t, d2Exited, "exit channel 2")
}

// Tests triggers updates channels.
func TestTriggerUpdates(t *testing.T) {
	fo := NewFanOut()
	idd1, d1 := fo.SubscribeTriggerUpdates()
	idd2, d2 := fo.SubscribeTriggerUpdates()
	var m1 string
	var m2 string
	d1Exited := false
	d2Exited := false

	go func() {
		for m := range d1 {
			m1 = m
		}

		d1Exited = true
	}()

	go func() {
		for m := range d2 {
			m2 = m
		}

		d2Exited = true
	}()

	fo.ChannelInTriggerUpdates() <- "test"
	time.Sleep(1 * time.Second)

	assert.Equal(t, "test", m1, "channel 1")
	assert.Equal(t, "test", m2, "channel 2")

	m1 = ""
	m2 = ""

	fo.UnSubscribeTriggerUpdates(idd1)
	fo.UnSubscribeDeviceUpdates(idd1)
	fo.ChannelInTriggerUpdates() <- "test"
	time.Sleep(1 * time.Second)

	assert.Equal(t, "", m1, "unsubscribe channel 1")
	assert.Equal(t, "test", m2, "unsubscribe channel 2")
	assert.True(t, d1Exited, "exit channel 1")

	m1 = ""
	m2 = ""

	fo.UnSubscribeTriggerUpdates(idd2)
	fo.UnSubscribeDeviceUpdates(idd2)
	fo.ChannelInTriggerUpdates() <- "test"
	time.Sleep(1 * time.Second)

	assert.Equal(t, "", m1, "full unsubscribe channel 1")
	assert.Equal(t, "", m2, "full unsubscribe channel 2")
	assert.True(t, d2Exited, "exit channel 2")

}
