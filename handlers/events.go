	package handlers

	import (
		"log"
		"time"
		"AttendanceManagementSystem/config"
		"AttendanceManagementSystem/models"
		"AttendanceManagementSystem/utils"

		"github.com/gofiber/fiber/v2"
	)

	// CreateEvent - Create new event (Admin/SuperAdmin only)
	func CreateEvent(c *fiber.Ctx) error {
		var req models.CreateEventRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
		}

		// Check if user is admin or superadmin
		userRole := c.Get(utils.HeaderUserRole)
		if userRole != "admin" && userRole != "superadmin" {
			return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
		}

		// Validate required fields
		if req.EventName == "" || req.EventType == "" || req.Description == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Event name, type, and description are required"})
		}

		// Validate time
		if req.StartTime.After(req.EndTime) {
			return c.Status(400).JSON(fiber.Map{"error": "Start time cannot be after end time"})
		}

		if req.StartTime.Before(time.Now()) {
			return c.Status(400).JSON(fiber.Map{"error": "Start time cannot be in the past"})
		}

		// Get user ID from context
		userID := c.Get(utils.HeaderUserID)

		event := models.Event{
			EventName:   req.EventName,
			EventType:   req.EventType,
			Description: req.Description,
			Location:    req.Location,
			Course:      req.Course,
			Department:  req.Department,
			College:     req.College,
			CreatedBy:   userID,
			IsActive:    true,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := config.DB.Create(&event).Error; err != nil {
			log.Printf("❌ Failed to create event: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create event"})
		}

		log.Printf("✅ Event created successfully: %s by %s", event.EventName, userID)

		return c.JSON(fiber.Map{
			"message": "Event created successfully",
			"event": fiber.Map{
				"id":          event.ID,
				"event_name":  event.EventName,
				"event_type":  event.EventType,
				"description": event.Description,
				"location":    event.Location,
				"course":      event.Course,
				"department":  event.Department,
				"college":     event.College,
				"created_by":  event.CreatedBy,
				"start_time":  event.StartTime,
				"end_time":    event.EndTime,
				"is_active":   event.IsActive,
			},
		})
	}

	// GetEvents - Get all active events
	func GetEvents(c *fiber.Ctx) error {
		var events []models.Event
		
		// Get query parameters for filtering
		eventType := c.Query("type")
		course := c.Query("course")
		department := c.Query("department")
		college := c.Query("college")

		query := config.DB.Where("is_active = ? AND end_time > ?", true, time.Now())

		// Apply filters if provided
		if eventType != "" {
			query = query.Where("event_type = ?", eventType)
		}
		if course != "" {
			query = query.Where("course = ?", course)
		}
		if department != "" {
			query = query.Where("department = ?", department)
		}
		if college != "" {
			query = query.Where("college = ?", college)
		}

		if err := query.Order("start_time ASC").Find(&events).Error; err != nil {
			log.Printf("❌ Failed to fetch events: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch events"})
		}

		return c.JSON(fiber.Map{
			"events": events,
			"count":  len(events),
		})
	}

	// GetEventByID - Get specific event by ID
	func GetEventByID(c *fiber.Ctx) error {
		eventID := c.Params("id")

		var event models.Event
		if err := config.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Event not found"})
		}

		return c.JSON(fiber.Map{
			"event": event,
		})
	}

	// UpdateEvent - Update event (Admin/SuperAdmin only)
	func UpdateEvent(c *fiber.Ctx) error {
		eventID := c.Params("id")

		var req models.CreateEventRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
		}

		// Check if user is admin or superadmin
		userRole := c.Get(utils.HeaderUserRole)
		if userRole != "admin" && userRole != "superadmin" {
			return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
		}

		// Find event
		var event models.Event
		if err := config.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Event not found"})
		}

		// Validate time if provided
		if !req.StartTime.IsZero() && !req.EndTime.IsZero() {
			if req.StartTime.After(req.EndTime) {
				return c.Status(400).JSON(fiber.Map{"error": "Start time cannot be after end time"})
			}
		}

		// Update event
		updates := map[string]interface{}{
			"event_name":  req.EventName,
			"event_type":  req.EventType,
			"description": req.Description,
			"location":    req.Location,
			"course":      req.Course,
			"department":  req.Department,
			"college":     req.College,
			"start_time":  req.StartTime,
			"end_time":    req.EndTime,
			"updated_at":  time.Now(),
		}

		if err := config.DB.Model(&event).Updates(updates).Error; err != nil {
			log.Printf(" Failed to update event: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update event"})
		}

		log.Printf(" Event updated successfully: %s", event.EventName)
		return c.JSON(fiber.Map{
			"message": "Event updated successfully",
			"event":   event,
		})
	}

	// DeleteEvent - Soft delete event (Admin/SuperAdmin only)
	func DeleteEvent(c *fiber.Ctx) error {
		eventID := c.Params("id")

		// Check if user is admin or superadmin
		userRole := c.Get(utils.HeaderUserRole)
		if userRole != "admin" && userRole != "superadmin" {
			return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
		}

		// Find event
		var event models.Event
		if err := config.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Event not found"})
		}

		// Soft delete by setting is_active to false
		if err := config.DB.Model(&event).Update("is_active", false).Error; err != nil {
			log.Printf(" Failed to delete event: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to delete event"})
		}

		log.Printf(" Event deleted successfully: %s", event.EventName)

		return c.JSON(fiber.Map{
			"message": "Event deleted successfully",
		})
	}

	// GetMyEvents - Get events created by current user
	func GetMyEvents(c *fiber.Ctx) error {
		userID := c.Get(utils.HeaderUserID)
		userRole := c.Get(utils.HeaderUserRole)

		var events []models.Event
		query := config.DB.Where("is_active = ?", true)

		// If user is student, show all active events
		// If user is admin/superadmin, show events they created
		if userRole == "admin" || userRole == "superadmin" {
			query = query.Where("created_by = ?", userID)
		}

		if err := query.Order("start_time ASC").Find(&events).Error; err != nil {
			log.Printf(" Failed to fetch events: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch events"})
		}

		return c.JSON(fiber.Map{
			"events": events,
			"count":  len(events),
		})
	}