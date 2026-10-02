package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kr/pretty"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/models"
	"gorm.io/gorm"
	"fmt"
)

func (h *Handler) UpdatePermission(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing ID",
			Code:    "VALIDATION_ERROR",
			Details: "ID is required in URL path",
		})
		return
	}

	var input models.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}
	pretty.Print(input, "input")
	var permission models.Permissions
	if err := db.First(&permission, "id = ?", id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Permission not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	// Update basic fields
	// permission.Roles = input.Roles

	if err := db.Save(&permission); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update permission",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// 🔹 Update UserGroups (many2many)
	var groupIDs []int
	for _, g := range input.UserGroups {
		groupIDs = append(groupIDs, g.ID)
	}

	var groups []models.UserGroups
	if len(groupIDs) > 0 {
		if err := db.Where("id IN ?", groupIDs).Find(&groups); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to fetch user groups",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	if err := db.Model(&permission).Association("UserGroups").Replace(groups); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update user groups",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// 🔹 Update Users (Agents)
	// FIX: Agent.ID is string (varchar(27)), not int
	var userIDs []string
	for _, u := range input.User {
		userIDs = append(userIDs, u.ID)
	}

	var agents []models.Agent
	if len(userIDs) > 0 {
		if err := db.Where("id IN ?", userIDs).Find(&agents); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to fetch users",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	if err := db.Model(&permission).Association("User").Replace(agents); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update users",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// 🔹 Update show_prices (optional, independent of permission itself)
	if input.ShowPricesUpdate != nil && len(input.ShowPricesUpdate.AgentIDs) > 0 {
		if err := db.Model(&models.Agent{}).
			Where("id IN ?", input.ShowPricesUpdate.AgentIDs).
			Update("show_prices", input.ShowPricesUpdate.ShowPrices); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update show_prices",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
			return
		}
	}

	// Reload with associations
	db.Preload("UserGroups").Preload("User").First(&permission, permission.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Permission updated successfully",
		"data":    permission,
	})
}

// ─────────────────────────────────────────────
// CheckPermission
// GET /permissions/check/:id
// Returns all source:action combos for an agent
// (direct | group | role — any one match counts)
// ─────────────────────────────────────────────
// func (h *Handler) CheckPermission(c *gin.Context) {
// 	db := middleware.GetDB(c)

// 	agentID := c.Param("id")
// 	if agentID == "" {
// 		c.JSON(http.StatusBadRequest, ErrorResponse{
// 			Error:   "Missing agent_id",
// 			Code:    "VALIDATION_ERROR",
// 			Details: "agent_id is required",
// 		})
// 		return
// 	}

// 	var agent models.Agent
// 	if err := db.First(&agent, "id = ?", agentID); err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			c.JSON(http.StatusNotFound, ErrorResponse{
// 				Error:   "Agent not found",
// 				Code:    "NOT_FOUND",
// 				Details: err.Error(),
// 			})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, ErrorResponse{
// 				Error:   "Database error",
// 				Code:    "DATABASE_ERROR",
// 				Details: err.Error(),
// 			})
// 		}
// 		return
// 	}

// 	permissionMap := make(map[string]bool)

// 	// 1️⃣ Direct — query Permissions from its own side via permission_agents.
// 	// Cannot use db.Model(&agent).Association("Permissions") because the
// 	// many2many tag is defined on Permissions.User, not on Agent.
// 	var directPermissions []models.Permissions
// 	if err := db.
// 		Joins("JOIN permission_agents ON permission_agents.permissions_id = permissions.id").
// 		Where("permission_agents.agent_id = ?", agent.ID).
// 		Find(&directPermissions); err == nil {
// 		for _, p := range directPermissions {
// 			permissionMap[p.Source+":"+p.Action] = true
// 		}
// 	}

// 	// 2️⃣ Group — find group IDs this agent belongs to via user_group_agents,
// 	// then query Permissions from its own side via permission_user_groups.
// 	// Cannot use db.Model(&group).Association("Permissions") for the same reason.
// 	var groups []models.UserGroups
// 	if err := db.
// 		Joins("JOIN user_group_agents ON user_group_agents.user_group_id = user_groups.id").
// 		Where("user_group_agents.agent_id = ?", agent.ID).
// 		Find(&groups); err == nil && len(groups) > 0 {

// 		groupIDs := make([]int, 0, len(groups))
// 		for _, g := range groups {
// 			groupIDs = append(groupIDs, g.ID)
// 		}

// 		var groupPermissions []models.Permissions
// 		if err := db.
// 			Joins("JOIN permission_user_groups ON permission_user_groups.permissions_id = permissions.id").
// 			Where("permission_user_groups.user_groups_id IN ?", groupIDs).
// 			Find(&groupPermissions); err == nil {
// 			for _, p := range groupPermissions {
// 				permissionMap[p.Source+":"+p.Action] = true
// 			}
// 		}
// 	}

// 	// 3️⃣ Role — agent.Role present in permissions.roles[]
// 	if agent.Role != "" {
// 		var rolePermissions []models.Permissions
// 		if err := db.Where("? = ANY(roles)", agent.Role).Find(&rolePermissions); err == nil {
// 			for _, p := range rolePermissions {
// 				permissionMap[p.Source+":"+p.Action] = true
// 			}
// 		}
// 	}

// 	permissionList := make([]string, 0, len(permissionMap))
// 	for k := range permissionMap {
// 		permissionList = append(permissionList, k)
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"agent_id":    agent.ID,
// 		"permissions": permissionList,
// 	})
// }

// ─────────────────────────────────────────────
// Response shapes for GetAllPermissions
// ─────────────────────────────────────────────

func (h *Handler) CheckPermission(c *gin.Context) {
	db := middleware.GetDB(c)

	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing agent_id",
			Code:    "VALIDATION_ERROR",
			Details: "agent_id is required",
		})
		return
	}

	var agent models.Agent
	if err := db.First(&agent, "id = ?", agentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Agent not found",
				Code:    "NOT_FOUND",
				Details: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Code:    "DATABASE_ERROR",
				Details: err.Error(),
			})
		}
		return
	}

	permissionMap := make(map[string]bool)

	var directPermissions []models.Permissions
	if err := db.
		Joins("JOIN permission_agents ON permission_agents.permissions_id = permissions.id").
		Where("permission_agents.agent_id = ?", agent.ID).
		Find(&directPermissions); err == nil {
		for _, p := range directPermissions {
			permissionMap[p.Source+":"+p.Action] = true
		}
	}

	// 2️⃣ Group permissions
	var groups []models.UserGroups
	db.
		Joins("JOIN user_group_agents ON user_group_agents.user_groups_id = user_groups.id"). // ← was user_group_id
		Where("user_group_agents.agent_id = ?", agent.ID).
		Find(&groups)
	if len(groups) > 0 {
		groupIDs := make([]uint, 0, len(groups))
		for _, g := range groups {
			groupIDs = append(groupIDs, uint(g.ID))
		}
		log.Printf("DEBUG groupIDs: %v", groupIDs)

		var groupPermissions []models.Permissions
		db.
			Joins("JOIN permission_user_groups ON permission_user_groups.permissions_id = permissions.id").
			Where("permission_user_groups.user_groups_id IN ?", groupIDs).
			Find(&groupPermissions)

		log.Printf("DEBUG groupPermissions found: %d → %+v", len(groupPermissions), groupPermissions)

		for _, p := range groupPermissions {
			permissionMap[p.Source+":"+p.Action] = true
		}
	}
	// if agent.Role != "" {
	// 	var rolePermissions []models.Permissions
	// 	if err := db.Where("? = ANY(roles)", agent.Role).Find(&rolePermissions); err == nil {
	// 		for _, p := range rolePermissions {
	// 			permissionMap[p.Source+":"+p.Action] = true
	// 		}
	// 	}
	// }

	permissionList := make([]string, 0, len(permissionMap))
	for k := range permissionMap {
		permissionList = append(permissionList, k)
	}

	c.JSON(http.StatusOK, gin.H{
		"agent_id":    agent.ID,
		"permissions": permissionList,
	})
}

type PermissionRow struct {
	Id      int         `json:"id,omitempty"`
	Source  string      `json:"source"`
	Action  string      `json:"action"`
	Entries interface{} `json:"entries"` // []string | []AgentInfo | []GroupInfo
}

type AgentInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GroupInfo struct {
	ID        int         `json:"id"`
	GroupName string      `json:"group_name"`
	Users     []AgentInfo `json:"users"`
}

// ─────────────────────────────────────────────
// GetAllPermissions
// GET /permissions/all
// ─────────────────────────────────────────────
func (h *Handler) GetAllPermissions(c *gin.Context) {
	db := middleware.GetDB(c)

	var permissions []models.Permissions
	if err := db.
		Preload("User").
		Preload("UserGroups").
		Preload("UserGroups.Users").
		Find(&permissions); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to fetch permissions",
			Code:    "DATABASE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// var roleRows []PermissionRow
	var userRows []PermissionRow
	var groupRows []PermissionRow

	for _, p := range permissions {
		// ── Roles — always emit this permission row ──
		// entries := []string(p.Roles)
		// if entries == nil {
		// 	entries = []string{}
		// }
		// roleRows = append(roleRows, PermissionRow{
		// 	Id:      p.ID,
		// 	Source:  p.Source,
		// 	Action:  p.Action,
		// 	Entries: entries,
		// })

		// ── Users — always emit ──
		agents := make([]AgentInfo, 0, len(p.User))
		for _, a := range p.User {
			agents = append(agents, AgentInfo{ID: a.ID, Name: a.Name, Email: a.Email})
		}
		userRows = append(userRows, PermissionRow{
			Id:      p.ID,
			Source:  p.Source,
			Action:  p.Action,
			Entries: agents,
		})

		// ── Groups — always emit ──
		groups := make([]GroupInfo, 0, len(p.UserGroups))
		for _, g := range p.UserGroups {
			members := make([]AgentInfo, 0, len(g.Users))
			for _, a := range g.Users {
				members = append(members, AgentInfo{ID: a.ID, Name: a.Name, Email: a.Email})
			}
			groups = append(groups, GroupInfo{ID: g.ID, GroupName: g.GroupName, Users: members})
		}
		groupRows = append(groupRows, PermissionRow{
			Id:      p.ID,
			Source:  p.Source,
			Action:  p.Action,
			Entries: groups,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		// "role_permissions":  roleRows,
		"user_permissions":  userRows,
		"group_permissions": groupRows,
	})
}

func (h *Handler) BulkUpdatePermissions(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var input models.BulkPermissionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": err.Error(),
		})
		return
	}

	tx := db.Begin()

	if err := tx.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Transaction initialization failed: " + err.Error(),
		})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "transaction panic recovered",
			})
		}
	}()

	for _, upd := range input.Updates {

		var permission models.Permissions

		if upd.ID != 0 {
			if err := tx.First(&permission, upd.ID); err != nil {
				tx.Rollback()
				c.JSON(http.StatusNotFound, gin.H{
					"error": fmt.Sprintf("permission with id %d not found", upd.ID),
				})
				return
			}
		} else {
			if err := tx.Where("source = ? AND action = ?", upd.Source, upd.Action).
				First(&permission); err != nil {
				tx.Rollback()
				c.JSON(http.StatusNotFound, gin.H{
					"error": fmt.Sprintf("permission not found for source=%s action=%s", upd.Source, upd.Action),
				})
				return
			}
		}

		var agents []models.Agent
		var userIDs []string

		for _, u := range upd.User {
			if u.ID != "" {
				userIDs = append(userIDs, u.ID)
			}
		}

		if len(userIDs) > 0 {
			if err := tx.Where("id IN ?", userIDs).Find(&agents); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
				return
			}
		}

		if err := tx.Model(&permission).Association("User").Replace(agents); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var groups []models.UserGroups
		var groupIDs []int

		for _, g := range upd.UserGroups {
			if g.ID != 0 {
				groupIDs = append(groupIDs, g.ID)
			}
		}

		if len(groupIDs) > 0 {
			if err := tx.Where("id IN ?", groupIDs).Find(&groups); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
				return
			}
		}

		if err := tx.Model(&permission).Association("UserGroups").Replace(groups); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if input.ShowPricesUpdate != nil && len(input.ShowPricesUpdate.AgentIDs) > 0 {
		if err := tx.Model(&models.Agent{}).
			Where("id IN ?", input.ShowPricesUpdate.AgentIDs).
			Update("show_prices", input.ShowPricesUpdate.ShowPrices); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update show_prices: " + err.Error(),
			})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk permissions updated successfully",
	})
}
