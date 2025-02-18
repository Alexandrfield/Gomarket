package storage

import "github.com/Alexandrfield/Gomarket/internal/common"

type StorageCommunicator interface {
	CreateNewUser(login string, password string) (string, error)
	AytorizationUser(login string, password string) (string, error) // return user_id
	IsUserLoginExist(login string) (bool, error)
	GetOrderStatus(user string, order string) (string, error)
	GetCountMarketPoints(user string) (float32, float32, error)
	UseMarketPoints(user string, withdrawOrd common.WithdrawOrder) error
	GetAllUserOrders(user string) ([]common.PaymentOrder, error)
	GetAllWithdrawls(user string) ([]common.WithdrawOrder, error)
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
func (stor *Storage) GetCountMarketPoints(user string) (float32, float32, error) {
	return 0.0, 0.0, nil
}
func (stor *Storage) UseMarketPoints(user string, withdrawOrd common.WithdrawOrder) error {
	return nil
}
func (stor *Storage) GetAllUserOrders(user string) ([]common.PaymentOrder, error) {
	return []common.PaymentOrder{}, nil
}
func (stor *Storage) GetAllWithdrawls(user string) ([]common.WithdrawOrder, error) {
	return []common.WithdrawOrder{}, nil
}
