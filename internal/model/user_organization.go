package model

import (
	"time"
)

// UserOrganization represents a user's membership in an organization with a role.
// Each user can be a member of multiple organizations.
type UserOrganization struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"uniqueIndex:idx_user_org;not null" json:"user_id"`
	OrganizationID uint      `gorm:"uniqueIndex:idx_user_org;not null" json:"organization_id"`
	Role           string    `gorm:"not null;size:50;default:org_member" json:"role"` // org_admin, org_member, org_auditor
	CreatedAt      time.Time `json:"created_at"`
}

// TableName specifies the table name for UserOrganization.
func (UserOrganization) TableName() string {
	return "user_organizations"
}
