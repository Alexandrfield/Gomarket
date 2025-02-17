package common

import "time"

type PaymentOrder struct {
	number      string
	status      string
	accural     int
	uploaded_at time.Time
}

type UsertOrder struct {
	IdUser string
	Ord    PaymentOrder
}

func CreatUserOrder(idUser string, idOrder string) UsertOrder {
	return UsertOrder{IdUser: idUser, Ord: PaymentOrder{number: idOrder, uploaded_at: time.Now(), status: OrderStatusProcessing}}
}
