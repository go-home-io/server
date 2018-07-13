//+build !release

package mocks

type fakeCron struct {
}

func (*fakeCron) AddFunc(spec string, cmd func()) (int, error) {
	return 1, nil
}
func (*fakeCron) RemoveFunc(id int) {

}

func FakeNewCron() *fakeCron {
	return &fakeCron{}
}
