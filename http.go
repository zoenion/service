package service

import (
	"fmt"
	"github.com/omecodes/common/log"
	pb "github.com/omecodes/common/proto/service"
	"net/http"
	"strings"
)

func (box *Box) StartGateway(params *GatewayNodeParams) error {
	box.serverMutex.Lock()
	defer box.serverMutex.Unlock()

	listener, err := box.listen(params.Port, params.Node.Security, params.Tls)
	if err != nil {
		return err
	}

	address := listener.Addr().String()
	if box.params.Domain != "" {
		address = strings.Replace(address, strings.Split(address, ":")[0], box.params.Domain, 1)
	}
	router := params.ProvideRouter()

	var handler http.Handler

	if len(params.MiddlewareList) > 0 {
		handler = router
		for _, m := range params.MiddlewareList {
			handler = m.Middleware(handler)
		}

	} else {
		handler = router
	}

	log.Info("starting HTTP server", log.Field("gPRCNode", params.Node.Id), log.Field("address", address))
	srv := &http.Server{
		Addr:    address,
		Handler: handler,
	}
	gt := &httpNode{}
	gt.Server = srv
	gt.Address = address
	if params.Tls != nil || params.Node.Security != pb.Security_None {
		gt.Scheme = "https"
	} else {
		gt.Scheme = "http"
	}

	gt.Name = params.Node.Id
	box.httpNodes[params.Node.Id] = gt
	go func() {
		err = srv.Serve(listener)
		if err != nil {
			if err != http.ErrServerClosed {
				log.Error("http server stopped", err)
			}

			if box.info != nil {
				var newNodeList []*pb.Node
				for _, node := range box.info.Nodes {
					if node.Id != params.Node.Id {
						newNodeList = append(newNodeList, node)
					}
				}
				box.info.Nodes = newNodeList
				_ = box.registry.RegisterService(box.info)
			}
		}

	}()

	if !box.params.Autonomous && box.registry != nil {
		n := params.Node
		n.Address = address
		if box.info == nil {
			box.info = new(pb.Info)
			box.info.Id = box.Name()
			box.info.Type = params.ServiceType
			if box.info.Meta == nil {
				box.info.Meta = map[string]string{}
			}
		}
		box.info.Nodes = append(box.info.Nodes, n)

		// gt.RegistryID, err = box.registry.RegisterService(info, pb.ActionOnRegisterExistingService_AddNodes|pb.ActionOnRegisterExistingService_UpdateExisting)
		err = box.registry.RegisterService(box.info)
		if err != nil {
			log.Error("could not register gateway", err, log.Field("name", params.Node.Id))
		}
	}
	return nil
}

func (box *Box) stopGateways() error {
	box.serverMutex.Lock()
	defer box.serverMutex.Unlock()
	for name, srv := range box.httpNodes {
		err := srv.Stop()
		if err != nil {
			log.Error(fmt.Sprintf("gateway stopped"), err, log.Field("gPRCNode", name))
		}
	}
	return nil
}
