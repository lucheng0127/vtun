package client

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/lucheng0127/vtun/pkg/cipher"
	"github.com/lucheng0127/vtun/pkg/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
	"golang.org/x/sync/errgroup"
)

const (
	INTERVAL = 5
)

type Client struct {
	Target string
	Cipher cipher.Cipher
	User   string
	Passwd string
	Conn   *net.UDPConn
	IPAddr *netlink.Addr
	Iface  *water.Interface
}

func NewClient(target, key, user, passwd string) (C, error) {
	c := &Client{
		Target: target,
		User:   user,
		Passwd: passwd,
	}

	aesCipher, err := cipher.NewAESCipher(key)
	if err != nil {
		return nil, err
	}
	c.Cipher = aesCipher

	return c, nil
}

func (c *Client) Launch() error {
	udpAddr, err := net.ResolveUDPAddr("udp4", c.Target)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	c.Conn = conn

	// Handle VTPkt
	g := new(errgroup.Group)
	g.Go(func() error {
		for {
			var buf [protocol.MAX_FRG_SIZE]byte
			n, _, err := c.Conn.ReadFromUDP(buf[:])
			if err != nil {
				// Client not work well return error
				return err
			}

			flag, payload, err := protocol.Decode(buf[:n], c.Cipher)
			if err != nil {
				log.Warnf("parse VTPkt from address %s %s", udpAddr.String(), err)
				continue
			}

			switch flag {
			case protocol.HDR_FLG_ACK:
				if err := c.HandleAck(payload); err != nil {
					return err
				}
			case protocol.HDR_FLG_DAT:
				fmt.Println(payload)
			case protocol.HDR_FLG_FIN:
				c.HandleFin()
			default:
				continue
			}
		}
	})

	// Send REQ to auth and get ip
	if err := c.SendReq(); err != nil {
		return err
	}

	return g.Wait()
}

func (c *Client) SendHeartbeat() {
	ticker := time.NewTicker(INTERVAL * time.Second)

	for {
		if err := c.SendPsh(); err != nil {
			log.Error(err)
		}
		log.Debug("Heartbeat sent")

		<-ticker.C
	}
}

func (c *Client) Teardown() {
	// Send fin
	if err := c.SendFin(); err != nil {
		log.Error(err)
	}
}

func (c *Client) HandleSignal(sigChan chan os.Signal) {
	sig := <-sigChan
	log.Infof("received signal: %v, stop server", sig)
	c.Teardown()
	os.Exit(0)
}
