package server

import "net"

func (svc *Server) HandleReq(payload []byte, raddr *net.UDPAddr) {}

func (svc *Server) HandlePsh(raddr *net.UDPAddr) {}

func (svc *Server) HandleDat(payload []byte, raddr *net.UDPAddr) {}

func (svc *Server) HandleFin(raddr *net.UDPAddr) {}
