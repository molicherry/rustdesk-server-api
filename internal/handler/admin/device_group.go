package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

// DeviceGroupResponse is the public device group representation.
type DeviceGroupResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	PeerCount int64  `json:"peer_count"`
}

// ListDeviceGroups handles GET /api/admin/device_group/list
func ListDeviceGroups(c *gin.Context) {
	groups, err := service.ListDeviceGroups()
	if err != nil {
		logrus.WithError(err).Error("list device groups failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	if groups == nil {
		groups = []model.DeviceGroup{}
	}

	c.JSON(http.StatusOK, gin.H{"data": groups})
}

// CreateDeviceGroupRequest is the request body for creating a device group.
type CreateDeviceGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateDeviceGroup handles POST /api/admin/device_group/create
func CreateDeviceGroup(c *gin.Context) {
	var req CreateDeviceGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "name is required",
		})
		return
	}

	group, err := service.CreateDeviceGroup(req.Name)
	if err != nil {
		logrus.WithError(err).Error("create device group failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, group)
}

// UpdateDeviceGroupRequest is the request body for updating a device group.
type UpdateDeviceGroupRequest struct {
	ID   uint   `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// UpdateDeviceGroup handles POST /api/admin/device_group/update
func UpdateDeviceGroup(c *gin.Context) {
	var req UpdateDeviceGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "id and name are required",
		})
		return
	}

	if err := service.UpdateDeviceGroup(req.ID, req.Name); err != nil {
		logrus.WithError(err).Error("update device group failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	group, _ := service.FindDeviceGroupByID(req.ID)
	if group != nil {
		c.JSON(http.StatusOK, group)
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Updated"})
	}
}

// DeleteDeviceGroupRequest is the request body for deleting a device group.
type DeleteDeviceGroupRequest struct {
	ID uint `json:"id" binding:"required"`
}

// DeleteDeviceGroup handles POST /api/admin/device_group/delete
func DeleteDeviceGroup(c *gin.Context) {
	var req DeleteDeviceGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "id is required",
		})
		return
	}

	if err := service.DeleteDeviceGroup(req.ID); err != nil {
		logrus.WithError(err).Error("delete device group failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device group deleted"})
}

// GetDeviceGroup handles GET /api/admin/device_group/detail/:id
func GetDeviceGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid device group ID",
		})
		return
	}

	group, err := service.FindDeviceGroupByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Device group not found",
		})
		return
	}

	c.JSON(http.StatusOK, group)
}
