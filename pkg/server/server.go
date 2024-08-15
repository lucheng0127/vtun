package server

import (
	"os"

	"github.com/lucheng0127/vtun/pkg/auth"
	"github.com/lucheng0127/vtun/pkg/endpoint"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
)

type Svc interface {
	Launch() error
	Teardown()
	HandleSignal(chan os.Signal)
}

type Server struct {
	Iface   *water.Interface
	Port    int
	AuthMgr *auth.BaseAuthMgr
	IPMgr   *endpoint.IPMgr
}

func NewServer(iface *water.Interface, ipRange, userDB string, port, maskLen int) (Svc, error) {
	svc := new(Server)

	svc.Iface = iface
	svc.Port = port
	svc.AuthMgr = &auth.BaseAuthMgr{DB: userDB}

	ipMgr, err := endpoint.NewIPMgr(ipRange, maskLen)
	if err != nil {
		return nil, err
	}

	svc.IPMgr = ipMgr
	return svc, nil
}

func (svc *Server) Launch() error {
	log.Infof("launch server iface %s", svc.Iface.Name())
	return nil
}

func (svc *Server) Teardown() {
}

func (svc *Server) HandleSignal(sigChan chan os.Signal) {
	sig := <-sigChan
	log.Infof("received signal: %v, stop server", sig)
	svc.Teardown()
}
