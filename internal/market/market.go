package market

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Alexandrfield/Gomarket/internal/common"
	"github.com/Alexandrfield/Gomarket/internal/storage"
)

var ErrOrder = errors.New("order not registrate")
var ErrTooManyRequests = errors.New("too many request to accuracy server")

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
	for {
		ans, retry, err := communicator.sendToAddServer(order)
		if err != nil {
			communicator.Logger.Errorf("sendToAddServer err:%s; retry:%d", err, retry)
			if errors.Is(err, ErrTooManyRequests) {
				time.Sleep(time.Second * time.Duration(retry))
				continue
			}
			if errors.Is(err, ErrOrder) {
				ans.Status = common.OrderStatusInvalid
				ans.Accural = 0.0
			}
		}
		ans.UploadedAt = order.Ord.UploadedAt
		communicator.Logger.Debugf("Send     updated order:%s", order.Ord)
		communicator.Logger.Debugf("Received updated order:%s", ans)
		temp := common.UserOrder{IDUser: order.IDUser, Ord: ans}
		err = communicator.Storage.UpdateUserOrder(&temp)
		if err != nil {
			communicator.Logger.Errorf("Storage.UpdateUserOrder:%s", err)
		}
		break
	}
}
func (communicator *CommunicatorAddServer) sendToAddServer(order *common.UserOrder) (common.PaymentOrder, int, error) {
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
		return ord, 0, fmt.Errorf("http.NewRequest.Do err:%w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			communicator.Logger.Warnf("resp.Body.Close() err: %s\n", err)
		}
	}()
	ans, err := io.ReadAll(resp.Body)
	if err != nil {
		return ord, 0, fmt.Errorf("error reading body. err:%w", err)
	}
	communicator.Logger.Warnf("accuravy servvice ans.Status: %s", resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusNoContent:
		return ord, 0, ErrOrder
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		ret, err := strconv.Atoi(retryAfter)
		if err != nil {
			communicator.Logger.Warnf("problem with atoi str%s, err:%s", retryAfter, err)
			ret = 0
		}
		return ord, ret, ErrTooManyRequests
	}

	if err := json.Unmarshal(ans, &ord); err != nil {
		communicator.Logger.Warnf("try unmarshal ans: %s", ans)
		return ord, 0, fmt.Errorf("error umarshal response. err:%s", err.Error())
	}
	return ord, 0, nil
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
