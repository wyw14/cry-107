package console

const styles = `
:root { color-scheme: dark; font-family: Inter, Segoe UI, sans-serif; background:#111417; color:#e7ecef; }
* { box-sizing:border-box; }
body { margin:0; min-height:100vh; background:#111417; }
header { height:58px; display:flex; align-items:center; gap:24px; padding:0 24px; border-bottom:1px solid #343b40; background:#181c20; }
.brand { font-size:18px; font-weight:700; color:#f4f6f7; }
.plant { color:#9daab2; font-size:13px; }
nav { margin-left:auto; display:flex; height:100%; }
nav a { display:flex; align-items:center; padding:0 14px; color:#aeb9bf; text-decoration:none; border-bottom:2px solid transparent; }
nav a:hover, nav a.active { color:#fff; border-bottom-color:#e8b44c; }
main { padding:22px 24px 40px; max-width:1440px; margin:0 auto; }
.titlebar { display:flex; align-items:end; justify-content:space-between; margin-bottom:18px; }
h1 { margin:0; font-size:24px; letter-spacing:0; }
.status { display:flex; align-items:center; gap:8px; color:#aeb9bf; font-size:13px; }
.status i { width:9px; height:9px; border-radius:50%; background:#4dc989; box-shadow:0 0 0 3px #183c2d; }
.process { display:grid; grid-template-columns:repeat(5,minmax(130px,1fr)); border:1px solid #343b40; background:#181c20; }
.stage { min-height:118px; padding:16px; border-right:1px solid #343b40; position:relative; }
.stage:last-child { border-right:0; }
.stage small { display:block; color:#8d9aa1; text-transform:uppercase; }
.stage strong { display:block; margin-top:12px; font-size:26px; font-weight:600; }
.stage span { color:#aeb9bf; font-size:12px; }
.grid { display:grid; grid-template-columns:2fr 1fr; gap:18px; margin-top:18px; }
.panel { border:1px solid #343b40; background:#181c20; padding:16px; min-width:0; }
.panel h2 { margin:0 0 14px; font-size:15px; color:#d9dfe2; font-weight:600; }
pre { margin:0; min-height:250px; overflow:auto; color:#b9d8e8; background:#0e1113; padding:14px; border:1px solid #293035; font:12px/1.5 Consolas,monospace; }
form { display:grid; gap:12px; }
label { display:grid; gap:6px; color:#aeb9bf; font-size:12px; }
input { width:100%; height:36px; border:1px solid #414a50; background:#111417; color:#fff; padding:0 10px; }
button { height:36px; border:1px solid #d39d38; background:#c48b24; color:#121212; font-weight:700; cursor:pointer; }
button:hover { background:#dfa83f; }
.message { min-height:20px; color:#e8b44c; font-size:12px; }
.flowline { height:8px; margin-top:20px; background:linear-gradient(90deg,#5b7786 0 42%,#e8b44c 42% 68%,#b4534b 68%); }
@media (max-width:820px) { header { height:auto; padding:12px 16px; flex-wrap:wrap; } nav { order:3; width:100%; overflow:auto; } nav a { height:38px; } main { padding:16px; } .process { grid-template-columns:1fr 1fr; } .stage { border-bottom:1px solid #343b40; } .grid { grid-template-columns:1fr; } }
`

const script = `
const stateNode = document.querySelector('#state');
const messageNode = document.querySelector('#message');
const api = document.body.dataset.api;
async function refresh() {
  try {
    const response = await fetch(api, {headers:{'Accept':'application/json'}});
    const body = await response.json();
    stateNode.textContent = JSON.stringify(body, null, 2);
    messageNode.textContent = response.ok ? 'Live control state refreshed' : (body.error || 'Request failed');
  } catch (error) { messageNode.textContent = error.message; }
}
document.querySelector('#control')?.addEventListener('submit', async event => {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const payload = Object.fromEntries([...form.entries()].map(([key,value]) => [key, Number(value)]));
  const response = await fetch(event.currentTarget.action, {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify(payload)});
  const body = await response.json();
  messageNode.textContent = response.ok ? 'Control target accepted' : (body.error || 'Control target rejected');
  await refresh();
});
refresh();
setInterval(refresh, 5000);
`
