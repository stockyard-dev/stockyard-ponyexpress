package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
	"github.com/stockyard-dev/stockyard-ponyexpress/internal/store"
)

type Server struct { db *store.DB; mux *http.ServeMux; port int; limits Limits; smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom string }

func New(db *store.DB, port int, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), port: port, limits: limits,
		smtpHost: os.Getenv("SMTP_HOST"), smtpPort: os.Getenv("SMTP_PORT"),
		smtpUser: os.Getenv("SMTP_USER"), smtpPass: os.Getenv("SMTP_PASS"),
		smtpFrom: os.Getenv("SMTP_FROM")}
	if s.smtpPort == "" { s.smtpPort = "587" }
	s.mux.HandleFunc("POST /api/send", s.hSend)
	s.mux.HandleFunc("GET /api/emails", s.hListEmails)
	s.mux.HandleFunc("GET /api/emails/{id}", s.hGetEmail)
	s.mux.HandleFunc("POST /api/templates", s.hCreateTemplate)
	s.mux.HandleFunc("GET /api/templates", s.hListTemplates)
	s.mux.HandleFunc("DELETE /api/templates/{id}", s.hDelTemplate)
	s.mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) { wj(w, 200, s.db.Stats()) })
	s.mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { wj(w, 200, map[string]string{"status": "ok"}) })
	s.mux.HandleFunc("GET /ui", s.handleUI)
	s.mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) { wj(w, 200, map[string]any{"product": "stockyard-ponyexpress", "version": "0.1.0"}) })
	return s
}

func (s *Server) Start() error {
	go s.processQueue()
	log.Printf("[ponyexpress] :%d", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.mux)
}

func (s *Server) processQueue() {
	for {
		time.Sleep(2 * time.Second)
		emails, _ := s.db.PendingEmails()
		for _, e := range emails {
			if s.smtpHost == "" { s.db.MarkFailed(e.ID, "SMTP not configured"); continue }
			body := e.BodyHTML
			if body == "" { body = e.BodyText }
			from := e.From; if from == "" { from = s.smtpFrom }
			msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", from, e.To, e.Subject, body)
			auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)
			err := smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, from, []string{e.To}, []byte(msg))
			if err != nil { s.db.MarkFailed(e.ID, err.Error()) } else { s.db.MarkSent(e.ID) }
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *Server) hSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		To string `json:"to"`; Subject string `json:"subject"`; BodyHTML string `json:"body_html"`
		BodyText string `json:"body_text"`; From string `json:"from"`; Template string `json:"template"`
		Vars map[string]string `json:"vars"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.To == "" { wj(w, 400, map[string]string{"error": "to required"}); return }
	if req.Template != "" {
		tpl, err := s.db.GetTemplateByName(req.Template)
		if err != nil { wj(w, 404, map[string]string{"error": "template not found"}); return }
		req.Subject = tpl.Subject; req.BodyHTML = tpl.BodyHTML; req.BodyText = tpl.BodyText
		for k, v := range req.Vars { req.BodyHTML = strings.ReplaceAll(req.BodyHTML, "{{"+k+"}}", v); req.Subject = strings.ReplaceAll(req.Subject, "{{"+k+"}}", v) }
	}
	e, err := s.db.QueueEmail(req.To, req.From, req.Subject, req.BodyHTML, req.BodyText)
	if err != nil { wj(w, 500, map[string]string{"error": err.Error()}); return }
	wj(w, 201, map[string]any{"email": e})
}

func (s *Server) hListEmails(w http.ResponseWriter, r *http.Request) {
	es, _ := s.db.ListEmails(50); if es == nil { es = []store.Email{} }
	wj(w, 200, map[string]any{"emails": es, "count": len(es)})
}
func (s *Server) hGetEmail(w http.ResponseWriter, r *http.Request) {
	e, err := s.db.GetEmail(r.PathValue("id")); if err != nil { wj(w, 404, map[string]string{"error": "not found"}); return }
	wj(w, 200, map[string]any{"email": e})
}
func (s *Server) hCreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct { Name string `json:"name"`; Subject string `json:"subject"`; BodyHTML string `json:"body_html"`; BodyText string `json:"body_text"` }
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.Name == "" { wj(w, 400, map[string]string{"error": "name required"}); return }
	t, err := s.db.CreateTemplate(req.Name, req.Subject, req.BodyHTML, req.BodyText)
	if err != nil { wj(w, 500, map[string]string{"error": err.Error()}); return }
	wj(w, 201, map[string]any{"template": t})
}
func (s *Server) hListTemplates(w http.ResponseWriter, r *http.Request) {
	ts, _ := s.db.ListTemplates(); if ts == nil { ts = []store.Template{} }
	wj(w, 200, map[string]any{"templates": ts, "count": len(ts)})
}
func (s *Server) hDelTemplate(w http.ResponseWriter, r *http.Request) { s.db.DeleteTemplate(r.PathValue("id")); wj(w, 200, map[string]string{"status": "deleted"}) }

func wj(w http.ResponseWriter, code int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(code); json.NewEncoder(w).Encode(v) }
