package server
import ("context";"encoding/json";"fmt";"log";"net/http";"github.com/stockyard-dev/stockyard-ponyexpress/internal/store";"github.com/stockyard-dev/stockyard/bus")
type Server struct{db *store.DB;mux *http.ServeMux;limits Limits;bus *bus.Bus}
func New(db *store.DB,limits Limits,b *bus.Bus)*Server{s:=&Server{db:db,mux:http.NewServeMux(),limits:limits,bus:b}
s.mux.HandleFunc("GET /api/deliveries",s.list)
s.mux.HandleFunc("POST /api/deliveries",s.create)
s.mux.HandleFunc("GET /api/deliveries/{id}",s.get)
s.mux.HandleFunc("PUT /api/deliveries/{id}",s.update)
s.mux.HandleFunc("DELETE /api/deliveries/{id}",s.del)
s.mux.HandleFunc("GET /api/stats",s.stats)
s.mux.HandleFunc("GET /api/health",s.health)
s.mux.HandleFunc("GET /ui",s.dashboard);s.mux.HandleFunc("GET /ui/",s.dashboard);s.mux.HandleFunc("GET /",s.root);
s.mux.HandleFunc("GET /api/tier",func(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"tier":s.limits.Tier,"upgrade_url":"https://stockyard.dev/ponyexpress/"})})
s.subscribeBus()
return s}
func(s *Server)ServeHTTP(w http.ResponseWriter,r *http.Request){s.mux.ServeHTTP(w,r)}
func wj(w http.ResponseWriter,c int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(c);json.NewEncoder(w).Encode(v)}
func we(w http.ResponseWriter,c int,m string){wj(w,c,map[string]string{"error":m})}
func(s *Server)root(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};http.Redirect(w,r,"/ui",302)}
func(s *Server)list(w http.ResponseWriter,r *http.Request){
    q:=r.URL.Query().Get("q")
    filters:=map[string]string{}
    if v:=r.URL.Query().Get("channel");v!=""{filters["channel"]=v}
    if v:=r.URL.Query().Get("status");v!=""{filters["status"]=v}
    if q!=""||len(filters)>0{wj(w,200,map[string]any{"deliveries":oe(s.db.Search(q,filters))});return}
    wj(w,200,map[string]any{"deliveries":oe(s.db.List())})
}
func(s *Server)create(w http.ResponseWriter,r *http.Request){if s.limits.MaxItems>0{items:=s.db.List();if len(items)>=s.limits.MaxItems{we(w,402,"Free tier limit reached. Upgrade at https://stockyard.dev/ponyexpress/");return}};var e store.Delivery;json.NewDecoder(r.Body).Decode(&e);if e.Recipient==""{we(w,400,"recipient required");return};s.db.Create(&e);wj(w,201,s.db.Get(e.ID))}
func(s *Server)get(w http.ResponseWriter,r *http.Request){e:=s.db.Get(r.PathValue("id"));if e==nil{we(w,404,"not found");return};wj(w,200,e)}
func(s *Server)update(w http.ResponseWriter,r *http.Request){
    existing:=s.db.Get(r.PathValue("id"));if existing==nil{we(w,404,"not found");return}
    var patch store.Delivery;json.NewDecoder(r.Body).Decode(&patch);patch.ID=existing.ID;patch.CreatedAt=existing.CreatedAt
    if patch.Recipient==""{patch.Recipient=existing.Recipient}
    s.db.Update(&patch);wj(w,200,s.db.Get(patch.ID))
}
func(s *Server)del(w http.ResponseWriter,r *http.Request){s.db.Delete(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)stats(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.Stats())}
func(s *Server)health(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"status":"ok","service":"ponyexpress","deliveries":s.db.Count()})}
func oe[T any](s []T)[]T{if s==nil{return[]T{}};return s}
func init(){log.SetFlags(log.LstdFlags|log.Lshortfile)}

// subscribeBus wires cross-tool events to auto-drafted email
// deliveries. No-op when s.bus is nil (standalone mode).
//
// This is STARTER WIRING — it subscribes to a deliberate allowlist
// rather than SubscribeAll so unexpected new topics don't silently
// spam customer inboxes. The subject-line and body-line logic here
// is placeholder; real templated email is a future feature that
// reads personalization config (same pattern as dossier/billfold).
//
// Handlers are idempotent by payload ID — the bus may redeliver on
// process restart, but s.db.Create generates a fresh Delivery each
// time regardless, which matches the at-least-once reality. Dedup
// by (topic, payload_id) can arrive as a refinement.
//
// All deliveries created here land with status="pending" and a
// channel derived from the topic context (email for now). A future
// dispatcher drains them; this wiring just queues.
func (s *Server) subscribeBus() {
	if s.bus == nil {
		return
	}
	// Topics that map to a customer-facing outbound email. Expanding
	// this list is a PR-reviewed change, not a config flag — every
	// addition is "a new email customers will now receive."
	topics := []string{
		"contacts.created",          // welcome / intake confirmation
		"appointment.booked",        // booking confirmation
		"appointment.cancelled",     // cancellation notice
		"invoice.sent",              // invoice notification
		"invoice.overdue",           // overdue reminder
		"quote.sent",                // quote delivery
		"waiver.signed",             // signed-waiver receipt
	}
	for _, topic := range topics {
		t := topic // capture for closure
		s.bus.Subscribe(t, func(_ context.Context, e bus.Event) error {
			return s.draftDelivery(t, e)
		})
	}
	log.Printf("ponyexpress: subscribed to %d bus topics", len(topics))
}

// draftDelivery converts a bus event into a pending Delivery row.
// Best-effort: a decode failure logs and drops the event — a broken
// handler must not block the bus dispatch loop.
func (s *Server) draftDelivery(topic string, e bus.Event) error {
	var payload map[string]any
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		log.Printf("ponyexpress: decode %s payload: %v", topic, err)
		return nil // nil because we don't want the bus retrying forever
	}
	recipient := firstNonEmpty(
		stringField(payload, "client_email"),
		stringField(payload, "email"),
		stringField(payload, "recipient"),
	)
	if recipient == "" {
		// No email address on this payload — the event is real but
		// there's nothing to deliver to. Log and drop.
		log.Printf("ponyexpress: %s event has no email, skipping", topic)
		return nil
	}
	subject := fmt.Sprintf("[%s] %s", e.Source, topic)
	body := string(e.Payload) // raw JSON body placeholder; templating is future work
	d := store.Delivery{
		Recipient: recipient,
		Channel:   "email",
		Subject:   subject,
		Body:      body,
		Status:    "pending",
	}
	if err := s.db.Create(&d); err != nil {
		log.Printf("ponyexpress: create delivery from %s: %v", topic, err)
	}
	return nil
}

func stringField(m map[string]any, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
