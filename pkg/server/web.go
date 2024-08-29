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
	ExNet   string
	LoginAt string
}

func getEpForwardMap() map[string]string {
	data := make(map[string]string)

	for eNet, ep := range VSvc.(*Server).DstMgr.ExNetMap {
		eData, ok := data[ep.User]
		if ok {
			eData += fmt.Sprintf(",%s", eNet.String())
			data[ep.User] = eData
		} else {
			data[ep.User] = eNet.String()
		}
	}

	return data
}

func listEndpoints(c *gin.Context) {
	exNetMap := getEpForwardMap()
	var data []*EPEntry

	for ipKey := range VSvc.(*Server).EPMgr.EPIPMap {
		ep := VSvc.(*Server).EPMgr.GetEPByIP(ipKey)

		epEn := &EPEntry{
			User:    ep.User,
			Addr:    ep.RAddr.String(),
			IP:      ep.IP.String(),
			LoginAt: ep.LoginTime,
			ExNet:   exNetMap[ep.User],
		}

		data = append(data, epEn)
	}

	c.JSON(http.StatusOK, data)
}
