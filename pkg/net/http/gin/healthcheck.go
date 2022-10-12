package gin

import (
	"github.com/gin-gonic/gin"
	types "github.com/gogo/protobuf/types"
	"net/http"
)

func healthcheck(c *gin.Context) {
	resp := TOJSON(&types.Empty{}, nil)
	c.JSON(http.StatusOK, resp)
}
