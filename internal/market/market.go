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
	bufferOrder  chan common.UsertOrder
	client       *http.Client
	AddresMarket string
}

func (communicator *CommunicatorAddServer) Init() chan common.UsertOrder {
	communicator.bufferOrder = make(chan common.UsertOrder, 10)
	communicator.client = &http.Client{
		Timeout: time.Second * 5, // интервал ожидания: 1 секунда
	}

	return communicator.bufferOrder
}
func (communicator *CommunicatorAddServer) proccesOrderToAddServer(order common.UsertOrder) {
	waitsTime := []int{1, 2, 3, 4, 5}
	for _, val := range waitsTime {
		ans, err := communicator.sendToAddServer(order)
		if err != nil {
			communicator.Logger.Errorf("sendToAddServer err:%s", err)
		}
		temp := common.UsertOrder{IdUser: order.IdUser, Ord: ans}
		communicator.Storage.UpdateUserOrder(temp)
		if ans.Status == common.OrderStatusProcessing || ans.Status == common.OrderStatusInvalid {
			break
		}
		time.Sleep(time.Second * time.Duration(val))
	}

}
func (communicator *CommunicatorAddServer) sendToAddServer(order common.UsertOrder) (common.PaymentOrder, error) {

	var ord common.PaymentOrder
	url := fmt.Sprintf("http://%s/api/orders/%s", communicator.AddresMarket, order.Ord.Number)
	req, err := http.NewRequest(
		http.MethodGet, url, nil,
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
	ans := make([]byte, 4000)
	n, err := io.Copy(ans, resp.Body)
	if err != nil {
		return ord, fmt.Errorf("error reading body. err:%w", err)
	}
	ans = ans[:n]
	if err := json.Unmarshal(ans, &ord); err != nil {
		return ord, fmt.Errorf("Error umarshal response. err:%s", err.Error())
	}
	return ord, nil
}
func (communicator *CommunicatorAddServer) processorSendToServer() {

	for {
		select {
		case ord := <-communicator.bufferOrder:
			go communicator.sendToAddServer(ord)
		case <-communicator.done:
			communicator.Logger.Debugf("Stop Proccesing payment orders")
			return
		}
	}
}
