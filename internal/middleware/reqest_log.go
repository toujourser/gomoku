package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/toujourser/gomoku/pkg/logger"
	"github.com/toujourser/gomoku/pkg/mongodb"
	"time"
)

func RequestLogger() gin.HandlerFunc {
	mgoHook, err := mongodb.NewHooker(viper.GetString("mongodb.addr"), viper.GetString("mongodb.db"), viper.GetString("mongodb.collection"))
	if err == nil {
		logger.AddHook(mgoHook)
	}

	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		reqProto := c.Request.Proto
		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		fields := logrus.Fields{
			"client_ip":   clientIP,
			"protocol":    reqProto,
			"method":      reqMethod,
			"uri":         reqUri,
			"status_code": statusCode,
			"start_time":  startTime,
		}
		logger.WithFields(fields).Info()
	}
}
