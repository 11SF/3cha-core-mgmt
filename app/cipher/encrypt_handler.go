package cipher

import (
	"portal/backend/pkg/aesgcm"
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type EncryptRequest struct {
	Text string `json:"text" binding:"required"`
	Key  string `json:"key" binding:"required,len=16|len=24|len=32"`
}

func (h *handler) Encrypt(c *gin.Context) {
	var req EncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	result, err := aesgcm.EncryptGCM(req.Text, req.Key)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.OK(c, CipherResponse{Result: result})
}
