package storage

type StorageCommunicator interface {
	CreateNewUser(login string, password string) (string, error)
	AytorizationUser(login string, password string) (string, error) // return user_id
	IsUserLoginExist(login string) (bool, error)
	GetOrderStatus(user string, order string) (string, error)
	GetCountMarketPoints(user string) (int, error)
	UseMarketPoints(user string, points int) error
}

func GetStorage(config Config) (StorageCommunicator, error) {
	return &Storage{}, nil
}

type Storage struct {
}

func (stor *Storage) CreateNewUser(login string, password string) (string, error) {
	return "qwerty", nil
}
func (stor *Storage) AytorizationUser(login string, password string) (string, error) { // return user_id
	return "qwerty", nil
}
func (stor *Storage) IsUserLoginExist(login string) (bool, error) {
	return false, nil
}
func (stor *Storage) GetOrderStatus(user string, order string) (string, error) {
	return "test", nil
}
func (stor *Storage) GetCountMarketPoints(user string) (int, error) {
	return 0, nil
}
func (stor *Storage) UseMarketPoints(user string, points int) error {
	return nil
}
