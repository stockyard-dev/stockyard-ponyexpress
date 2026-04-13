package main
import ("fmt";"log";"net/http";"os";"path/filepath";"github.com/stockyard-dev/stockyard-ponyexpress/internal/server";"github.com/stockyard-dev/stockyard-ponyexpress/internal/store";"github.com/stockyard-dev/stockyard/bus")
func main(){port:=os.Getenv("PORT");if port==""{port="9700"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./ponyexpress-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("ponyexpress: %v",err)};defer db.Close()
// Bus: one level up from the private data dir so every tool in a bundle
// shares one _bus.db. Non-fatal: ponyexpress still serves its REST API
// if bus is unreachable; just won't auto-draft from events.
var b *bus.Bus
if bb,berr:=bus.Open(filepath.Dir(dataDir),"ponyexpress");berr!=nil{log.Printf("ponyexpress: bus disabled: %v",berr)}else{b=bb;defer b.Close()}
srv:=server.New(db,server.DefaultLimits(),b)
fmt.Printf("\n  Pony Express — Self-hosted transactional email sender\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Questions? hello@stockyard.dev — I read every message\n\n",port,port)
log.Printf("ponyexpress: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
