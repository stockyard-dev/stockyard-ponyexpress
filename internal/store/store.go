package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "modernc.org/sqlite"
)

type DB struct{ conn *sql.DB }

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil { return nil, fmt.Errorf("mkdir: %w", err) }
	conn, err := sql.Open("sqlite", filepath.Join(dataDir, "ponyexpress.db"))
	if err != nil { return nil, err }
	conn.Exec("PRAGMA journal_mode=WAL"); conn.Exec("PRAGMA busy_timeout=5000"); conn.SetMaxOpenConns(4)
	db := &DB{conn: conn}; return db, db.migrate()
}

func (db *DB) Close() error { return db.conn.Close() }

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS emails (
    id TEXT PRIMARY KEY, to_addr TEXT NOT NULL, from_addr TEXT DEFAULT '',
    subject TEXT DEFAULT '', body_html TEXT DEFAULT '', body_text TEXT DEFAULT '',
    status TEXT DEFAULT 'queued', error TEXT DEFAULT '',
    sent_at TEXT DEFAULT '', created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_emails_status ON emails(status);
CREATE TABLE IF NOT EXISTS templates (
    id TEXT PRIMARY KEY, name TEXT NOT NULL UNIQUE, subject TEXT DEFAULT '',
    body_html TEXT DEFAULT '', body_text TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);`)
	return err
}

type Email struct {
	ID string `json:"id"`; To string `json:"to"`; From string `json:"from"`
	Subject string `json:"subject"`; BodyHTML string `json:"body_html,omitempty"`; BodyText string `json:"body_text,omitempty"`
	Status string `json:"status"`; Error string `json:"error,omitempty"`
	SentAt string `json:"sent_at,omitempty"`; CreatedAt string `json:"created_at"`
}

func (db *DB) QueueEmail(to, from, subject, bodyHTML, bodyText string) (*Email, error) {
	id := "eml_" + genID(8); now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO emails (id,to_addr,from_addr,subject,body_html,body_text,created_at) VALUES (?,?,?,?,?,?,?)",
		id, to, from, subject, bodyHTML, bodyText, now)
	if err != nil { return nil, err }
	return &Email{ID: id, To: to, From: from, Subject: subject, BodyHTML: bodyHTML, BodyText: bodyText, Status: "queued", CreatedAt: now}, nil
}

func (db *DB) ListEmails(limit int) ([]Email, error) {
	if limit <= 0 { limit = 50 }
	rows, err := db.conn.Query("SELECT id,to_addr,from_addr,subject,status,error,sent_at,created_at FROM emails ORDER BY created_at DESC LIMIT ?", limit)
	if err != nil { return nil, err }; defer rows.Close()
	var out []Email
	for rows.Next() { var e Email; rows.Scan(&e.ID, &e.To, &e.From, &e.Subject, &e.Status, &e.Error, &e.SentAt, &e.CreatedAt); out = append(out, e) }
	return out, rows.Err()
}

func (db *DB) GetEmail(id string) (*Email, error) {
	var e Email
	err := db.conn.QueryRow("SELECT id,to_addr,from_addr,subject,body_html,body_text,status,error,sent_at,created_at FROM emails WHERE id=?", id).
		Scan(&e.ID, &e.To, &e.From, &e.Subject, &e.BodyHTML, &e.BodyText, &e.Status, &e.Error, &e.SentAt, &e.CreatedAt)
	return &e, err
}

func (db *DB) MarkSent(id string) { now := time.Now().UTC().Format(time.RFC3339); db.conn.Exec("UPDATE emails SET status='sent',sent_at=? WHERE id=?", now, id) }
func (db *DB) MarkFailed(id, errMsg string) { db.conn.Exec("UPDATE emails SET status='failed',error=? WHERE id=?", errMsg, id) }

func (db *DB) PendingEmails() ([]Email, error) {
	rows, err := db.conn.Query("SELECT id,to_addr,from_addr,subject,body_html,body_text,status,error,sent_at,created_at FROM emails WHERE status='queued' LIMIT 50")
	if err != nil { return nil, err }; defer rows.Close()
	var out []Email
	for rows.Next() { var e Email; rows.Scan(&e.ID, &e.To, &e.From, &e.Subject, &e.BodyHTML, &e.BodyText, &e.Status, &e.Error, &e.SentAt, &e.CreatedAt); out = append(out, e) }
	return out, rows.Err()
}

type Template struct {
	ID string `json:"id"`; Name string `json:"name"`; Subject string `json:"subject"`
	BodyHTML string `json:"body_html"`; BodyText string `json:"body_text"`; CreatedAt string `json:"created_at"`
}

func (db *DB) CreateTemplate(name, subject, bodyHTML, bodyText string) (*Template, error) {
	id := "tpl_" + genID(6); now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO templates (id,name,subject,body_html,body_text,created_at) VALUES (?,?,?,?,?,?)", id, name, subject, bodyHTML, bodyText, now)
	if err != nil { return nil, err }
	return &Template{ID: id, Name: name, Subject: subject, BodyHTML: bodyHTML, BodyText: bodyText, CreatedAt: now}, nil
}

func (db *DB) ListTemplates() ([]Template, error) {
	rows, err := db.conn.Query("SELECT id,name,subject,body_html,body_text,created_at FROM templates ORDER BY name")
	if err != nil { return nil, err }; defer rows.Close()
	var out []Template
	for rows.Next() { var t Template; rows.Scan(&t.ID, &t.Name, &t.Subject, &t.BodyHTML, &t.BodyText, &t.CreatedAt); out = append(out, t) }
	return out, rows.Err()
}

func (db *DB) GetTemplateByName(name string) (*Template, error) {
	var t Template
	err := db.conn.QueryRow("SELECT id,name,subject,body_html,body_text,created_at FROM templates WHERE name=?", name).
		Scan(&t.ID, &t.Name, &t.Subject, &t.BodyHTML, &t.BodyText, &t.CreatedAt)
	return &t, err
}

func (db *DB) DeleteTemplate(id string) { db.conn.Exec("DELETE FROM templates WHERE id=?", id) }

func (db *DB) Stats() map[string]any {
	var queued, sent, failed, templates int
	db.conn.QueryRow("SELECT COUNT(*) FROM emails WHERE status='queued'").Scan(&queued)
	db.conn.QueryRow("SELECT COUNT(*) FROM emails WHERE status='sent'").Scan(&sent)
	db.conn.QueryRow("SELECT COUNT(*) FROM emails WHERE status='failed'").Scan(&failed)
	db.conn.QueryRow("SELECT COUNT(*) FROM templates").Scan(&templates)
	return map[string]any{"queued": queued, "sent": sent, "failed": failed, "templates": templates}
}

func genID(n int) string { b := make([]byte, n); rand.Read(b); return hex.EncodeToString(b) }
