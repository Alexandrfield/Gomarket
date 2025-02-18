package common

import "time"

type PaymentOrder struct {
	Number      string    `json:"number,omitempty"`
	Status      string    `json:"status,omitempty"`
	Accural     int       `json:"accrual,omitempty"`
	Uploaded_at time.Time `json:"uploaded_at,omitempty"`
}

type UsertOrder struct {
	IdUser string
	Ord    PaymentOrder
}

func CreatUserOrder(idUser string, idOrder string) UsertOrder {
	return UsertOrder{IdUser: idUser, Ord: PaymentOrder{Number: idOrder, Uploaded_at: time.Now(), Status: OrderStatusProcessing}}
}

type WithdrawOrder struct {
	Order     string    `json:"order,omitempty"`
	Sum       int       `json:"sum,omitempty"`
	Processed time.Time `json:"processed_at,omitempty"`
}
