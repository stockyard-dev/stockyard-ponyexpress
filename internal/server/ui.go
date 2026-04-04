package server
import "net/http"
func(s *Server)dashboard(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","text/html");w.Write([]byte(dashHTML))}
const dashHTML=`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Pony Express</title>
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--orange:#d4843a;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-size:.9rem;letter-spacing:2px}
.main{padding:1.5rem;max-width:900px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(4,1fr);gap:.5rem;margin-bottom:1.2rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.6rem;text-align:center}.st-v{font-size:1.1rem}.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.1rem}
.msg{background:var(--bg2);border:1px solid var(--bg3);padding:.7rem 1rem;margin-bottom:.4rem;font-size:.75rem}
.msg-top{display:flex;justify-content:space-between;align-items:center}
.msg-to{color:var(--cream)}.msg-chan{font-size:.55rem;padding:.1rem .3rem;background:var(--bg3);color:var(--cm)}
.msg-subj{color:var(--cd);font-size:.7rem;margin-top:.2rem}
.msg-meta{font-size:.6rem;color:var(--cm);margin-top:.2rem;display:flex;gap:.8rem}
.badge-delivered{color:var(--green)}.badge-failed{color:var(--red)}.badge-pending{color:var(--orange)}
.btn{font-size:.6rem;padding:.25rem .6rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:var(--bg)}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.6);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:420px;max-width:90vw}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust)}
.fr{margin-bottom:.5rem}.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.15rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.35rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:.8rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="hdr"><h1>PONY EXPRESS</h1><button class="btn btn-p" onclick="openForm()">+ Send</button></div>
<div class="main">
<div class="stats" id="stats"></div>
<div id="msgs"></div>
</div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)cm()"><div class="modal" id="mdl"></div></div>
<script>
const A='/api';let deliveries=[];
async function load(){const[d,s]=await Promise.all([fetch(A+'/deliveries').then(r=>r.json()),fetch(A+'/stats').then(r=>r.json())]);
deliveries=d.deliveries||[];
const pending=deliveries.filter(d=>d.status==='pending').length;
const delivered=deliveries.filter(d=>d.status==='delivered').length;
const failed=deliveries.filter(d=>d.status==='failed').length;
document.getElementById('stats').innerHTML='<div class="st"><div class="st-v">'+deliveries.length+'</div><div class="st-l">Total</div></div><div class="st"><div class="st-v" style="color:var(--orange)">'+pending+'</div><div class="st-l">Pending</div></div><div class="st"><div class="st-v" style="color:var(--green)">'+delivered+'</div><div class="st-l">Delivered</div></div><div class="st"><div class="st-v" style="color:var(--red)">'+failed+'</div><div class="st-l">Failed</div></div>';
render();}
function render(){if(!deliveries.length){document.getElementById('msgs').innerHTML='<div class="empty">No messages sent yet.</div>';return;}
let h='';deliveries.forEach(d=>{
h+='<div class="msg"><div class="msg-top"><div><span class="msg-to">→ '+esc(d.recipient)+'</span> <span class="msg-chan">'+d.channel+'</span></div><span class="badge-'+d.status+'">'+d.status+'</span></div>';
if(d.subject)h+='<div class="msg-subj">'+esc(d.subject)+'</div>';
h+='<div class="msg-meta"><span>'+ft(d.created_at)+'</span>';
if(d.attempts>1)h+='<span>'+d.attempts+' attempts</span>';
if(d.error_msg)h+='<span style="color:var(--red)">'+esc(d.error_msg)+'</span>';
if(d.delivered_at)h+='<span>Delivered '+ft(d.delivered_at)+'</span>';
h+='</div></div>';});
document.getElementById('msgs').innerHTML=h;}
function openForm(){document.getElementById('mdl').innerHTML='<h2>Send Message</h2><div class="fr"><label>Recipient</label><input id="f-r" placeholder="email or endpoint"></div><div class="fr"><label>Channel</label><select id="f-c"><option value="email">Email</option><option value="webhook">Webhook</option><option value="slack">Slack</option><option value="sms">SMS</option></select></div><div class="fr"><label>Subject</label><input id="f-s"></div><div class="fr"><label>Body</label><textarea id="f-b" rows="4"></textarea></div><div class="acts"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-p" onclick="sub()">Send</button></div>';document.getElementById('mbg').classList.add('open');}
async function sub(){await fetch(A+'/deliveries',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({recipient:document.getElementById('f-r').value,channel:document.getElementById('f-c').value,subject:document.getElementById('f-s').value,body:document.getElementById('f-b').value})});cm();load();}
function cm(){document.getElementById('mbg').classList.remove('open');}
function ft(t){if(!t)return'';return new Date(t).toLocaleDateString()+' '+new Date(t).toLocaleTimeString([],{hour:'2-digit',minute:'2-digit'});}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
