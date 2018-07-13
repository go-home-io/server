package providers

// ICronProvider defines cron provider logic.
type ICronProvider interface {
	AddFunc(spec string, cmd func()) (int, error)
	RemoveFunc(id int)
}
