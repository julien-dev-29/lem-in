package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type Room struct {
	Name    string
	X, Y    int
	IsStart bool
	IsEnd   bool
}

type Move struct {
	Ant  int
	Room string
}

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: reading stdin: %v\n", err)
		os.Exit(1)
	}

	if len(data) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: empty input")
		os.Exit(1)
	}

	parts := strings.SplitN(string(data), "\n\n", 2)
	if len(parts) < 2 {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format (no blank line separating map and turns)")
		os.Exit(1)
	}

	mapSection := parts[0]
	turnsSection := strings.TrimSpace(parts[1])

	rooms, links, antCount, startName, endName := parseMap(mapSection)
	if len(rooms) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: no rooms found in map")
		os.Exit(1)
	}

	turns := parseTurns(turnsSection)

	html := generateHTML(rooms, links, antCount, startName, endName, turns)

	if err := os.WriteFile("visualization.html", []byte(html), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: writing visualization.html: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated visualization.html")

	openBrowser("visualization.html")
}

func parseMap(s string) (map[string]*Room, [][2]string, int, string, string) {
	lines := strings.Split(s, "\n")
	rooms := make(map[string]*Room)
	var links [][2]string
	var pendingLinks []string
	antCount := 0
	startName := ""
	endName := ""
	expectStart := false
	expectEnd := false

	for i, line := range lines {
		line := strings.TrimSpace(line)
		if line == "" || line == "#" {
			continue
		}
		if line == "##start" {
			expectStart = true
			continue
		}
		if line == "##end" {
			expectEnd = true
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if i == 0 {
			antCount, _ = strconv.Atoi(line)
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			name := parts[0]
			x, errX := strconv.Atoi(parts[1])
			y, errY := strconv.Atoi(parts[2])
			if errX == nil && errY == nil {
				r := &Room{Name: name, X: x, Y: y}
				if expectStart {
					r.IsStart = true
					startName = name
					expectStart = false
				}
				if expectEnd {
					r.IsEnd = true
					endName = name
					expectEnd = false
				}
				rooms[name] = r
				continue
			}
		}
		if strings.Contains(line, "-") {
			pendingLinks = append(pendingLinks, line)
		}
	}
	for _, line := range pendingLinks {
		if a, b := splitLink(line, rooms); a != "" {
			links = append(links, [2]string{a, b})
		}
	}
	return rooms, links, antCount, startName, endName
}

func splitLink(link string, rooms map[string]*Room) (string, string) {
	for i := 1; i < len(link); i++ {
		if link[i] == '-' {
			a, b := link[:i], link[i+1:]
			if rooms[a] != nil && rooms[b] != nil {
				return a, b
			}
		}
	}
	return "", ""
}

func parseTurns(s string) [][]Move {
	lines := strings.Split(s, "\n")
	var turns [][]Move
	re := regexp.MustCompile(`L(\d+)-(\S+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := re.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}
		var moves []Move
		for _, m := range matches {
			ant, _ := strconv.Atoi(m[1])
			moves = append(moves, Move{Ant: ant, Room: m[2]})
		}
		turns = append(turns, moves)
	}
	return turns
}

func sanitizeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}

func generateHTML(rooms map[string]*Room, links [][2]string, antCount int, startName, endName string, turns [][]Move) string {
	roomNames := make([]string, 0, len(rooms))
	for name := range rooms {
		roomNames = append(roomNames, name)
	}

	startInfo := startName
	if startName == "" && len(roomNames) > 0 {
		for _, r := range rooms {
			if r.IsStart {
				startInfo = r.Name
				break
			}
		}
	}
	endInfo := endName
	if endName == "" && len(roomNames) > 0 {
		for _, r := range rooms {
			if r.IsEnd {
				endInfo = r.Name
				break
			}
		}
	}

	minX, maxX, minY, maxY := 0, 0, 0, 0
	first := true
	for _, r := range rooms {
		if first {
			minX, maxX = r.X, r.X
			minY, maxY = r.Y, r.Y
			first = false
			continue
		}
		if r.X < minX {
			minX = r.X
		}
		if r.X > maxX {
			maxX = r.X
		}
		if r.Y < minY {
			minY = r.Y
		}
		if r.Y > maxY {
			maxY = r.Y
		}
	}

	roomW := maxX - minX
	roomH := maxY - minY
	if roomW == 0 {
		roomW = 1
	}
	if roomH == 0 {
		roomH = 1
	}

	padding := 60.0
	canvasW := 800.0
	canvasH := 600.0
	drawW := canvasW - 2*padding
	drawH := canvasH - 2*padding
	scaleX := drawW / float64(roomW)
	scaleY := drawH / float64(roomH)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	cx := padding + (drawW-float64(roomW)*scale)/2.0
	cy := padding + (drawH-float64(roomH)*scale)/2.0

	roomJS := "["
	for i, name := range roomNames {
		if i > 0 {
			roomJS += ","
		}
		r := rooms[name]
		roomJS += fmt.Sprintf(`{id:"%s",x:%d,y:%d,cx:%.2f,cy:%.2f,isStart:%v,isEnd:%v}`,
			sanitizeJS(r.Name), r.X, r.Y,
			float64(r.X-minX)*scale+cx,
			float64(r.Y-minY)*scale+cy,
			r.IsStart, r.IsEnd)
	}
	roomJS += "]"

	linksJS := "["
	for i, l := range links {
		if i > 0 {
			linksJS += ","
		}
		linksJS += fmt.Sprintf(`["%s","%s"]`, sanitizeJS(l[0]), sanitizeJS(l[1]))
	}
	linksJS += "]"

	turnsJS := "["
	for i, turn := range turns {
		if i > 0 {
			turnsJS += ","
		}
		turnsJS += "{"
		for j, m := range turn {
			if j > 0 {
				turnsJS += ","
			}
			turnsJS += fmt.Sprintf(`%d:"%s"`, m.Ant, sanitizeJS(m.Room))
		}
		turnsJS += "}"
	}
	turnsJS += "]"

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Lem-in Visualizer</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{display:flex;justify-content:center;align-items:center;min-height:100vh;background:#0d1117;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Oxygen,Ubuntu,sans-serif}
.container{text-align:center}
.canvas-wrap{position:relative;border-radius:12px;overflow:hidden;box-shadow:0 8px 32px rgba(0,0,0,0.5)}
canvas{display:block}
.header{display:flex;justify-content:space-between;align-items:center;margin-bottom:10px;padding:0 4px}
.header h1{color:#e6edf3;font-size:16px;font-weight:600}
.header span{color:#8b949e;font-size:13px}
.legend{display:flex;justify-content:center;gap:20px;margin-bottom:10px}
.legend-item{display:flex;align-items:center;gap:6px;color:#8b949e;font-size:12px}
.legend-dot{width:10px;height:10px;border-radius:50%%;display:inline-block;flex-shrink:0}
.controls{display:flex;align-items:center;justify-content:center;gap:8px;margin-top:12px;flex-wrap:wrap}
.controls button{padding:7px 18px;border:1px solid #30363d;border-radius:6px;cursor:pointer;font-size:13px;font-weight:500;transition:all .15s;background:#21262d;color:#c9d1d9}
.controls button:hover{background:#30363d;border-color:#8b949e}
.controls button.active{background:#1f6feb;border-color:#1f6feb;color:#fff}
.controls button.active:hover{background:#388bfd}
.ctrl-group{display:flex;align-items:center;gap:8px;padding:4px 12px;background:#161b22;border-radius:6px;border:1px solid #30363d}
.ctrl-group label{color:#8b949e;font-size:12px;white-space:nowrap}
.ctrl-group input[type="range"]{width:80px;cursor:pointer;accent-color:#58a6ff;height:4px}
.info{color:#8b949e;font-family:monospace;font-size:13px}
#turnInfo{min-width:100px;text-align:left}
#statusInfo{min-width:140px;text-align:right}
#stepBtn{font-size:11px;padding:4px 10px}
.ants-panel{display:flex;flex-wrap:wrap;gap:4px;margin-top:8px;justify-content:center;max-width:800px}
.ant-tag{display:inline-flex;align-items:center;gap:4px;padding:2px 8px;border-radius:10px;font-size:11px;font-family:monospace;color:#fff;opacity:0.85}
.ant-tag.finished{opacity:0.3;text-decoration:line-through}
.ant-tag .ant-dot{width:6px;height:6px;border-radius:50%%;display:inline-block}
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>Lem-in Visualizer</h1>
<span>%d ants · %d turns · %d rooms</span>
</div>
<div class="legend">
<div class="legend-item"><span class="legend-dot" style="background:#3fb950"></span> Start</div>
<div class="legend-item"><span class="legend-dot" style="background:#f85149"></span> End</div>
<div class="legend-item"><span class="legend-dot" style="background:#58a6ff"></span> Room</div>
<div class="legend-item"><span class="legend-dot" style="background:#d29922"></span> Ant</div>
</div>
<div class="canvas-wrap">
<canvas id="canvas" width="800" height="600"></canvas>
</div>
<div class="controls">
<button id="playBtn">&#9654; Play</button>
<button id="resetBtn">&#8634; Reset</button>
<button id="stepBtn">Step</button>
<div class="ctrl-group">
<label>Speed</label>
<input type="range" id="speedSlider" min="0.1" max="5" step="0.1" value="1">
</div>
<span id="turnInfo" class="info">Turn 0 / %d</span>
<span id="statusInfo" class="info"></span>
</div>
<div id="antsPanel" class="ants-panel"></div>
</div>
<script>
(function(){
var canvas=document.getElementById('canvas');
var ctx=canvas.getContext('2d');
var W=800,H=600;

var rooms=%s;
var links=%s;
var turns=%s;
var antCount=%d;
var startName="%s";
var endName="%s";

var roomMap={};
rooms.forEach(function(r){roomMap[r.id]=r;});

var antPaths={};
for(var i=1;i<=antCount;i++) antPaths[i]=[startName];
for(var t=0;t<turns.length;t++){
	var turn=turns[t];
	for(var ant in turn){
		if(!antPaths[ant])antPaths[ant]=[startName];
		antPaths[ant].push(turn[ant]);
	}
}
for(var ant in antPaths){
	var p=antPaths[ant],cl=[p[0]];
	for(var j=1;j<p.length;j++) if(p[j]!==p[j-1]) cl.push(p[j]);
	antPaths[ant]=cl;
}

var maxTurn=0;
for(var ant in antPaths) if(antPaths[ant].length-1>maxTurn) maxTurn=antPaths[ant].length-1;

var palette=['#d29922','#f85149','#3fb950','#58a6ff','#bc8cff','#f0883e','#79c0ff','#ff7b72',
             '#7ee787','#a5d6ff','#ffa657','#d2a8ff','#c9d1d9','#f778ba','#56d4dd','#e3b341',
             '#ffc107','#4caf50','#2196f3','#9c27b0','#ff5722','#00bcd4','#e91e63','#8bc34a'];
var antColors=[];
for(var i=0;i<antCount;i++) antColors[i]=palette[i%%palette.length];

var state={turnIndex:0,progress:0,running:false,speed:1};
var animationId=null;

function getAntPosition(ant){
	var path=antPaths[ant];
	if(!path||path.length<2||state.turnIndex>=path.length-1) return null;
	var from=roomMap[path[state.turnIndex]];
	var to=roomMap[path[state.turnIndex+1]];
	if(!from||!to) return null;
	return {
		x:from.cx+(to.cx-from.cx)*state.progress,
		y:from.cy+(to.cy-from.cy)*state.progress
	};
}

function drawLinks(){ctx.save();
	ctx.strokeStyle='#21262d';ctx.lineWidth=2.5;
	links.forEach(function(l){
		var a=roomMap[l[0]],b=roomMap[l[1]];
		if(!a||!b)return;
		ctx.beginPath();ctx.moveTo(a.cx,a.cy);ctx.lineTo(b.cx,b.cy);ctx.stroke();
	});
	ctx.strokeStyle='rgba(48,54,61,0.5)';ctx.lineWidth=1;
	links.forEach(function(l){
		var a=roomMap[l[0]],b=roomMap[l[1]];
		if(!a||!b)return;
		ctx.beginPath();ctx.moveTo(a.cx,a.cy);ctx.lineTo(b.cx,b.cy);ctx.stroke();
	});
ctx.restore()}

function drawRooms(){ctx.save();
	rooms.forEach(function(r){
		var rad=r.isStart||r.isEnd?22:18;
		// shadow
		if(r.isStart||r.isEnd){
			ctx.shadowColor=r.isStart?'rgba(63,185,80,0.4)':'rgba(248,81,73,0.4)';
			ctx.shadowBlur=20;
		}
		// outer ring
		var grad=ctx.createRadialGradient(r.cx-5,r.cy-5,2,r.cx,r.cy,rad+2);
		if(r.isStart){grad.addColorStop(0,'#2ea043');grad.addColorStop(1,'#238636')}
		else if(r.isEnd){grad.addColorStop(0,'#da3633');grad.addColorStop(1,'#b62324')}
		else{grad.addColorStop(0,'#1f6feb');grad.addColorStop(1,'#1158c7')}
		ctx.beginPath();ctx.arc(r.cx,r.cy,rad,0,Math.PI*2);
		ctx.fillStyle=grad;ctx.fill();
		ctx.shadowBlur=0;
		// inner highlight
		var hl=ctx.createRadialGradient(r.cx-6,r.cy-6,1,r.cx,r.cy,rad-2);
		hl.addColorStop(0,'rgba(255,255,255,0.15)');
		hl.addColorStop(1,'rgba(255,255,255,0)');
		ctx.beginPath();ctx.arc(r.cx,r.cy,rad-1,0,Math.PI*2);
		ctx.fillStyle=hl;ctx.fill();
		// label
		ctx.fillStyle='#fff';
		ctx.font='bold 12px monospace';
		ctx.textAlign='center';ctx.textBaseline='middle';
		var lbl=r.id;
		if(lbl.length>10) lbl=lbl.substring(0,8)+'..';
		ctx.fillText(lbl,r.cx,r.cy);
	});
ctx.restore()}

function drawAnts(){ctx.save();
	var finishedCount=0,antStates=[];
	for(var i=1;i<=antCount;i++){
		var pos=getAntPosition(i);
		if(!pos){finishedCount++;antStates[i]='finished';continue}
		var col=antColors[i-1];
		// glow
		ctx.shadowColor=col;ctx.shadowBlur=12;
		// body
		ctx.beginPath();ctx.arc(pos.x,pos.y,7,0,Math.PI*2);
		ctx.fillStyle=col;ctx.fill();
		ctx.shadowBlur=0;
		// number
		ctx.fillStyle='#000';ctx.font='bold 8px monospace';
		ctx.textAlign='center';ctx.textBaseline='middle';
		ctx.fillText(String(i),pos.x,pos.y+0.5);
		antStates[i]=pos;
	}
	var active=antCount-finishedCount;
	document.getElementById('turnInfo').textContent='Turn '+state.turnIndex+' / '+maxTurn;
	document.getElementById('statusInfo').textContent=active+' moving | '+finishedCount+' done';
	
	// ants panel
	var panel=document.getElementById('antsPanel');
	panel.innerHTML='';
	for(var i=1;i<=antCount;i++){
		var tag=document.createElement('span');
		tag.className='ant-tag'+(antStates[i]==='finished'?' finished':'');
		var dot=document.createElement('span');
		dot.className='ant-dot';dot.style.background=antColors[i-1];
		tag.appendChild(dot);
		if(antStates[i]&&antStates[i]!=='finished'){
			tag.innerHTML+=i+'-'+antPaths[i][state.turnIndex+1];
		}else if(antStates[i]==='finished'){
			tag.innerHTML+=i+'-'+endName;
		}else{
			tag.innerHTML+=i+'-'+startName;
		}
		panel.appendChild(tag);
	}
ctx.restore()}

function draw(){
	ctx.clearRect(0,0,W,H);
	drawLinks();
	drawRooms();
	drawAnts();
}

function advanceTurn(){
	state.progress=0;
	state.turnIndex++;
	if(state.turnIndex>=maxTurn){
		state.turnIndex=maxTurn;
		state.running=false;
		document.getElementById('playBtn').textContent='\u25B6 Play';
		document.getElementById('playBtn').classList.remove('active');
		if(animationId){cancelAnimationFrame(animationId);animationId=null}
	}
}

var lastFrame=0;
function loop(timestamp){
	if(!state.running)return;
	var dt=lastFrame?(timestamp-lastFrame)/1000:0;
	lastFrame=timestamp;
	if(dt>0.1)dt=0.1;
	state.progress+=state.speed*dt*1.5;
	while(state.progress>=1&&state.running){
		advanceTurn();
	}
	draw();
	if(state.running) animationId=requestAnimationFrame(loop);
}

function togglePlay(){
	if(state.running){
		state.running=false;
		document.getElementById('playBtn').textContent='\u25B6 Play';
		document.getElementById('playBtn').classList.remove('active');
		if(animationId){cancelAnimationFrame(animationId);animationId=null}
		return;
	}
	if(state.turnIndex>=maxTurn){
		state.turnIndex=0;state.progress=0;
	}
	state.running=true;
	lastFrame=0;
	document.getElementById('playBtn').textContent='\u23F8 Pause';
	document.getElementById('playBtn').classList.add('active');
	animationId=requestAnimationFrame(loop);
}

function stepOnce(){
	if(state.running) return;
	state.progress=0;
	state.turnIndex++;
	if(state.turnIndex>maxTurn) state.turnIndex=0;
	draw();
}

function reset(){
	state.running=false;
	state.turnIndex=0;state.progress=0;
	if(animationId){cancelAnimationFrame(animationId);animationId=null}
	document.getElementById('playBtn').textContent='\u25B6 Play';
	document.getElementById('playBtn').classList.remove('active');
	draw();
}

document.getElementById('playBtn').addEventListener('click',togglePlay);
document.getElementById('resetBtn').addEventListener('click',reset);
document.getElementById('stepBtn').addEventListener('click',stepOnce);
document.getElementById('speedSlider').addEventListener('input',function(){
	state.speed=parseFloat(this.value);
});

draw();
})();
</script>
</body>
</html>`, antCount, len(turns), len(rooms), len(turns), roomJS, linksJS, turnsJS, antCount, sanitizeJS(startInfo), sanitizeJS(endInfo))
}

func openBrowser(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		absPath, _ := os.Getwd()
		absPath += "\\" + path
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", absPath)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not open browser: %v\n", err)
	}
}
