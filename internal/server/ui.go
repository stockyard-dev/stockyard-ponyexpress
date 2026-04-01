package server

import "net/http"

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Pony Express — Stockyard</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
<style>*{margin:0;padding:0;box-sizing:border-box}body{background:#1a1410;color:#f0e6d3;font-family:'JetBrains Mono',monospace;padding:2rem}
.hdr{font-size:.7rem;color:#a0845c;letter-spacing:3px;text-transform:uppercase;margin-bottom:2rem;border-bottom:2px solid #8b3d1a;padding-bottom:.8rem}
.cards{display:grid;grid-template-columns:repeat(4,1fr);gap:1rem;margin-bottom:2rem}.card{background:#241e18;border:1px solid #2e261e;padding:1rem}.card-val{font-size:1.6rem;font-weight:700;display:block}.card-lbl{font-size:.55rem;letter-spacing:2px;text-transform:uppercase;color:#a0845c;margin-top:.2rem}
.section{margin-bottom:2rem}.section h2{font-size:.65rem;letter-spacing:3px;text-transform:uppercase;color:#e8753a;margin-bottom:.8rem}
table{width:100%;border-collapse:collapse;font-size:.72rem}th{background:#2e261e;padding:.4rem .6rem;text-align:left;color:#c4a87a;font-size:.6rem;letter-spacing:1px;text-transform:uppercase}td{padding:.4rem .6rem;border-bottom:1px solid #2e261e;color:#bfb5a3}.empty{color:#7a7060;text-align:center;padding:2rem;font-style:italic}
.pill{font-size:.55rem;padding:.1rem .4rem;text-transform:uppercase}.pill-sent{background:#1a3a2a;color:#5ba86e}.pill-queued{background:#2a2a1a;color:#d4a843}.pill-failed{background:#2a1a1a;color:#c0392b}
</style></head><body>
<div class="hdr">Stockyard · Pony Express</div>
<div class="cards"><div class="card"><span class="card-val" id="s-queued">—</span><span class="card-lbl">Queued</span></div><div class="card"><span class="card-val" id="s-sent">—</span><span class="card-lbl">Sent</span></div><div class="card"><span class="card-val" id="s-failed">—</span><span class="card-lbl">Failed</span></div><div class="card"><span class="card-val" id="s-tpl">—</span><span class="card-lbl">Templates</span></div></div>
<div class="section"><h2>Recent Emails</h2><table><thead><tr><th>To</th><th>Subject</th><th>Status</th><th>Time</th></tr></thead><tbody id="emails-body"></tbody></table></div>
<script>
async function refresh(){
  try{const s=await(await fetch('/api/status')).json();document.getElementById('s-queued').textContent=s.queued||0;document.getElementById('s-sent').textContent=s.sent||0;document.getElementById('s-failed').textContent=s.failed||0;document.getElementById('s-tpl').textContent=s.templates||0;}catch(e){}
  try{const d=await(await fetch('/api/emails')).json();const es=d.emails||[];const tb=document.getElementById('emails-body');
  tb.innerHTML=es.length?es.map(e=>'<tr><td>'+esc(e.to)+'</td><td>'+esc(e.subject)+'</td><td><span class="pill pill-'+e.status+'">'+e.status+'</span></td><td style="font-size:.62rem;color:#7a7060">'+esc(e.created_at.slice(0,19))+'</td></tr>').join(''):'<tr><td colspan="4" class="empty">No emails yet</td></tr>';}catch(e){}
}
function esc(s){const d=document.createElement('div');d.textContent=s||'';return d.innerHTML;}
refresh();setInterval(refresh,5000);
</script></body></html>`))
}
