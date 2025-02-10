package handle

import (
	"encoding/json"
	"net/http"

	"github.com/Alexandrfield/Gomarket/internal/common"
	"github.com/Alexandrfield/Gomarket/internal/storage"
)

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type ServiceHandler struct {
	Logger  common.Logger
	Storage storage.StorageCommunicator
}

func (han *ServiceHandler) Rgistarte(res http.ResponseWriter, req *http.Request) {

}
func (han *ServiceHandler) Login(res http.ResponseWriter, req *http.Request) {

}
func (han *ServiceHandler) Orders(res http.ResponseWriter, req *http.Request) {

}
func (han *ServiceHandler) GetOrders(res http.ResponseWriter, req *http.Request) {

}
func (han *ServiceHandler) GetBalance(res http.ResponseWriter, req *http.Request) {

}
func (han *ServiceHandler) Withdraw(res http.ResponseWriter, req *http.Request) {

}
func (han *ServiceHandler) Withdrawals(res http.ResponseWriter, req *http.Request) {

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

	res.Header().Set("Content-Type", "application/json")
	if err != nil {
		han.Logger.Warnf("problem with unmarshal:%w", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
