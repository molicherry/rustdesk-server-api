package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

// AddressBookResponse is the public address book entry representation.
type AddressBookResponse struct {
	ID               uint   `json:"id"`
	UserID           uint   `json:"user_id"`
	PeerID           string `json:"peer_id"`
	Username         string `json:"username"`
	Hostname         string `json:"hostname"`
	Alias            string `json:"alias"`
	Platform         string `json:"platform"`
	Tags             string `json:"tags"`
	ForceAlwaysRelay bool   `json:"force_always_relay"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

func addressBookResponse(ab *model.AddressBook) AddressBookResponse {
	return AddressBookResponse{
		ID:               ab.ID,
		UserID:           ab.UserID,
		PeerID:           ab.PeerID,
		Username:         ab.Username,
		Hostname:         ab.Hostname,
		Alias:            ab.Alias,
		Platform:         ab.Platform,
		Tags:             ab.Tags,
		ForceAlwaysRelay: ab.ForceAlwaysRelay,
		CreatedAt:        ab.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        ab.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// AddressBookListRequest holds query parameters for address book listing.
type AddressBookListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
	UserID   uint   `form:"user_id"`
}

// AddressBookListResponse is the paginated list response.
type AddressBookListResponse struct {
	Total int64                 `json:"total"`
	Data  []AddressBookResponse `json:"data"`
}

// ListAddressBooks handles GET /api/admin/address_book/list
func ListAddressBooks(c *gin.Context) {
	var req AddressBookListRequest
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

	query := database.DB.Model(&model.AddressBook{})

	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.Search != "" {
		s := "%" + req.Search + "%"
		query = query.Where("hostname LIKE ? OR alias LIKE ? OR peer_id LIKE ? OR username LIKE ?", s, s, s, s)
	}

	var total int64
	query.Count(&total)

	var books []model.AddressBook
	offset := (req.Page - 1) * req.PageSize
	query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&books)

	data := make([]AddressBookResponse, len(books))
	for i, b := range books {
		data[i] = addressBookResponse(&b)
	}

	c.JSON(http.StatusOK, AddressBookListResponse{
		Total: total,
		Data:  data,
	})
}

// CreateAddressBookRequest is the request body for creating an address book entry.
type CreateAddressBookRequest struct {
	PeerID           string `json:"peer_id" binding:"required"`
	Username         string `json:"username"`
	Hostname         string `json:"hostname"`
	Alias            string `json:"alias"`
	Platform         string `json:"platform"`
	Tags             string `json:"tags"`
	ForceAlwaysRelay bool   `json:"force_always_relay"`
	UserID           uint   `json:"user_id" binding:"required"`
}

// CreateAddressBook handles POST /api/admin/address_book/create
func CreateAddressBook(c *gin.Context) {
	var req CreateAddressBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "peer_id and user_id are required",
		})
		return
	}

	entry := &model.AddressBook{
		UserID:           req.UserID,
		PeerID:           req.PeerID,
		Username:         req.Username,
		Hostname:         req.Hostname,
		Alias:            req.Alias,
		Platform:         req.Platform,
		Tags:             req.Tags,
		ForceAlwaysRelay: req.ForceAlwaysRelay,
	}

	if err := service.CreateAddressBook(entry); err != nil {
		logrus.WithError(err).Error("create address book entry failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, addressBookResponse(entry))
}

// UpdateAddressBookRequest is the request body for updating an address book entry.
type UpdateAddressBookRequest struct {
	ID               uint   `json:"id" binding:"required"`
	Alias            string `json:"alias"`
	Tags             string `json:"tags"`
	ForceAlwaysRelay *bool  `json:"force_always_relay"`
	Username         string `json:"username"`
	Hostname         string `json:"hostname"`
	Platform         string `json:"platform"`
}

// UpdateAddressBook handles POST /api/admin/address_book/update
func UpdateAddressBook(c *gin.Context) {
	var req UpdateAddressBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "id is required",
		})
		return
	}

	updates := map[string]any{}
	if req.Alias != "" {
		updates["alias"] = req.Alias
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.ForceAlwaysRelay != nil {
		updates["force_always_relay"] = *req.ForceAlwaysRelay
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Hostname != "" {
		updates["hostname"] = req.Hostname
	}
	if req.Platform != "" {
		updates["platform"] = req.Platform
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "no fields to update",
		})
		return
	}

	if err := service.UpdateAddressBook(req.ID, updates); err != nil {
		logrus.WithError(err).Error("update address book entry failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	// Fetch updated entry
	book, _ := service.FindAddressBookByID(req.ID)
	if book != nil {
		c.JSON(http.StatusOK, addressBookResponse(book))
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Updated"})
	}
}

// DeleteAddressBookRequest is the request body for deleting address book entries.
type DeleteAddressBookRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// DeleteAddressBook handles POST /api/admin/address_book/delete
func DeleteAddressBook(c *gin.Context) {
	var req DeleteAddressBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "ids is required (array of address book IDs)",
		})
		return
	}

	// Admin can delete any user's address book entries
	if err := database.DB.Where("id IN ?", req.IDs).Delete(&model.AddressBook{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to delete address book entries",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Address book entries deleted successfully"})
}

// GetAddressBook handles GET /api/admin/address_book/detail/:id
func GetAddressBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid ID",
		})
		return
	}

	book, err := service.FindAddressBookByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Address book entry not found",
		})
		return
	}

	c.JSON(http.StatusOK, addressBookResponse(book))
}
