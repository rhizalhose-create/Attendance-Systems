// repository/event_repository.go
package repository

import (
	"AttendanceManagementSystem/config"
	"AttendanceManagementSystem/models"
)

var db = config.DB

// CreateEvent creates a new event
func CreateEvent(event *models.Event) error {
	return db.Create(event).Error
}

// GetEventByID returns event by ID
func GetEventByID(id uint) (*models.Event, error) {
	var event models.Event
	err := db.First(&event, id).Error
	return &event, err
}

// GetAllEvents returns all events
func GetAllEvents() ([]models.Event, error) {
	var events []models.Event
	err := db.Find(&events).Error
	return events, err
}

// UpdateEvent updates an event
func UpdateEvent(event *models.Event) error {
	return db.Save(event).Error
}

// DeleteEvent deletes an event
func DeleteEvent(id uint) error {
	return db.Delete(&models.Event{}, id).Error
}

// GetEventsByUser returns events for a specific user
func GetEventsByUser(userID string) ([]models.Event, error) {
	var events []models.Event
	// This is a basic implementation - adjust based on your business logic
	err := db.Where("created_by = ? OR college IN (SELECT college FROM users WHERE user_id = ?)", 
		userID, userID).Find(&events).Error
	return events, err
}