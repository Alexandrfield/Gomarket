package common

import "time"

type PaymentOrder struct {
	Uploaded_at time.Time `json:"uploaded_at,omitempty"`
	Number      string    `json:"number,omitempty"`
	Status      string    `json:"status,omitempty"`
	Accural     int       `json:"accrual,omitempty"`
}

type UserOrder struct {
	Ord    PaymentOrder
	IDUser string
}

func CreatUserOrder(idUser string, idOrder string) UserOrder {
	return UserOrder{IDUser: idUser, Ord: PaymentOrder{Number: idOrder,
		Uploaded_at: time.Now(), Status: OrderStatusProcessing}}
}

type WithdrawOrder struct {
	Sum       float64   `json:"sum,omitempty"`
	Processed time.Time `json:"processed_at,omitempty"`
	Order     string    `json:"order,omitempty"`
}
