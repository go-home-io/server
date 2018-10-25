//+build !release

package mocks

type fakeCron struct {
}

func (*fakeCron) AddFunc(spec string, cmd func()) (int, error) {
	return 1, nil
}
func (*fakeCron) RemoveFunc(id int) {

}

// FakeNewCron creates a fake cron provider.
func FakeNewCron() *fakeCron {
	return &fakeCron{}
}
