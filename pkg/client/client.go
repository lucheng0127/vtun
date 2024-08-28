package client

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/lucheng0127/vtun/pkg/cipher"
	"github.com/lucheng0127/vtun/pkg/iface"
	"github.com/lucheng0127/vtun/pkg/protocol"
	"github.com/lucheng0127/vtun/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sync/errgroup"
)

const (
	INTERVAL = 5
)

type Client struct {
	Target     string
	Cipher     cipher.Cipher
	User       string
	Passwd     string
	Conn       *net.UDPConn
	IPAddr     *netlink.Addr
	Iface      iface.IFace
	AllowedIPs []string
}

func NewClient(target, key, user, passwd string, allowedIPs []string) (C, error) {
	c := &Client{
		Target:     target,
		User:       user,
		Passwd:     passwd,
		AllowedIPs: allowedIPs,
	}

	aesCipher, err := cipher.NewAESCipher(key)
	if err != nil {
		return nil, err
	}
	c.Cipher = aesCipher

	return c, nil
}

func (c *Client) Launch() error {
	if runtime.GOOS == "windows" {
		if err := utils.ExtractWintun(); err != nil {
			return fmt.Errorf("extract wintun.dll %s", err.Error())
		}
	}

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
				if c.Iface == nil {
					continue
				}

				// Forward fragement from udp to tun interface
				if _, err := c.Iface.Write(payload); err != nil {
					return err
				}
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

	// Sync allowed-ips
	if err := c.SendIps(); err != nil {
		return err
	}

	// Forward traffic from iface to udp
	go c.IfaceToNet()

	return g.Wait()
}

func (c *Client) IfaceToNet() {
	for {
		if c.Iface == nil {
			// Waiting for tun interface ready
			time.Sleep(1 * time.Second)
			continue
		}

		var buf [protocol.MAX_FRG_SIZE]byte

		n, err := c.Iface.Read(buf[:])
		if err != nil {
			log.Error(err)
		}

		if err := c.SendDat(buf[:n]); err != nil {
			log.Error(err)
		}
	}
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
