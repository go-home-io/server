package user

type IUserStorage interface {
	Init() error
	FIndUser(headers map[string]string) (username string, err error)
}
