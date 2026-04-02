package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Delivery struct {
	ID string `json:"id"`
	Recipient string `json:"recipient"`
	Channel string `json:"channel"`
	Subject string `json:"subject"`
	Body string `json:"body"`
	Status string `json:"status"`
	Attempts int `json:"attempts"`
	DeliveredAt string `json:"delivered_at"`
	ErrorMsg string `json:"error_msg"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"ponyexpress.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS deliveries(id TEXT PRIMARY KEY,recipient TEXT NOT NULL,channel TEXT DEFAULT 'email',subject TEXT DEFAULT '',body TEXT DEFAULT '',status TEXT DEFAULT 'pending',attempts INTEGER DEFAULT 0,delivered_at TEXT DEFAULT '',error_msg TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Delivery)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO deliveries(id,recipient,channel,subject,body,status,attempts,delivered_at,error_msg,created_at)VALUES(?,?,?,?,?,?,?,?,?,?)`,e.ID,e.Recipient,e.Channel,e.Subject,e.Body,e.Status,e.Attempts,e.DeliveredAt,e.ErrorMsg,e.CreatedAt);return err}
func(d *DB)Get(id string)*Delivery{var e Delivery;if d.db.QueryRow(`SELECT id,recipient,channel,subject,body,status,attempts,delivered_at,error_msg,created_at FROM deliveries WHERE id=?`,id).Scan(&e.ID,&e.Recipient,&e.Channel,&e.Subject,&e.Body,&e.Status,&e.Attempts,&e.DeliveredAt,&e.ErrorMsg,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Delivery{rows,_:=d.db.Query(`SELECT id,recipient,channel,subject,body,status,attempts,delivered_at,error_msg,created_at FROM deliveries ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Delivery;for rows.Next(){var e Delivery;rows.Scan(&e.ID,&e.Recipient,&e.Channel,&e.Subject,&e.Body,&e.Status,&e.Attempts,&e.DeliveredAt,&e.ErrorMsg,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Update(e *Delivery)error{_,err:=d.db.Exec(`UPDATE deliveries SET recipient=?,channel=?,subject=?,body=?,status=?,attempts=?,delivered_at=?,error_msg=? WHERE id=?`,e.Recipient,e.Channel,e.Subject,e.Body,e.Status,e.Attempts,e.DeliveredAt,e.ErrorMsg,e.ID);return err}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM deliveries WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM deliveries`).Scan(&n);return n}

func(d *DB)Search(q string, filters map[string]string)[]Delivery{
    where:="1=1"
    args:=[]any{}
    if q!=""{
        where+=" AND (recipient LIKE ? OR subject LIKE ? OR body LIKE ?)"
        args=append(args,"%"+q+"%");args=append(args,"%"+q+"%");args=append(args,"%"+q+"%");
    }
    if v,ok:=filters["channel"];ok&&v!=""{where+=" AND channel=?";args=append(args,v)}
    if v,ok:=filters["status"];ok&&v!=""{where+=" AND status=?";args=append(args,v)}
    rows,_:=d.db.Query(`SELECT id,recipient,channel,subject,body,status,attempts,delivered_at,error_msg,created_at FROM deliveries WHERE `+where+` ORDER BY created_at DESC`,args...)
    if rows==nil{return nil};defer rows.Close()
    var o []Delivery;for rows.Next(){var e Delivery;rows.Scan(&e.ID,&e.Recipient,&e.Channel,&e.Subject,&e.Body,&e.Status,&e.Attempts,&e.DeliveredAt,&e.ErrorMsg,&e.CreatedAt);o=append(o,e)};return o
}

func(d *DB)Stats()map[string]any{
    m:=map[string]any{"total":d.Count()}
    rows,_:=d.db.Query(`SELECT status,COUNT(*) FROM deliveries GROUP BY status`)
    if rows!=nil{defer rows.Close();by:=map[string]int{};for rows.Next(){var s string;var c int;rows.Scan(&s,&c);by[s]=c};m["by_status"]=by}
    return m
}
