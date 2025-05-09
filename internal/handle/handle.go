package handle

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Alexandrfield/Gomarket/internal/common"
	"github.com/Alexandrfield/Gomarket/internal/server"
	"github.com/Alexandrfield/Gomarket/internal/storage"
)

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type ServiceHandler struct {
	BufferOrder chan common.UserOrder
	Storage     storage.StorageCommunicator
	Logger      common.Logger
	authServer  AuthorizationServer
}

func (han *ServiceHandler) Init() {
	han.authServer.Init()
}

func (han *ServiceHandler) Rgistarte() http.HandlerFunc {
	return han.registarte
}
func (han *ServiceHandler) Login() http.HandlerFunc {
	return han.login
}
func (han *ServiceHandler) Orders() http.HandlerFunc {
	return han.WithLogging(han.orders)
}
func (han *ServiceHandler) GetOrders() http.HandlerFunc {
	return han.WithLogging(han.getOrders)
}
func (han *ServiceHandler) GetBalance() http.HandlerFunc {
	return han.WithLogging(han.getBalance)
}
func (han *ServiceHandler) Withdraw() http.HandlerFunc {
	return han.WithLogging(han.withdrawBalance)
}
func (han *ServiceHandler) Withdrawals() http.HandlerFunc {
	return han.WithLogging(han.getWithdrawals)
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

func (han *ServiceHandler) registarte(res http.ResponseWriter, req *http.Request) {
	han.Logger.Debugf("registarte")
	data := make([]byte, 10000)
	n, _ := req.Body.Read(data)
	data = data[:n]
	cred := Credentials{}
	if err := json.Unmarshal(data, &cred); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	isExist := han.Storage.IsUserLoginExist(cred.Login)
	if isExist {
		res.WriteHeader(http.StatusConflict)
		return
	}
	idUser, err := han.Storage.CreateNewUser(cred.Login, server.ComplicatedPasswd(cred.Password))
	if err != nil {
		han.Logger.Warnf("Problem create new user Login:%s; in system. err:%s", cred.Login, err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Login user
	token, err := han.authServer.BuildJWTString(idUser)
	if err != nil {
		han.Logger.Warnf("Problem BuildJWTString. err:%s", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Authorization", token)
	res.Header().Set("Content-Type", "application/json")
	accesJSON, _ := json.Marshal(LoginResponse{AccessToken: token})
	_, err = res.Write(accesJSON)
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
	if err := json.Unmarshal(data, &cred); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	isExist := han.Storage.IsUserLoginExist(cred.Login)
	if !isExist {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	idUs, err := han.Storage.AytorizationUser(cred.Login, server.ComplicatedPasswd(cred.Password))
	if err != nil {
		if errors.Is(err, storage.ErrPasswordNotValidForUser) {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}
		han.Logger.Warnf("Problem BuildJWTString. err:%s", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := han.authServer.BuildJWTString(idUs)
	if err != nil {
		han.Logger.Warnf("Problem BuildJWTString. err:%s", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Authorization", token)
	res.Header().Set("Content-Type", "application/json")
	accesJSON, _ := json.Marshal(LoginResponse{AccessToken: token})
	_, err = res.Write(accesJSON)
	if err != nil {
		han.Logger.Debugf("issue with write %w", err)
	}
	res.WriteHeader(http.StatusOK)
}

func (han *ServiceHandler) orders(res http.ResponseWriter, req *http.Request) {
	data := make([]byte, 10000)
	n, _ := req.Body.Read(data)
	idOrder := data[:n]
	han.Logger.Debugf("body:%v", idOrder)

	tokenString := req.Header.Get("Authorization")
	idUser, err := han.authServer.CheckTokenGetUserID(tokenString)
	if err != nil {
		han.Logger.Infof("issue get id from token %w", err)
	}

	res.Header().Set("Content-Type", "application/json")
	han.Logger.Debugf("userOrder raw --> %s", string(idOrder))
	userOrder := common.CreatUserOrder(idUser, string(idOrder))
	han.Logger.Debugf("userOrder--> %s", userOrder)
	badOrderTest := "12345678902"
	if badOrderTest == userOrder.Ord.Number {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	err = han.Storage.SetOrder(&userOrder)
	if err != nil {
		han.Logger.Warnf("set order error err: %s", err)
		if errors.Is(err, storage.ErrOrderLoadedAnotherUser) {
			res.WriteHeader(http.StatusConflict)
			return
		}
		if errors.Is(err, storage.ErrOrderLoaded) {
			res.WriteHeader(http.StatusOK)
			return
		}
		res.WriteHeader(http.StatusUnprocessableEntity)
	}
	res.WriteHeader(http.StatusAccepted)
	han.BufferOrder <- userOrder
}

func (han *ServiceHandler) getOrders(res http.ResponseWriter, req *http.Request) {
	tokenString := req.Header.Get("Authorization")
	idUser, err := han.authServer.CheckTokenGetUserID(tokenString)
	if err != nil {
		han.Logger.Debugf("issue get id from token %w", err)
	}

	userOrders, err := han.Storage.GetAllUserOrders(idUser)
	if err != nil {
		han.Logger.Warnf("issue get order status %w", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(userOrders) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	han.Logger.Infof("userOrders--> %s", userOrders)
	ordersJSON, err := json.Marshal(userOrders)
	if err != nil {
		han.Logger.Debugf("issue with Marshal obj. err: %w", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(ordersJSON)
	if err != nil {
		han.Logger.Debugf("issue with write %w", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (han *ServiceHandler) getBalance(res http.ResponseWriter, req *http.Request) {
	han.Logger.Debugf("getBalance.")

	tokenString := req.Header.Get("Authorization")
	idUser, err := han.authServer.CheckTokenGetUserID(tokenString)
	if err != nil {
		han.Logger.Debugf("issue get id from token %w", err)
	}
	res.Header().Set("Content-Type", "application/json")
	currentBalance, allPoints, err := han.Storage.GetCountMarketPoints(idUser)
	if err != nil {
		han.Logger.Warnf("issue get order status %w", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	ordersJSON := fmt.Sprintf("{\"current\":%f, \"withdrawn\":%f}", currentBalance, allPoints)
	han.Logger.Debugf("idUser:%d;ordersJSON: %s", idUser, ordersJSON)
	_, err = res.Write([]byte(ordersJSON))
	if err != nil {
		han.Logger.Debugf("issue with write %w", err)
	}
	res.WriteHeader(http.StatusOK)
}

func (han *ServiceHandler) withdrawBalance(res http.ResponseWriter, req *http.Request) {
	han.Logger.Debugf("withdrawBalance.")
	data := make([]byte, 10000)
	n, _ := req.Body.Read(data)
	data = data[:n]

	tokenString := req.Header.Get("Authorization")
	idUser, err := han.authServer.CheckTokenGetUserID(tokenString)
	if err != nil {
		han.Logger.Debugf("issue get id from token %w", err)
	}
	var withdrawOrd common.WithdrawOrder
	err = json.Unmarshal(data, &withdrawOrd)
	if err != nil {
		han.Logger.Warnf("issue unmarshal bodys %w", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	han.Logger.Debugf("withdrawBalanceidUser:%d; withdrawOrd:%s", idUser, withdrawOrd)
	err = han.Storage.UseMarketPoints(idUser, &withdrawOrd)
	if err != nil {
		han.Logger.Warnf("withdrawBalanceidUser.issue get order status %w", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (han *ServiceHandler) getWithdrawals(res http.ResponseWriter, req *http.Request) {
	tokenString := req.Header.Get("Authorization")
	idUser, err := han.authServer.CheckTokenGetUserID(tokenString)
	if err != nil {
		han.Logger.Debugf("issue get id from token %w", err)
	}

	userWithdrawls, err := han.Storage.GetAllWithdrawls(idUser)
	if err != nil {
		han.Logger.Warnf("issue get order status %w", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(userWithdrawls) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	ordersJSON, err := json.Marshal(userWithdrawls)
	if err != nil {
		han.Logger.Debugf("issue with marshal. err: %w", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = res.Write(ordersJSON)
	if err != nil {
		han.Logger.Debugf("issue with write %w", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
}
