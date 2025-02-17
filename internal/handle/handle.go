package handle

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Alexandrfield/Gomarket/internal/common"
	"github.com/Alexandrfield/Gomarket/internal/market"
	"github.com/Alexandrfield/Gomarket/internal/storage"
)

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type ServiceHandler struct {
	Logger      common.Logger
	Storage     storage.StorageCommunicator
	authServer  AuthorizationServer
	BufferOrder chan market.PaymentOrder
}

func (han *ServiceHandler) Init() {
	han.authServer.Init()
}

func (han *ServiceHandler) Rgistarte() http.HandlerFunc {
	return han.WithLogging(han.registarte)
}
func (han *ServiceHandler) Login() http.HandlerFunc {
	return han.WithLogging(han.login)
}
func (han *ServiceHandler) Orders() http.HandlerFunc {
	return han.WithLogging(han.orders)
}
func (han *ServiceHandler) GetOrders() http.HandlerFunc {
	return han.WithLogging(han.getOrders)

}
func (han *ServiceHandler) GetBalance(res http.ResponseWriter, req *http.Request) http.HandlerFunc {

}
func (han *ServiceHandler) Withdraw(res http.ResponseWriter, req *http.Request) http.HandlerFunc {

}
func (han *ServiceHandler) Withdrawals(res http.ResponseWriter, req *http.Request) http.HandlerFunc {

}

func (han *ServiceHandler) registarte(res http.ResponseWriter, req *http.Request) {
	data := make([]byte, 10000)
	n, _ := req.Body.Read(data)
	data = data[:n]
	cred := Credentials{}
	han.Logger.Debugf("GetJSONValue body:%v", data)
	if err := json.Unmarshal(data, &cred); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	isExist, err := han.Storage.IsUserLoginExist(cred.Login)
	if err != nil {
		han.Logger.Warnf("Problem check Login:%s; in system. err:%s", cred.Login, err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if isExist {
		http.Error(res, err.Error(), http.StatusConflict)
		return
	}
	_, err = han.Storage.CreateNewUser(cred.Login, cred.Password)
	if err != nil {
		han.Logger.Warnf("Problem create new user Login:%s; in system. err:%s", cred.Login, err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Login user
	token, err := han.authServer.BuildJWTString()
	if err != nil {
		han.Logger.Warnf("Problem BuildJWTString. err:%s", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write([]byte(token))
	if err != nil {
		han.Logger.Debugf("issue with write %w", err)
	}
	res.WriteHeader(http.StatusOK)
}
func (han *ServiceHandler) login(res http.ResponseWriter, req *http.Request) {
	data := make([]byte, 10000)
	n, _ := req.Body.Read(data)
	data = data[:n]
	cred := Credentials{}
	han.Logger.Debugf("GetJSONValue body:%v", data)
	if err := json.Unmarshal(data, &cred); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	isExist, err := han.Storage.IsUserLoginExist(cred.Login)
	if err != nil {
		han.Logger.Warnf("Problem check Login:%s; in system. err:%s", cred.Login, err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if isExist {
		http.Error(res, err.Error(), http.StatusConflict)
		return
	}

	token, err := han.authServer.BuildJWTString()
	if err != nil {
		han.Logger.Warnf("Problem BuildJWTString. err:%s", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write([]byte(token))
	if err != nil {
		han.Logger.Debugf("issue with write %w", err)
	}
	res.WriteHeader(http.StatusOK)
}
func (han *ServiceHandler) orders(res http.ResponseWriter, req *http.Request) {
	data := make([]byte, 10000)
	n, _ := req.Body.Read(data)
	idOrder := data[:n]
	han.Logger.Debugf("body:%v", data)

	tokenString := req.Header.Get("Authorization")
	idUser, err := han.authServer.CheckTokenGetUserID(tokenString)
	if err != nil {
		han.Logger.Debugf("issue get id from token %w", err)
	}

	st, err := han.Storage.GetOrderStatus(string(idUser), string(idOrder))
	if err != nil {
		han.Logger.Warnf("issue get order status %w", err)
	}

	switch st {
	case "empty":
		res.WriteHeader(http.StatusAccepted)
	case "busy":
		res.WriteHeader(http.StatusConflict)
		return
	default:
		res.WriteHeader(http.StatusOK)
	}
	han.BufferOrder <- common.CreatUserOrder(idUser, data)
	res.Header().Set("Content-Type", "application/json")
	// _, err = res.Write([]byte(token))
	// if err != nil {
	// 	han.Logger.Debugf("issue with write %w", err)
	// }
	//res.WriteHeader(http.StatusOK)
}
func (han *ServiceHandler) getOrders(res http.ResponseWriter, req *http.Request) {
	data := make([]byte, 10000)
	n, _ := req.Body.Read(data)
	data = data[:n]
	han.Logger.Debugf("body:%v", data)

	tokenString := req.Header.Get("Authorization")
	idOrder, err := han.authServer.CheckTokenGetUserID(tokenString)
	if err != nil {
		han.Logger.Debugf("issue get id from token %w", err)
	}

	st, err := han.Storage.GetOrderStatus(string(idOrder), strconv.Itoa(idOrder))
	if err != nil {
		han.Logger.Warnf("issue get order status %w", err)
	}

	switch st {
	case "empty":
		res.WriteHeader(http.StatusAccepted)
	case "busy":
		res.WriteHeader(http.StatusConflict)
		return
	default:
		res.WriteHeader(http.StatusOK)
	}
	han.BufferOrder <- market.PaymentOrder{IdUser: idOrder, IdOrder: data}
	res.Header().Set("Content-Type", "application/json")

}
