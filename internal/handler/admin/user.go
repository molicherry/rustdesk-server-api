package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/middleware"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
)

// UserListRequest holds query parameters for the user list endpoint.
type UserListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
}

// UserListResponse is the response body for the user list endpoint.
type UserListResponse struct {
	Total int64          `json:"total"`
	Data  []UserResponse `json:"data"`
}

// ListUsers handles GET /api/admin/user/list
func ListUsers(c *gin.Context) {
	var req UserListRequest
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

	query := database.DB.Model(&model.User{})

	// Search by username or email
	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("username LIKE ? OR email LIKE ?", search, search)
	}

	var total int64
	query.Count(&total)

	var users []model.User
	offset := (req.Page - 1) * req.PageSize
	query.Order("id ASC").Offset(offset).Limit(req.PageSize).Find(&users)

	data := make([]UserResponse, len(users))
	for i, u := range users {
		data[i] = userResponse(&u)
	}

	c.JSON(http.StatusOK, UserListResponse{
		Total: total,
		Data:  data,
	})
}

// GetUser handles GET /api/admin/user/detail/:id
func GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid user ID",
		})
		return
	}

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, userResponse(&user))
}

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
	Nickname string `json:"nickname"`
}

// CreateUser handles POST /api/admin/user/create
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Username and password are required",
		})
		return
	}

	user, err := service.CreateUser(req.Username, req.Password, req.IsAdmin)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "conflict",
			"message": err.Error(),
		})
		return
	}

	// Update optional fields
	updates := map[string]any{}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if len(updates) > 0 {
		database.DB.Model(user).Updates(updates)
		// Re-fetch to get updated fields
		database.DB.First(user, user.ID)
	}

	c.JSON(http.StatusOK, userResponse(user))
}

// UpdateUserRequest is the request body for updating a user.
type UpdateUserRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	IsAdmin  *bool  `json:"is_admin"`
	Status   *int   `json:"status"`
}

// UpdateUser handles POST /api/admin/user/update
func UpdateUser(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "User ID is required",
		})
		return
	}

	// Get current user from context to prevent self-demotion
	currentUser, _ := c.Get(middleware.ContextKeyUser)
	if currentUser != nil {
		cu := currentUser.(*model.User)
		if cu.ID == req.ID && req.IsAdmin != nil && !*req.IsAdmin {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "Cannot remove your own admin privileges",
			})
			return
		}
	}

	var user model.User
	if err := database.DB.First(&user, req.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "User not found",
		})
		return
	}

	updates := map[string]any{}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.IsAdmin != nil {
		updates["is_admin"] = *req.IsAdmin
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		database.DB.Model(&user).Updates(updates)
		// Refresh
		database.DB.First(&user, user.ID)
	}

	c.JSON(http.StatusOK, userResponse(&user))
}

// DeleteUserRequest is the request body for deleting a user.
type DeleteUserRequest struct {
	ID uint `json:"id" binding:"required"`
}

// DeleteUser handles POST /api/admin/user/delete
func DeleteUser(c *gin.Context) {
	var req DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "User ID is required",
		})
		return
	}

	// Prevent deleting self
	currentUser, exists := c.Get(middleware.ContextKeyUser)
	if exists {
		cu := currentUser.(*model.User)
		if cu.ID == req.ID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "Cannot delete your own account",
			})
			return
		}
	}

	var user model.User
	if err := database.DB.First(&user, req.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "User not found",
		})
		return
	}

	// Cascade: delete tokens and address books
	if err := service.DeleteUserTokens(req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to delete user tokens",
		})
		return
	}

	database.DB.Where("user_id = ?", req.ID).Delete(&model.AddressBook{})
	database.DB.Where("user_id = ?", req.ID).Delete(&model.Tag{})
	database.DB.Where("id = ?", req.ID).Delete(&model.User{})

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
