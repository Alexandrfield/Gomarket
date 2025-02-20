package common

import "time"

type PaymentOrder struct {
	Uploaded_at time.Time `json:"uploaded_at,omitempty"`
	Number      string    `json:"number,omitempty"`
	Status      string    `json:"status,omitempty"`
	Accural     int       `json:"accrual,omitempty"`
}

type UserOrder struct {
	IDUser string
	Ord    PaymentOrder
}

func CreatUserOrder(idUser string, idOrder string) UserOrder {
	return UserOrder{IDUser: idUser, Ord: PaymentOrder{Number: idOrder,
		Uploaded_at: time.Now(), Status: OrderStatusProcessing}}
}

type WithdrawOrder struct {
	Processed time.Time `json:"processed_at,omitempty"`
	Order     string    `json:"order,omitempty"`
	Sum       float64   `json:"sum,omitempty"`
}
