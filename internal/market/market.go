package market

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Alexandrfield/Gomarket/internal/common"
	"github.com/Alexandrfield/Gomarket/internal/storage"
)

type CommunicatorAddServer struct {
	Logger       common.Logger
	Storage      storage.StorageCommunicator
	done         chan struct{}
	bufferOrder  chan common.UserOrder
	client       *http.Client
	AddresMarket string
}

func (communicator *CommunicatorAddServer) Init() chan common.UserOrder {
	communicator.bufferOrder = make(chan common.UserOrder, 10)
	communicator.client = &http.Client{
		Timeout: time.Second * 5, // интервал ожидания: 1 секунда
	}
	go communicator.processorSendToServer()
	return communicator.bufferOrder
}
func (communicator *CommunicatorAddServer) proccesOrderToAddServer(order *common.UserOrder) {
	communicator.Logger.Debugf("proccesOrderToAddServer order:%s", *order)
	waitsTime := []int{1, 2, 3, 4, 5}
	for _, val := range waitsTime {
		ans, err := communicator.sendToAddServer(order)
		if err != nil {
			communicator.Logger.Errorf("sendToAddServer err:%s", err)
		}
		ans.UploadedAt = order.Ord.UploadedAt
		communicator.Logger.Debugf("Send     updated order:%s", order.Ord)
		communicator.Logger.Debugf("Received updated order:%s", ans)
		temp := common.UserOrder{IDUser: order.IDUser, Ord: ans}
		err = communicator.Storage.UpdateUserOrder(&temp)
		if err != nil {
			communicator.Logger.Errorf("Storage.UpdateUserOrder:%s", err)
		}
		if ans.Status == common.OrderStatusProcessing || ans.Status == common.OrderStatusInvalid {
			break
		}
		time.Sleep(time.Second * time.Duration(val))
	}
}
func (communicator *CommunicatorAddServer) sendToAddServer(order *common.UserOrder) (common.PaymentOrder, error) {
	var ord common.PaymentOrder
	communicator.Logger.Infof("communicator.AddresMarket: %s", communicator.AddresMarket)
	url := fmt.Sprintf("http://%s/api/orders/%s", communicator.AddresMarket, order.Ord.Number)
	req, err := http.NewRequest(
		http.MethodGet, url, http.NoBody,
	)
	if err != nil {
		communicator.Logger.Warnf("http.NewRequest. err: %s\n", err)
	}
	const encod = "gzip"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", encod)
	req.Header.Set("Content-Encoding", encod)

	resp, err := communicator.client.Do(req)
	if err != nil {
		return ord, fmt.Errorf("http.NewRequest.Do err:%w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			communicator.Logger.Warnf("resp.Body.Close() err: %s\n", err)
		}
	}()
	ans, err := io.ReadAll(resp.Body)
	if err != nil {
		return ord, fmt.Errorf("error reading body. err:%w", err)
	}
	communicator.Logger.Warnf("accuravy servvice ans.Status: %s", resp.StatusCode)
	if err := json.Unmarshal(ans, &ord); err != nil {
		communicator.Logger.Warnf("try unmarshal ans: %s", ans)
		return ord, fmt.Errorf("error umarshal response. err:%s", err.Error())
	}
	return ord, nil
}
func (communicator *CommunicatorAddServer) processorSendToServer() {
	for {
		select {
		case ord := <-communicator.bufferOrder:
			go communicator.proccesOrderToAddServer(&ord)
		case <-communicator.done:
			communicator.Logger.Debugf("Stop Proccesing payment orders")
			return
		}
	}
}
