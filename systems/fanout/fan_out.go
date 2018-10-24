// Package fanout contains implementation of pub-sub fanout channels.
package fanout

import (
	"math/rand"
	"sync"

	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/utils"
)

// Implements IInternalFanOutProvider.
type provider struct {
	device  sync.Mutex
	trigger sync.Mutex

	inDeviceUpdates  chan *common.MsgDeviceUpdate
	outDeviceUpdates map[int64]chan *common.MsgDeviceUpdate

	inTriggerUpdates  chan string
	outTriggerUpdates map[int64]chan string
}

// NewFanOut constructs new FanOut provider.
func NewFanOut() providers.IInternalFanOutProvider {
	p := &provider{
		inTriggerUpdates:  make(chan string, 10),
		outTriggerUpdates: make(map[int64]chan string),
		inDeviceUpdates:   make(chan *common.MsgDeviceUpdate, 10),
		outDeviceUpdates:  make(map[int64]chan *common.MsgDeviceUpdate),

		device:  sync.Mutex{},
		trigger: sync.Mutex{},
	}

	go p.internalCycle()
	return p
}

// SubscribeDeviceUpdates allows to subscribe to the devices updates.
func (p *provider) SubscribeDeviceUpdates() (int64, chan *common.MsgDeviceUpdate) {
	p.device.Lock()
	defer p.device.Unlock()

	c := make(chan *common.MsgDeviceUpdate, 10)
	rnd := p.getID()
	p.outDeviceUpdates[rnd] = c
	return rnd, c
}

// UnSubscribeDeviceUpdates allows to un-subscribe from the device updates.
// nolint:dupl
func (p *provider) UnSubscribeDeviceUpdates(id int64) {
	p.device.Lock()
	defer p.device.Unlock()

	c, ok := p.outDeviceUpdates[id]
	if !ok {
		return
	}

	close(c)
	delete(p.outDeviceUpdates, id)
}

// ChannelInDeviceUpdates returns input channel for the device updates.
func (p *provider) ChannelInDeviceUpdates() chan *common.MsgDeviceUpdate {
	return p.inDeviceUpdates
}

// SubscribeTriggerUpdates allows to subscribe for the triggers updates.
func (p *provider) SubscribeTriggerUpdates() (int64, chan string) {
	p.trigger.Lock()
	defer p.trigger.Unlock()

	c := make(chan string, 10)
	rnd := p.getID()
	p.outTriggerUpdates[rnd] = c
	return rnd, c
}

// UnSubscribeTriggerUpdates allows to un-subscribe from the triggers updates.
// nolint:dupl
func (p *provider) UnSubscribeTriggerUpdates(id int64) {
	p.trigger.Lock()
	defer p.trigger.Unlock()

	c, ok := p.outTriggerUpdates[id]
	if !ok {
		return
	}

	close(c)
	delete(p.outTriggerUpdates, id)
}

// ChannelInTriggerUpdates returns input channel for the triggers updates.
func (p *provider) ChannelInTriggerUpdates() chan string {
	return p.inTriggerUpdates
}

// Returns random ID.
func (p *provider) getID() int64 {
	return utils.TimeNow() + rand.Int63()
}

func (p *provider) internalCycle() {
	for {
		select {
		case u := <-p.inDeviceUpdates:
			go p.deviceUpdates(u)
		case u := <-p.inTriggerUpdates:
			go p.triggerUpdates(u)
		}
	}
}

// Broadcasts device updates.
func (p *provider) deviceUpdates(update *common.MsgDeviceUpdate) {
	p.device.Lock()
	defer p.device.Unlock()

	for _, v := range p.outDeviceUpdates {
		v <- update
	}
}

// Broadcasts trigger updates.
func (p *provider) triggerUpdates(update string) {
	p.trigger.Lock()
	defer p.trigger.Unlock()

	for _, v := range p.outTriggerUpdates {
		v <- update
	}
}
