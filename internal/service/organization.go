package service

import (
	"fmt"

	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/sirupsen/logrus"
)

// CreateOrganization creates a new organization.
func CreateOrganization(name, description string) (*model.Organization, error) {
	org := &model.Organization{
		Name:        name,
		Description: description,
	}
	if err := database.DB.Create(org).Error; err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"name": name,
		"id":   org.ID,
	}).Info("organization created")
	return org, nil
}

// UpdateOrganization updates an organization's name and description.
func UpdateOrganization(id uint, name, description string) (*model.Organization, error) {
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	updates := map[string]any{}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&org).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update organization: %w", err)
		}
		database.DB.First(&org, org.ID)
	}

	return &org, nil
}

// DeleteOrganization soft-deletes an organization by ID.
func DeleteOrganization(id uint) error {
	result := database.DB.Delete(&model.Organization{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete organization: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("organization not found")
	}
	logrus.WithField("id", id).Info("organization deleted")
	return nil
}

// ListOrganizations returns all organizations.
func ListOrganizations() ([]model.Organization, error) {
	var orgs []model.Organization
	if err := database.DB.Order("id ASC").Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	return orgs, nil
}

// FindOrganizationByID looks up an organization by its primary key.
func FindOrganizationByID(id uint) (*model.Organization, error) {
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}
	return &org, nil
}

// AddUserToOrg adds a user to an organization with the given role.
func AddUserToOrg(userID, orgID uint, role string) (*model.UserOrganization, error) {
	// Check if membership already exists
	var existing model.UserOrganization
	if err := database.DB.Where("user_id = ? AND organization_id = ?", userID, orgID).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("user is already a member of this organization")
	}

	uo := &model.UserOrganization{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
	}
	if err := database.DB.Create(uo).Error; err != nil {
		return nil, fmt.Errorf("add user to organization: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
		"role":    role,
	}).Info("user added to organization")
	return uo, nil
}

// RemoveUserFromOrg removes a user from an organization.
func RemoveUserFromOrg(userID, orgID uint) error {
	result := database.DB.Where("user_id = ? AND organization_id = ?", userID, orgID).Delete(&model.UserOrganization{})
	if result.Error != nil {
		return fmt.Errorf("remove user from organization: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user is not a member of this organization")
	}
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
	}).Info("user removed from organization")
	return nil
}

// UpdateUserOrgRole changes a user's role within an organization.
func UpdateUserOrgRole(userID, orgID uint, newRole string) error {
	result := database.DB.Model(&model.UserOrganization{}).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Update("role", newRole)
	if result.Error != nil {
		return fmt.Errorf("update user org role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user is not a member of this organization")
	}
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"org_id":  orgID,
		"role":    newRole,
	}).Info("user organization role updated")
	return nil
}

// FindUserOrganization returns the user-organization membership record.
func FindUserOrganization(userID, orgID uint) (*model.UserOrganization, error) {
	var uo model.UserOrganization
	if err := database.DB.Where("user_id = ? AND organization_id = ?", userID, orgID).First(&uo).Error; err != nil {
		return nil, fmt.Errorf("user organization membership not found")
	}
	return &uo, nil
}

// ListUserOrganizations returns all organizations a user belongs to.
func ListUserOrganizations(userID uint) ([]model.UserOrganization, error) {
	var memberships []model.UserOrganization
	if err := database.DB.Where("user_id = ?", userID).Order("id ASC").Find(&memberships).Error; err != nil {
		return nil, fmt.Errorf("list user organizations: %w", err)
	}
	return memberships, nil
}

// ListOrganizationUsers returns all user memberships for a given organization.
func ListOrganizationUsers(orgID uint) ([]model.UserOrganization, error) {
	var memberships []model.UserOrganization
	if err := database.DB.Where("organization_id = ?", orgID).Order("id ASC").Find(&memberships).Error; err != nil {
		return nil, fmt.Errorf("list organization users: %w", err)
	}
	return memberships, nil
}
