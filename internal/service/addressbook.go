package service

import (
	"fmt"

	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/sirupsen/logrus"
)

// AddressBookEntry represents a single contact in a user's address book.
type AddressBookEntry struct {
	PeerID           string   `json:"id"`
	Username         string   `json:"username"`
	Hostname         string   `json:"hostname"`
	Alias            string   `json:"alias"`
	Platform         string   `json:"platform"`
	Tags             []string `json:"tags"`
	ForceAlwaysRelay *bool    `json:"force_always_relay,omitempty"`
	Hash             string   `json:"hash"`
}

// AddressBookSyncPayload is the request body for full address book sync.
type AddressBookSyncPayload struct {
	Peers     []AddressBookEntry `json:"peers"`
	Tags      []string           `json:"tags"`
	TagColors []any              `json:"tag_colors"`
}

// ListAddressBooks returns paginated address book entries for a user.
func ListAddressBooks(userID uint, page, pageSize int, search string) ([]model.AddressBook, int64, error) {
	query := database.DB.Model(&model.AddressBook{}).Where("user_id = ?", userID)

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("hostname LIKE ? OR alias LIKE ? OR peer_id LIKE ?", s, s, s)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count address_books: %w", err)
	}

	var books []model.AddressBook
	offset := (page - 1) * pageSize
	if err := query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&books).Error; err != nil {
		return nil, 0, fmt.Errorf("list address_books: %w", err)
	}

	return books, total, nil
}

// ListAllAddressBooks returns all address book entries for a user (no pagination).
func ListAllAddressBooks(userID uint) ([]model.AddressBook, error) {
	var books []model.AddressBook
	if err := database.DB.Where("user_id = ?", userID).Order("id ASC").Find(&books).Error; err != nil {
		return nil, fmt.Errorf("list all address_books: %w", err)
	}
	return books, nil
}

// CreateAddressBook creates a new address book entry.
func CreateAddressBook(entry *model.AddressBook) error {
	if err := database.DB.Create(entry).Error; err != nil {
		return fmt.Errorf("create address_book: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"user_id": entry.UserID,
		"peer_id": entry.PeerID,
		"alias":   entry.Alias,
	}).Info("address book entry created")
	return nil
}

// UpdateAddressBook updates an existing address book entry.
func UpdateAddressBook(id uint, updates map[string]any) error {
	result := database.DB.Model(&model.AddressBook{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update address_book: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("address_book not found")
	}
	return nil
}

// DeleteAddressBooks deletes address book entries by their IDs for a given user.
func DeleteAddressBooks(userID uint, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	result := database.DB.Where("user_id = ? AND id IN ?", userID, ids).Delete(&model.AddressBook{})
	if result.Error != nil {
		return fmt.Errorf("delete address_books: %w", result.Error)
	}
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   result.RowsAffected,
	}).Info("address book entries deleted")
	return nil
}

// FindAddressBookByID looks up a single address book entry.
func FindAddressBookByID(id uint) (*model.AddressBook, error) {
	var book model.AddressBook
	if err := database.DB.First(&book, id).Error; err != nil {
		return nil, fmt.Errorf("address_book not found: %w", err)
	}
	return &book, nil
}

// SyncAddressBook replaces all address book entries and tags for a user
// with the provided peer entries and tags.
func SyncAddressBook(userID uint, peers []AddressBookEntry, tagNames []string, tagColors []any) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete existing address book entries for this user
	if err := tx.Where("user_id = ?", userID).Delete(&model.AddressBook{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete old address_books: %w", err)
	}

	// Insert new address book entries
	for _, p := range peers {
		tagsJSON := tagsToJSONString(p.Tags)
		forceRelay := false
		if p.ForceAlwaysRelay != nil {
			forceRelay = *p.ForceAlwaysRelay
		}

		entry := model.AddressBook{
			UserID:           userID,
			PeerID:           p.PeerID,
			Username:         p.Username,
			Hostname:         p.Hostname,
			Alias:            p.Alias,
			Platform:         p.Platform,
			Tags:             tagsJSON,
			ForceAlwaysRelay: forceRelay,
		}
		if err := tx.Create(&entry).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("create address_book entry: %w", err)
		}
	}

	// Delete existing tags for this user
	if err := tx.Where("user_id = ?", userID).Delete(&model.Tag{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete old tags: %w", err)
	}

	// Insert new tags
	for i, name := range tagNames {
		var color int64 = 0
		if i < len(tagColors) {
			switch v := tagColors[i].(type) {
			case float64:
				color = int64(v)
			case int64:
				color = v
			case int:
				color = int64(v)
			}
		}
		tag := model.Tag{
			UserID: userID,
			Name:   name,
			Color:  color,
		}
		if err := tx.Create(&tag).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("create tag: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit sync: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":     userID,
		"peers_count": len(peers),
		"tags_count":  len(tagNames),
	}).Info("address book synced")

	return nil
}

// tagsToJSONString converts a list of tag names to a JSON array string.
// Returns empty string if no tags, or JSON like ["tag1","tag2"].
func tagsToJSONString(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	// Manual JSON encoding to avoid importing encoding/json for such simple case
	result := "["
	for i, t := range tags {
		if i > 0 {
			result += ","
		}
		result += "\"" + t + "\""
	}
	result += "]"
	return result
}

// =============================================================================
// Tag service functions
// =============================================================================

// ListTags returns paginated tags for a user.
func ListTags(userID uint, page, pageSize int) ([]model.Tag, int64, error) {
	query := database.DB.Model(&model.Tag{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count tags: %w", err)
	}

	var tags []model.Tag
	offset := (page - 1) * pageSize
	if err := query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&tags).Error; err != nil {
		return nil, 0, fmt.Errorf("list tags: %w", err)
	}

	return tags, total, nil
}

// CreateTag creates a new tag.
func CreateTag(tag *model.Tag) error {
	if err := database.DB.Create(tag).Error; err != nil {
		return fmt.Errorf("create tag: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"user_id": tag.UserID,
		"name":    tag.Name,
	}).Info("tag created")
	return nil
}

// UpdateTag updates an existing tag.
func UpdateTag(id uint, userID uint, updates map[string]any) error {
	result := database.DB.Model(&model.Tag{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update tag: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("tag not found")
	}
	return nil
}

// DeleteTags deletes tags by their IDs for a given user.
func DeleteTags(userID uint, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	result := database.DB.Where("user_id = ? AND id IN ?", userID, ids).Delete(&model.Tag{})
	if result.Error != nil {
		return fmt.Errorf("delete tags: %w", result.Error)
	}
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   result.RowsAffected,
	}).Info("tags deleted")
	return nil
}

// FindTagByID looks up a single tag.
func FindTagByID(id uint) (*model.Tag, error) {
	var tag model.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}
	return &tag, nil
}

// =============================================================================
// Device Group service functions
// =============================================================================

// ListDeviceGroups returns all device groups.
func ListDeviceGroups() ([]model.DeviceGroup, error) {
	var groups []model.DeviceGroup
	if err := database.DB.Order("id ASC").Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("list device_groups: %w", err)
	}
	return groups, nil
}

// CreateDeviceGroup creates a new device group.
func CreateDeviceGroup(name string) (*model.DeviceGroup, error) {
	group := &model.DeviceGroup{Name: name}
	if err := database.DB.Create(group).Error; err != nil {
		return nil, fmt.Errorf("create device_group: %w", err)
	}
	logrus.WithField("name", name).Info("device group created")
	return group, nil
}

// UpdateDeviceGroup updates a device group's name.
func UpdateDeviceGroup(id uint, name string) error {
	result := database.DB.Model(&model.DeviceGroup{}).Where("id = ?", id).Update("name", name)
	if result.Error != nil {
		return fmt.Errorf("update device_group: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("device_group not found")
	}
	return nil
}

// DeleteDeviceGroup deletes a device group by ID.
func DeleteDeviceGroup(id uint) error {
	// Unset device_group_id on peers that reference this group
	database.DB.Model(&model.Peer{}).Where("device_group_id = ?", id).Update("device_group_id", nil)

	result := database.DB.Where("id = ?", id).Delete(&model.DeviceGroup{})
	if result.Error != nil {
		return fmt.Errorf("delete device_group: %w", result.Error)
	}
	logrus.WithField("id", id).Info("device group deleted")
	return nil
}

// FindDeviceGroupByID looks up a device group.
func FindDeviceGroupByID(id uint) (*model.DeviceGroup, error) {
	var group model.DeviceGroup
	if err := database.DB.First(&group, id).Error; err != nil {
		return nil, fmt.Errorf("device_group not found: %w", err)
	}
	return &group, nil
}
