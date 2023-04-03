package rpc

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func HandleInotifyEvent(c *gin.Context) {
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Info().Msg(string(data))
	}
	c.String(http.StatusOK, "success")
}
