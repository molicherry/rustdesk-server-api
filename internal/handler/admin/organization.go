package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
	"github.com/sirupsen/logrus"
)

// OrgResponse is the public organization representation.
type OrgResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func orgResponse(o *model.Organization) OrgResponse {
	return OrgResponse{
		ID:          o.ID,
		Name:        o.Name,
		Description: o.Description,
		CreatedAt:   o.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   o.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// OrgUserResponse is the public organization user membership representation.
type OrgUserResponse struct {
	ID             uint   `json:"id"`
	UserID         uint   `json:"user_id"`
	Username       string `json:"username"`
	OrganizationID uint   `json:"organization_id"`
	Role           string `json:"role"`
	CreatedAt      string `json:"created_at"`
}

func orgUserResponse(uo *model.UserOrganization) OrgUserResponse {
	username := ""
	var user model.User
	if err := database.DB.First(&user, uo.UserID).Error; err == nil {
		username = user.Username
	}
	return OrgUserResponse{
		ID:             uo.ID,
		UserID:         uo.UserID,
		Username:       username,
		OrganizationID: uo.OrganizationID,
		Role:           uo.Role,
		CreatedAt:      uo.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// =============================================================================
// Organization CRUD
// =============================================================================

// ListOrganizations handles GET /api/admin/organizations/list
func ListOrganizations(c *gin.Context) {
	orgs, err := service.ListOrganizations()
	if err != nil {
		logrus.WithError(err).Error("list organizations failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	if orgs == nil {
		orgs = []model.Organization{}
	}

	c.JSON(http.StatusOK, gin.H{"data": orgs})
}

// CreateOrganizationRequest is the request body for creating an organization.
type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CreateOrganization handles POST /api/admin/organizations/create
func CreateOrganization(c *gin.Context) {
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "name is required",
		})
		return
	}

	org, err := service.CreateOrganization(req.Name, req.Description)
	if err != nil {
		logrus.WithError(err).Error("create organization failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, orgResponse(org))
}

// UpdateOrganizationRequest is the request body for updating an organization.
type UpdateOrganizationRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateOrganization handles POST /api/admin/organizations/update
func UpdateOrganization(c *gin.Context) {
	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "id is required",
		})
		return
	}

	org, err := service.UpdateOrganization(req.ID, req.Name, req.Description)
	if err != nil {
		logrus.WithError(err).Error("update organization failed")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, orgResponse(org))
}

// DeleteOrganizationRequest is the request body for deleting an organization.
type DeleteOrganizationRequest struct {
	ID uint `json:"id" binding:"required"`
}

// DeleteOrganization handles POST /api/admin/organizations/delete
func DeleteOrganization(c *gin.Context) {
	var req DeleteOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "id is required",
		})
		return
	}

	if err := service.DeleteOrganization(req.ID); err != nil {
		logrus.WithError(err).Error("delete organization failed")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Organization deleted successfully"})
}

// =============================================================================
// Organization User Management
// =============================================================================

// ListOrganizationUsers handles GET /api/admin/organizations/:orgID/users/list
func ListOrganizationUsers(c *gin.Context) {
	orgIDStr := c.Param("orgID")
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid organization ID",
		})
		return
	}

	memberships, err := service.ListOrganizationUsers(uint(orgID))
	if err != nil {
		logrus.WithError(err).Error("list organization users failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "internal server error",
		})
		return
	}

	if memberships == nil {
		memberships = []model.UserOrganization{}
	}

	data := make([]OrgUserResponse, len(memberships))
	for i, m := range memberships {
		data[i] = orgUserResponse(&m)
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// AddUserToOrgRequest is the request body for adding a user to an organization.
type AddUserToOrgRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

// AddUserToOrganization handles POST /api/admin/organizations/:orgID/users/add
func AddUserToOrganization(c *gin.Context) {
	orgIDStr := c.Param("orgID")
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid organization ID",
		})
		return
	}

	var req AddUserToOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "user_id and role are required",
		})
		return
	}

	uo, err := service.AddUserToOrg(req.UserID, uint(orgID), req.Role)
	if err != nil {
		logrus.WithError(err).Error("add user to organization failed")
		c.JSON(http.StatusConflict, gin.H{
			"error":   "conflict",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, orgUserResponse(uo))
}

// RemoveUserFromOrgRequest is the request body for removing a user from an organization.
type RemoveUserFromOrgRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

// RemoveUserFromOrganization handles POST /api/admin/organizations/:orgID/users/remove
func RemoveUserFromOrganization(c *gin.Context) {
	orgIDStr := c.Param("orgID")
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid organization ID",
		})
		return
	}

	var req RemoveUserFromOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "user_id is required",
		})
		return
	}

	if err := service.RemoveUserFromOrg(req.UserID, uint(orgID)); err != nil {
		logrus.WithError(err).Error("remove user from organization failed")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from organization successfully"})
}

// UpdateUserOrgRoleRequest is the request body for changing a user's role in an organization.
type UpdateUserOrgRoleRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

// UpdateUserOrgRole handles POST /api/admin/organizations/:orgID/users/update-role
func UpdateUserOrgRole(c *gin.Context) {
	orgIDStr := c.Param("orgID")
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid organization ID",
		})
		return
	}

	var req UpdateUserOrgRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "user_id and role are required",
		})
		return
	}

	if err := service.UpdateUserOrgRole(req.UserID, uint(orgID), req.Role); err != nil {
		logrus.WithError(err).Error("update user org role failed")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}
