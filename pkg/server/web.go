package server

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lucheng0127/vtun/pkg/utils"
	log "github.com/sirupsen/logrus"
)

//go:embed static/index.html
var indexFile []byte
var indexFilename = fmt.Sprintf("index_%s.html", utils.RandStr(4))

// Web server
type WebServer struct {
	Port int
}

func (svc *WebServer) Serve() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	router := gin.Default()

	// Extract and load index
	if err := os.WriteFile(indexFilename, indexFile, 0666); err != nil {
		log.Errorf("extract index.html to current dir %s", err.Error())
		return
	}
	router.LoadHTMLFiles(string(indexFilename))

	// URL
	router.GET("/endpoints", listEndpoints)
	router.GET("/destination", listDestination)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, string(indexFilename), nil)
	})

	router.Run(fmt.Sprintf(":%d", svc.Port))
}

// Handles
type EPEntry struct {
	User    string
	Addr    string
	IP      string
	LoginAt string
}

func listEndpoints(c *gin.Context) {
	data := []*EPEntry{
		{
			User:    "vtun server",
			Addr:    fmt.Sprintf(":%d", VSvc.(*Server).Port),
			IP:      VSvc.(*Server).IPAddr.String(),
			LoginAt: VSvc.(*Server).StartAt,
		},
	}

	for ipKey := range VSvc.(*Server).EPMgr.EPIPMap {
		ep := VSvc.(*Server).EPMgr.GetEPByIP(ipKey)

		epEn := &EPEntry{
			User:    ep.User,
			Addr:    ep.RAddr.String(),
			IP:      ep.IP.String(),
			LoginAt: ep.LoginTime,
		}

		data = append(data, epEn)
	}

	c.JSON(http.StatusOK, data)
}

type FwdEntry struct {
	CIDR   string
	EPUser string
	EPIP   string
}

func listDestination(c *gin.Context) {
	data := []*FwdEntry{
		{
			CIDR:   "Default",
			EPUser: "vtun server",
			EPIP:   VSvc.(*Server).IPAddr.String(),
		},
	}

	for eNet, ep := range VSvc.(*Server).DstMgr.ExNetMap {
		data = append(data, &FwdEntry{
			CIDR:   eNet.String(),
			EPUser: ep.User,
			EPIP:   ep.IP.String(),
		})
	}

	c.JSON(http.StatusOK, data)
}
