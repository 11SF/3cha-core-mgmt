package cipher

import (
	"portal/backend/pkg/aesgcm"
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type DecryptRequest struct {
	Text string `json:"text" binding:"required"`
	Key  string `json:"key" binding:"required,len=16|len=24|len=32"`
}

func (h *handler) Decrypt(c *gin.Context) {
	var req DecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	result, err := aesgcm.DecryptGCM(req.Text, req.Key)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.OK(c, CipherResponse{Result: result})
}
