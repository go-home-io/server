package fanout

import (
	"testing"
	"github.com/go-home-io/server/plugins/common"
	"time"
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

	if nil == m1 || nil == m2 {
		t.FailNow()
	}

	m1 = nil
	m2 = nil

	fo.UnSubscribeDeviceUpdates(idd1)
	fo.UnSubscribeTriggerUpdates(idd1)
	fo.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{}
	time.Sleep(1 * time.Second)

	if nil != m1 || nil == m2 || !d1Exited {
		t.FailNow()
	}

	m1 = nil
	m2 = nil

	fo.UnSubscribeDeviceUpdates(idd2)
	fo.UnSubscribeTriggerUpdates(idd2)
	fo.ChannelInDeviceUpdates() <- &common.MsgDeviceUpdate{}
	time.Sleep(1 * time.Second)

	if nil != m1 || nil != m2 || !d2Exited {
		t.Fail()
	}
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

	if "" == m1 || "" == m2 {
		t.FailNow()
	}

	m1 = ""
	m2 = ""

	fo.UnSubscribeTriggerUpdates(idd1)
	fo.UnSubscribeDeviceUpdates(idd1)
	fo.ChannelInTriggerUpdates() <- "test"
	time.Sleep(1 * time.Second)

	if "" != m1 || "" == m2 || !d1Exited {
		t.FailNow()
	}

	m1 = ""
	m2 = ""

	fo.UnSubscribeTriggerUpdates(idd2)
	fo.UnSubscribeDeviceUpdates(idd2)
	fo.ChannelInTriggerUpdates() <- "test"
	time.Sleep(1 * time.Second)

	if "" != m1 || "" != m2 || !d2Exited {
		t.Fail()
	}
}
