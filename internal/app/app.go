package app

import (
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/spf13/viper"
	"github.com/toujourser/gomoku/internal/middleware"
	"github.com/toujourser/gomoku/internal/websocket"
)

func InitServer() {
	router := gin.New()
	router.Use(middleware.RequestLogger(), gin.Recovery())
	m := melody.New()
	ms := &websocket.MelodySocket{M: m}
	m.HandleMessage(ms.Receive)
	m.HandleConnect(ms.Connect)
	m.HandleDisconnect(ms.Disconnect)
	router.GET("/ws", func(c *gin.Context) {
		_ = m.HandleRequest(c.Writer, c.Request)
	})
	_ = router.Run(viper.GetString("server.port"))
}
