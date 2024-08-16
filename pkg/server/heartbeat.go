package server

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	TIMEOUT = 30
)

type EPMonitor struct {
	Beat chan string
	IP   string
	Stop chan string
}

func (em *EPMonitor) Monitor(TimeoutChan chan string) {
	for {
		select {
		case <-em.Stop:
			log.Infof("stop monitor Endpoint with ip %s", em.IP)
			return
		case <-em.Beat:
			continue
		case <-time.After(TIMEOUT * time.Second):
			TimeoutChan <- em.IP
			return
		}
	}
}

type HeartbeatMgr struct {
	MLock        sync.Mutex
	IPMonitorMap map[string]*EPMonitor
	TimeoutChan  chan string
	Svc          *Server
}

func NewHeartbeatMgr(svc *Server) *HeartbeatMgr {
	mgr := &HeartbeatMgr{
		MLock:        sync.Mutex{},
		IPMonitorMap: make(map[string]*EPMonitor),
		TimeoutChan:  make(chan string),
		Svc:          svc,
	}

	go mgr.SendTimeoutFin()
	return mgr
}

func (mgr *HeartbeatMgr) GetEPMonitorByIP(ip string) *EPMonitor {
	em := mgr.IPMonitorMap[ip]
	return em
}

func (mgr *HeartbeatMgr) StopMonitorEPByIP(ip string) {
	em := mgr.GetEPMonitorByIP(ip)
	if em != nil {
		em.Stop <- "STOP"
	}

	mgr.MLock.Lock()
	delete(mgr.IPMonitorMap, ip)
	mgr.MLock.Unlock()
}

func (mgr *HeartbeatMgr) MonitorEPByIP(ip string) {
	em := &EPMonitor{
		Beat: make(chan string),
		IP:   ip,
		Stop: make(chan string),
	}

	mgr.MLock.Lock()
	mgr.IPMonitorMap[ip] = em
	mgr.MLock.Unlock()

	go em.Monitor(mgr.TimeoutChan)
}

func (mgr *HeartbeatMgr) SendTimeoutFin() {
	for {
		ip := <-mgr.TimeoutChan
		log.Infof("heartbeat timeout Endpoint ip %s, close it", ip)

		// Remove it from EPMgr
		mgr.Svc.CloseEPByIP(ip)

		mgr.MLock.Lock()
		delete(mgr.IPMonitorMap, ip)
		mgr.MLock.Unlock()
	}
}
