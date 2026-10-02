package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
)

type ActivityEvent struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	EntityType string         `json:"entity_type" gorm:"index"`
	EntityID   string         `json:"entity_id" gorm:"index"`
	Action     string         `json:"action" gorm:"index"`
	UserID     string         `json:"user_id,omitempty" gorm:"index"`
	UserName   string         `json:"user_name,omitempty"`
	Payload    map[string]any `json:"payload,omitempty" gorm:"type:jsonb"`
	CreatedAt  time.Time      `json:"created_at,omitempty" gorm:"index"`
}

func InsertActivityEvent(ctx context.Context, db *gorm.DB, event ActivityEvent) error {
	var err error
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	payloadBytes := []byte("null")
	if event.Payload != nil {
		payloadBytes, err = json.Marshal(event.Payload)
		if err != nil {
			return fmt.Errorf("marshal activity payload: %w", err)
		}
	}

	log.Printf("inserting activity event: entity=%s id=%s action=%s payload=%s",
		event.EntityType,
		event.EntityID,
		event.Action,
		string(payloadBytes),
	)

	var insertedID uint
	var insertedCreatedAt time.Time
	row := db.Raw(
		`INSERT INTO public.activity_events (entity_type, entity_id, "action", user_id, user_name, payload, created_at) VALUES (?, ?, ?, ?, ?, ?::jsonb, ?) RETURNING id, created_at`,
		event.EntityType,
		event.EntityID,
		event.Action,
		nullableString(event.UserID),
		nullableString(event.UserName),
		string(payloadBytes),
		event.CreatedAt,
	).Row()
	if err := row.Scan(&insertedID, &insertedCreatedAt); err != nil {
		log.Printf("activity event insert failed: %v", err)
		return err
	}
	event.ID = insertedID
	event.CreatedAt = insertedCreatedAt

	log.Printf("activity event inserted successfully: entity=%s id=%d action=%s", event.EntityType, event.ID, event.Action)

	notifyChannel := os.Getenv("ACTIVITY_LISTEN_CHANNEL")
	if notifyChannel == "" {
		notifyChannel = "activity_events_channel"
	}
	activityJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal activity event for notify: %v", err)
		return err
	}
	if err := db.Exec("SELECT pg_notify(?, ?)", notifyChannel, string(activityJSON)).Error; err != nil {
		log.Printf("activity notify failed: %v", err)
		return err
	}

	log.Printf("activity event notify sent on channel=%s", notifyChannel)
	return nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
