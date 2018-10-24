package utils

import (
	"go-home.io/x/server/providers"
	"gopkg.in/robfig/cron.v2"
)

// Cron implementation.
type provider struct {
	cron *cron.Cron
}

// NewCron creates a new scheduler.
func NewCron() providers.ICronProvider {
	p := provider{
		cron: cron.New(),
	}

	p.cron.Start()
	return &p
}

// AddFunc schedules a new job.
func (p *provider) AddFunc(spec string, cmd func()) (int, error) {
	id, err := p.cron.AddFunc(spec, cmd)
	return int(id), err
}

// RemoveFunc removes scheduled job from cron.
func (p *provider) RemoveFunc(id int) {
	p.cron.Remove(cron.EntryID(id))
}
