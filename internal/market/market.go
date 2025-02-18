package market

import (
	"github.com/Alexandrfield/Gomarket/internal/common"
	"github.com/Alexandrfield/Gomarket/internal/storage"
)

type CommunicatorAddServer struct {
	Logger      common.Logger
	Storage     storage.StorageCommunicator
	done        chan struct{}
	bufferOrder chan common.UsertOrder
}

func (communicator *CommunicatorAddServer) Init() chan common.UsertOrder {
	communicator.bufferOrder = make(chan common.UsertOrder, 10)
	return communicator.bufferOrder
}
func (communicator *CommunicatorAddServer) sendToAddServer(order common.UsertOrder) {
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
