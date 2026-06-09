package member

import (
	"portal/backend/app/member/access"
	"portal/backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReorderMembersRequest struct {
	Orders []struct {
		ID        uuid.UUID `json:"id" binding:"required"`
		SortOrder int       `json:"sortOrder"`
	} `json:"orders" binding:"required"`
}

func (h *handler) Reorder(c *gin.Context) {
	ctx := c.Request.Context()

	var req ReorderMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	updates := make([]access.SortOrderUpdate, len(req.Orders))
	for i, o := range req.Orders {
		updates[i] = access.SortOrderUpdate{ID: o.ID, SortOrder: o.SortOrder}
	}

	if err := h.memberStorage.BulkUpdateSortOrder(ctx, updates); err != nil {
		response.InternalError(c, err)
		return
	}

	members, err := h.memberStorage.List(ctx)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, members)
}
