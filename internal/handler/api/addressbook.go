package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rustdesk/rustdesk-api-server/internal/database"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"github.com/rustdesk/rustdesk-api-server/internal/service"
	"github.com/sirupsen/logrus"
)

// ABGetResponse is the response for GET /api/ab.
type ABGetResponse struct {
	Data []ABEntryResponse `json:"data"`
}

// ABEntryResponse represents a single address book entry in the client API format.
type ABEntryResponse struct {
	ID               uint     `json:"id"`
	PeerID           string   `json:"peer_id"`
	Username         string   `json:"username"`
	Hostname         string   `json:"hostname"`
	Alias            string   `json:"alias"`
	Platform         string   `json:"platform"`
	Tags             []string `json:"tags"`
	ForceAlwaysRelay bool     `json:"force_always_relay"`
}

// ABSyncRequest is the request body for POST /api/ab (full sync).
type ABSyncRequest struct {
	Peers     []ABSyncPeer      `json:"peers"`
	Tags      []string          `json:"tags"`
	TagColors []json.RawMessage `json:"tag_colors"`
}

// ABSyncPeer is a single peer entry in the sync request.
type ABSyncPeer struct {
	PeerID           string   `json:"id"`
	Username         string   `json:"username"`
	Hostname         string   `json:"hostname"`
	Alias            string   `json:"alias"`
	Platform         string   `json:"platform"`
	Tags             []string `json:"tags"`
	ForceAlwaysRelay *bool    `json:"force_always_relay"`
	Hash             string   `json:"hash"`
}

// ABGet handles GET /api/ab
// Returns the full address book for the authenticated user.
func ABGet(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	books, err := service.ListAllAddressBooks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to load address book",
		})
		return
	}

	data := make([]ABEntryResponse, 0, len(books))
	for _, b := range books {
		tags := parseTagsFromJSON(b.Tags)
		data = append(data, ABEntryResponse{
			ID:               b.ID,
			PeerID:           b.PeerID,
			Username:         b.Username,
			Hostname:         b.Hostname,
			Alias:            b.Alias,
			Platform:         b.Platform,
			Tags:             tags,
			ForceAlwaysRelay: b.ForceAlwaysRelay,
		})
	}

	c.JSON(http.StatusOK, ABGetResponse{Data: data})
}

// ABSync handles POST /api/ab (full address book sync)
// Replaces all address book entries and tags for the user.
func ABSync(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	var req ABSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid sync payload",
		})
		return
	}

	// Convert ABSyncPeer to service.AddressBookEntry
	peers := make([]service.AddressBookEntry, len(req.Peers))
	for i, p := range req.Peers {
		peers[i] = service.AddressBookEntry{
			PeerID:           p.PeerID,
			Username:         p.Username,
			Hostname:         p.Hostname,
			Alias:            p.Alias,
			Platform:         p.Platform,
			Tags:             p.Tags,
			ForceAlwaysRelay: p.ForceAlwaysRelay,
			Hash:             p.Hash,
		}
	}

	// Convert tag_colors to []any
	tagColors := make([]any, len(req.TagColors))
	for i, tc := range req.TagColors {
		var v any
		if err := json.Unmarshal(tc, &v); err == nil {
			tagColors[i] = v
		} else {
			tagColors[i] = int64(0)
		}
	}

	if err := service.SyncAddressBook(userID, peers, req.Tags, tagColors); err != nil {
		logrus.WithError(err).Error("failed to sync address book")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to sync address book",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "synced"})
}

// ABPersonal handles POST /api/ab/personal
// Returns the personal address book GUID for the user.
func ABPersonal(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"guid": "personal",
	})
}

// ABSettings handles POST /api/ab/settings
// Returns address book settings.
func ABSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"max_peer": 100,
		"strategy": map[string]any{},
	})
}

// ABSharedProfiles handles POST /api/ab/shared/profiles
// Returns list of shared address book profiles.
func ABSharedProfiles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []any{}})
}

// ABPeers handles POST /api/ab/peers
// Returns peers in a shared address book.
func ABPeers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []any{}})
}

// ABTags handles POST /api/ab/tags/:guid
// Returns tags belonging to the specified address book.
func ABTags(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	tags, _, err := service.ListTags(userID, 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to list tags",
		})
		return
	}

	data := make([]gin.H, len(tags))
	for i, t := range tags {
		data[i] = gin.H{
			"id":    t.ID,
			"name":  t.Name,
			"color": t.Color,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// ABPeerAddRequest is the request body for adding a peer to an address book.
type ABPeerAddRequest struct {
	PeerID           string   `json:"id"`
	Username         string   `json:"username"`
	Hostname         string   `json:"hostname"`
	Alias            string   `json:"alias"`
	Platform         string   `json:"platform"`
	Tags             []string `json:"tags"`
	ForceAlwaysRelay bool     `json:"force_always_relay"`
	Hash             string   `json:"hash"`
}

// ABPeerAdd handles POST /api/ab/peer/add/:guid
func ABPeerAdd(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	var req ABPeerAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
		})
		return
	}

	tagsJSON := tagsToJSON(req.Tags)
	entry := &model.AddressBook{
		UserID:           userID,
		PeerID:           req.PeerID,
		Username:         req.Username,
		Hostname:         req.Hostname,
		Alias:            req.Alias,
		Platform:         req.Platform,
		Tags:             tagsJSON,
		ForceAlwaysRelay: req.ForceAlwaysRelay,
	}

	if err := service.CreateAddressBook(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to add peer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "added"})
}

// ABPeerDelete handles DELETE /api/ab/peer/:guid
func ABPeerDelete(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	peerID := c.Param("guid")

	result := database.DB.Where("user_id = ? AND peer_id = ?", userID, peerID).Delete(&model.AddressBook{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Peer not found in address book",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ABPeerUpdateRequest is the request body for updating an address book peer.
type ABPeerUpdateRequest struct {
	Alias            string   `json:"alias"`
	Tags             []string `json:"tags"`
	ForceAlwaysRelay bool     `json:"force_always_relay"`
	Hash             string   `json:"hash"`
}

// ABPeerUpdate handles PUT /api/ab/peer/update/:guid
func ABPeerUpdate(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	peerID := c.Param("guid")

	var req ABPeerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
		})
		return
	}

	updates := map[string]any{
		"alias":              req.Alias,
		"tags":               tagsToJSON(req.Tags),
		"force_always_relay": req.ForceAlwaysRelay,
	}

	result := database.DB.Model(&model.AddressBook{}).
		Where("user_id = ? AND peer_id = ?", userID, peerID).
		Updates(updates)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Peer not found in address book",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// ABTagAddRequest is the request body for adding a tag.
type ABTagAddRequest struct {
	Name  string `json:"name"`
	Color int64  `json:"color"`
}

// ABTagAdd handles POST /api/ab/tag/add/:guid
func ABTagAdd(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	var req ABTagAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
		})
		return
	}

	tag := &model.Tag{
		UserID: userID,
		Name:   req.Name,
		Color:  req.Color,
	}

	if err := service.CreateTag(tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to add tag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "added", "id": tag.ID})
}

// ABTagRenameRequest is the request body for renaming a tag.
type ABTagRenameRequest struct {
	Name string `json:"name"`
}

// ABTagRename handles PUT /api/ab/tag/rename/:guid
func ABTagRename(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	tagIDStr := c.Param("guid")
	var req ABTagRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
		})
		return
	}

	// tag_id might be numeric — try both numeric and string lookup
	var tag model.Tag
	if err := database.DB.Where("name = ? AND user_id = ?", tagIDStr, userID).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Tag not found",
		})
		return
	}

	database.DB.Model(&tag).Update("name", req.Name)

	c.JSON(http.StatusOK, gin.H{"message": "renamed"})
}

// ABTagUpdateColorRequest is the request body for updating a tag's color.
type ABTagUpdateColorRequest struct {
	Color int64 `json:"color"`
}

// ABTagUpdate handles PUT /api/ab/tag/update/:guid
func ABTagUpdate(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	tagIDStr := c.Param("guid")
	var req ABTagUpdateColorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
		})
		return
	}

	var tag model.Tag
	if err := database.DB.Where("name = ? AND user_id = ?", tagIDStr, userID).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Tag not found",
		})
		return
	}

	database.DB.Model(&tag).Update("color", req.Color)

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// ABTagDelete handles DELETE /api/ab/tag/:guid
func ABTagDelete(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	tagName := c.Param("guid")

	result := database.DB.Where("name = ? AND user_id = ?", tagName, userID).Delete(&model.Tag{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Tag not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// =============================================================================
// Helpers
// =============================================================================

// getUserIDFromContext extracts the user ID from the gin context.
// Returns 0 if not authenticated.
func getUserIDFromContext(c *gin.Context) uint {
	// For client API, we need to extract user info from context
	// or from the Authorization header via RustAuth middleware.
	// Currently, the client AB endpoints are routed without auth middleware.
	// In Phase 1, these could be public/dev endpoints.
	//
	// For now, return 0 to indicate unauthenticated.
	// Extend this when RustAuth middleware is fully implemented.

	// Try to get from RustAuth context
	if user, exists := c.Get("rustUser"); exists {
		if u, ok := user.(*model.User); ok {
			return u.ID
		}
	}

	return 0
}

// parseTagsFromJSON parses a JSON array string like '["tag1","tag2"]' into a string slice.
func parseTagsFromJSON(tagsJSON string) []string {
	if tagsJSON == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
		return []string{}
	}
	return tags
}

// tagsToJSON converts a string slice to a JSON array string.
func tagsToJSON(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	data, err := json.Marshal(tags)
	if err != nil {
		return ""
	}
	return string(data)
}
