package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

// TagResponse is the public tag representation.
type TagResponse struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Name      string `json:"name"`
	Color     int64  `json:"color"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func tagResponse(t *model.Tag) TagResponse {
	return TagResponse{
		ID:        t.ID,
		UserID:    t.UserID,
		Name:      t.Name,
		Color:     t.Color,
		CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// TagListRequest holds query parameters for tag listing.
type TagListRequest struct {
	Page     int  `form:"page"`
	PageSize int  `form:"page_size"`
	UserID   uint `form:"user_id"`
}

// TagListResponse is the paginated list response.
type TagListResponse struct {
	Total int64         `json:"total"`
	Data  []TagResponse `json:"data"`
}

// ListTags handles GET /api/admin/tag/list
func ListTags(c *gin.Context) {
	var req TagListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// If no user_id specified, list all tags
	tags, total, err := service.ListTags(req.UserID, req.Page, req.PageSize)
	if err != nil {
		logrus.WithError(err).Error("list tags failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	data := make([]TagResponse, len(tags))
	for i, t := range tags {
		data[i] = tagResponse(&t)
	}

	c.JSON(http.StatusOK, TagListResponse{
		Total: total,
		Data:  data,
	})
}

// CreateTagRequest is the request body for creating a tag.
type CreateTagRequest struct {
	Name   string `json:"name" binding:"required"`
	Color  int64  `json:"color"`
	UserID uint   `json:"user_id" binding:"required"`
}

// CreateTag handles POST /api/admin/tag/create
func CreateTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "name and user_id are required",
		})
		return
	}

	tag := &model.Tag{
		UserID: req.UserID,
		Name:   req.Name,
		Color:  req.Color,
	}

	if err := service.CreateTag(tag); err != nil {
		logrus.WithError(err).Error("create tag failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, tagResponse(tag))
}

// UpdateTagRequest is the request body for updating a tag.
type UpdateTagRequest struct {
	ID    uint   `json:"id" binding:"required"`
	Name  string `json:"name"`
	Color *int64 `json:"color"`
}

// UpdateTag handles POST /api/admin/tag/update
func UpdateTag(c *gin.Context) {
	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "id is required",
		})
		return
	}

	// Find the tag to get its user_id
	tag, err := service.FindTagByID(req.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Tag not found",
		})
		return
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "no fields to update",
		})
		return
	}

	if err := service.UpdateTag(req.ID, tag.UserID, updates); err != nil {
		logrus.WithError(err).Error("update tag failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	// Fetch updated tag
	tag, _ = service.FindTagByID(req.ID)
	if tag != nil {
		c.JSON(http.StatusOK, tagResponse(tag))
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Updated"})
	}
}

// DeleteTagRequest is the request body for deleting tags.
type DeleteTagRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// DeleteTag handles POST /api/admin/tag/delete
func DeleteTag(c *gin.Context) {
	var req DeleteTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "ids is required (array of tag IDs)",
		})
		return
	}

	// Admin can delete any tags — use id-based delete without user_id filter
	if err := service.DeleteTags(0, req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to delete tags",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tags deleted successfully"})
}

// GetTag handles GET /api/admin/tag/detail/:id
func GetTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid tag ID",
		})
		return
	}

	tag, err := service.FindTagByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Tag not found",
		})
		return
	}

	c.JSON(http.StatusOK, tagResponse(tag))
}
