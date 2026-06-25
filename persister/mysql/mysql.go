// Package mysql implements the align.Persister interface using GORM and a
// MySQL-compatible database. Use New (with a DSN) or NewFromDB (with an
// already-open *gorm.DB) to create a Persister, then register it with a
// Scheduler's Persister option.
package mysql

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethanbaker/align"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// sessionRow is the GORM model used to persist a single Session.
type sessionRow struct {
	gorm.Model
	Name       string     `gorm:"uniqueIndex;size:256"`
	ContactDay *time.Time `gorm:"index"`
}

// entryRow is the GORM model used to persist a single align.Entry, one row
// per entry rather than a single serialized blob, so entries can be
// queried/indexed by key and person. ContactorKey scopes entries to a single
// Contactor (e.g. "discord", "telegram"), set by whoever calls SaveEntries.
// It is named ContactorKey (column contactor_key) rather than Key/key, which
// is a reserved word in MySQL.
type entryRow struct {
	gorm.Model
	ContactorKey string `gorm:"index;size:128"`
	SessionID    uint   `gorm:"index"`
	Person       string `gorm:"size:256"`
	Index        int
	Fields       string `gorm:"type:json"` // JSON-encoded map[string]string
	Availability string `gorm:"type:json"` // JSON-encoded align.AvailabilityMap, empty when nil
}

// Persister implements align.Persister using a GORM database.
type Persister struct {
	db *gorm.DB
}

// New opens a GORM connection to the given DSN, migrates the sessions and
// entries tables, and returns a ready Persister.
func New(dsn string) (*Persister, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("mysql persister: open db: %w", err)
	}

	return NewFromDB(db)
}

// NewFromDB migrates the sessions and entries tables on an already-open
// *gorm.DB and returns a ready Persister. Use this when the caller manages
// the database connection itself.
func NewFromDB(db *gorm.DB) (*Persister, error) {
	if err := db.AutoMigrate(&sessionRow{}, &entryRow{}); err != nil {
		return nil, fmt.Errorf("mysql persister: migrate: %w", err)
	}

	return &Persister{db: db}, nil
}

// SaveSession upserts the session row for session.Name.
func (p *Persister) SaveSession(session *align.Session) error {
	row := sessionRow{
		Name:       session.Name,
		ContactDay: session.ContactDay,
	}

	if err := p.db.Where(sessionRow{Name: session.Name}).Assign(row).FirstOrCreate(&row).Error; err != nil {
		return fmt.Errorf("mysql persister: save session %q: %w", session.Name, err)
	}

	session.Model.ID = row.ID

	return nil
}

// LoadSession retrieves the session with the given name.
// It returns (nil, nil) when no session exists yet.
func (p *Persister) LoadSession(name string) (*align.Session, error) {
	var row sessionRow
	result := p.db.Where("name = ?", name).First(&row)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("mysql persister: load session %q: %w", name, result.Error)
	}

	return &align.Session{
		Model:      row.Model,
		Name:       row.Name,
		ContactDay: row.ContactDay,
	}, nil
}

// SaveEntries upserts entries under key. An entry whose ID matches an
// existing row is updated in place; an entry with no ID (or an ID that no
// longer matches any row) is inserted as a new row and has its generated ID
// written back onto the entry. Existing rows that no longer have a
// corresponding entry are left untouched rather than deleted, so historical
// entries are never lost.
func (p *Persister) SaveEntries(key string, session *align.Session, entries []align.Entry) error {
	var sessionID uint
	if session != nil {
		sessionID = session.ID
	}

	return p.db.Transaction(func(tx *gorm.DB) error {
		for i, e := range entries {
			fieldsJSON, err := json.Marshal(e.Fields)
			if err != nil {
				return fmt.Errorf("mysql persister: marshal entry fields: %w", err)
			}

			availabilityJSON, err := json.Marshal(e.Availability)
			if err != nil {
				return fmt.Errorf("mysql persister: marshal entry availability: %w", err)
			}

			row := entryRow{
				Model: gorm.Model{
					ID:        e.ID,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				ContactorKey: key,
				SessionID:    sessionID,
				Person:       e.Person,
				Index:        e.Index,
				Fields:       string(fieldsJSON),
				Availability: string(availabilityJSON),
			}

			// Save inserts when the primary key is zero (or unknown) and
			// updates every column in place when it matches an existing
			// row, so entries with the same ID are saved, not replaced.
			if err := tx.Save(&row).Error; err != nil {
				return fmt.Errorf("mysql persister: save entry: %w", err)
			}

			entries[i].ID = row.ID
			entries[i].SessionID = row.SessionID
		}

		return nil
	})
}

// LoadEntries retrieves the entries previously saved under key for session.
func (p *Persister) LoadEntries(key string, session *align.Session) ([]align.Entry, error) {
	var sessionID uint
	if session != nil {
		sessionID = session.ID
	}

	var rows []entryRow
	if err := p.db.Where("contactor_key = ? AND session_id = ?", key, sessionID).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("mysql persister: load entries for %q: %w", key, err)
	}

	entries := make([]align.Entry, 0, len(rows))
	for _, row := range rows {
		var fields map[string]string
		if err := json.Unmarshal([]byte(row.Fields), &fields); err != nil {
			return nil, fmt.Errorf("mysql persister: unmarshal entry fields for %q: %w", key, err)
		}

		var availability align.AvailabilityMap
		if row.Availability != "" {
			if err := json.Unmarshal([]byte(row.Availability), &availability); err != nil {
				return nil, fmt.Errorf("mysql persister: unmarshal entry availability for %q: %w", key, err)
			}
		}

		entries = append(entries, align.Entry{
			ID:           row.ID,
			SessionID:    row.SessionID,
			Person:       row.Person,
			Index:        row.Index,
			Fields:       fields,
			Availability: availability,
		})
	}

	return entries, nil
}

// VoidEntries deletes every entry previously saved under key for session.
func (p *Persister) VoidEntries(key string, session *align.Session) error {
	var sessionID uint
	if session != nil {
		sessionID = session.ID
	}

	if err := p.db.Where("contactor_key = ? AND session_id = ?", key, sessionID).Delete(&entryRow{}).Error; err != nil {
		return fmt.Errorf("mysql persister: void entries for %q: %w", key, err)
	}

	return nil
}
