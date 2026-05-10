const NUM_NODES = 5;
const PORTS = [8080, 8081, 8082, 8083, 8084];
const RPC_COLORS = {"append_entries": '#f0c040', "request_vote": '#e6782f', "reply_success": '#40d080', "reply_fail": '#e04060'}

// Node positions in a circle
const CX = 380, CY = 380, RADIUS = 280;
const positions = [];
for (let i = 0; i < NUM_NODES; i++) {
  const angle = (i / NUM_NODES) * Math.PI * 2 - Math.PI / 2;
  positions.push({
    x: CX + RADIUS * Math.cos(angle),
    y: CY + RADIUS * Math.sin(angle)
  });
}

// State
const nodes = [];
const events = [];
let connections = {};

for (let i = 0; i < NUM_NODES; i++) {
  nodes.push({
    id: i,
    state: 'unknown',
    term: 0,
    votedFor: -1,
    log: [],
    commitIndex: 0,
    leaderID: -1,
    connected: false,
    paused: false,
    ws: null
  });
}

// Render node cards
const ring = document.getElementById('cluster-ring');
nodes.forEach((node, i) => {
  const card = document.createElement('div');
  card.className = 'node-card disconnected';
  card.id = `node-card-${i}`;

  const pos = positions[i];
  card.style.left = `${pos.x - 70}px`;
  card.style.top = `${pos.y - 55}px`;

  card.innerHTML = `
    <div class="node-id">Node ${i}</div>
    <div class="node-state disconnected" id="node-state-${i}">Offline</div>
    <div class="node-term">term <strong id="node-term-${i}">0</strong></div>
  `;

  ring.appendChild(card);
});

// draw edges between nodes
const svg = document.getElementById('connections');
svg.innerHTML = '';

nodes.forEach((node1, i) => {
  nodes.forEach((node2, j) => {
    if (i <= j) return;

    const from = positions[i];
    const to = positions[j];

    const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
    line.setAttribute('x1', from.x);
    line.setAttribute('y1', from.y);
    line.setAttribute('x2', to.x);
    line.setAttribute('y2', to.y);
    line.setAttribute('class', 'connection-line');
    svg.appendChild(line);
  });
});

// render controls
const grid = document.getElementById('pause-buttons');
grid.innerHTML = `
    <div class="controls-label">Pause</div>
    <div class="controls-buttons">
        ${nodes.map((node, i) => `
            <button class="control-node-btn" 
                    id="node-pause-${i}"
                    onclick="togglePauseNode(${i})" 
                    disabled>
                Node ${i}
            </button>
        `).join('')}
    </div>
`;

function updateNodeUI(id) {
  const node = nodes[id];
  const card = document.getElementById(`node-card-${id}`);
  const stateEl = document.getElementById(`node-state-${id}`);
  const termEl = document.getElementById(`node-term-${id}`);
  const pauseButton = document.getElementById(`node-pause-${id}`)

  card.className = `node-card ${node.connected ? node.state : 'disconnected'} ${node.paused ? 'paused' : ''}`;
  stateEl.className = `node-state ${node.connected ? node.state : 'disconnected'}`;
  stateEl.textContent = node.connected ? node.state : 'Offline';
  termEl.textContent = node.term;
  pauseButton.disabled = !node.connected;
  pauseButton.className = `control-node-btn ${node.paused ? 'paused' : ''}`
}

// animate lines travelling between nodes to represent RPCS
function animateRPC(fromId, toId, type) {
    const svg = document.getElementById('connections');
    const from = positions[fromId];
    const to = positions[toId];

    const dot = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
    dot.setAttribute('r', '3');
    dot.setAttribute('cx', from.x);
    dot.setAttribute('cy', from.y);
    dot.setAttribute('fill', RPC_COLORS[type]);
    dot.style.filter = `drop-shadow(0 0 4px ${RPC_COLORS[type]})`;
    svg.appendChild(dot);

    const duration = 400;
    const start = performance.now();

    function step(now) {
        const t = Math.min((now - start) / duration, 1);
        const ease = t * (2 - t); // ease out
        dot.setAttribute('cx', from.x + (to.x - from.x) * ease);
        dot.setAttribute('cy', from.y + (to.y - from.y) * ease);

        if (t < 1) {
            requestAnimationFrame(step);
        } else {
            dot.remove();
        }
    }

    requestAnimationFrame(step);
}

// Events
function addEvent(nodeId, msg) {
  const now = new Date();
  const time = now.toTimeString().slice(0, 8);
  events.unshift({ time, nodeId, msg });
  if (events.length > 200) events.pop();
  renderEvents();
}

function renderEvents() {
  const body = document.getElementById('event-body');
  if (events.length === 0) {
    body.innerHTML = '<div class="empty">Waiting for events...</div>';
    return;
  }

  const atBottom = body.scrollHeight - body.scrollTop - body.clientHeight < 30;

  body.innerHTML = '<div class="event-spacer"></div>' + events.slice(0, 80).reverse().map(e => `
    <div class="event">
      <span class="event-time">${e.time}</span>
      <span class="event-node n${e.nodeId}">node-${e.nodeId}</span>
      <span class="event-msg">${e.msg}</span>
    </div>
  `).join('');

  if (atBottom) {
        body.scrollTop = body.scrollHeight;
  }
}

document.getElementById('clear-events').addEventListener('click', () => {
  events.length = 0;
  renderEvents();
});

// Cluster info
function updateClusterInfo() {
  const connected = nodes.filter(n => n.connected).length;
  const leader = nodes.find(n => n.state === 'leader' && n.connected);
  const dot = document.getElementById('cluster-status');
  const text = document.getElementById('cluster-info-text');

  dot.className = `status-dot ${connected > 0 ? 'connected' : 'disconnected'}`;
  
  let info = `${connected}/${NUM_NODES} nodes`;
  if (leader) info += ` · leader: node-${leader.id}`;
  const maxTerm = Math.max(...nodes.map(n => n.term));
  if (maxTerm > 0) info += ` · term ${maxTerm}`;
  text.textContent = info;
}

// WebSocket connections
function connectNode(id) {
  const port = PORTS[id];
  try {
    const ws = new WebSocket(`ws://localhost:${port}/ws`);
    nodes[id].ws = ws;

    ws.onopen = () => {
      nodes[id].connected = true;
      addEvent(id, 'Connected');
      updateNodeUI(id); 
      updateClusterInfo();
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        if (data.kind === "state") {
          const prev = { ...nodes[id] };

          nodes[id].state = data.state || 'follower';
          nodes[id].term = data.term || 0;
          nodes[id].votedFor = data.voted_for ?? -1;
          nodes[id].log = data.log || [];
          nodes[id].commitIndex = data.commit_index || 0;
          nodes[id].leaderID = data.leader_id ?? -1;

          if (prev.state !== nodes[id].state) {
            addEvent(id, `→ ${nodes[id].state.toUpperCase()} (term ${nodes[id].term})`);
          }

          updateNodeUI(id);
          updateClusterInfo();

        } else if (data.kind === "rpc") {
          animateRPC(data.from, data.to, data.type)
        }
      } catch (e) {
        console.error(`Parse error node-${id}:`, e);
      }
    };

    ws.onclose = () => {
      nodes[id].connected = false;
      nodes[id].state = 'unknown';
      addEvent(id, 'Disconnected');
      updateNodeUI(id);
      updateClusterInfo();
      setTimeout(() => connectNode(id), 2000);
    };

    ws.onerror = () => {
      ws.close();
    };
  } catch (e) {
    setTimeout(() => connectNode(id), 2000);
  }
}

// Submit command
document.getElementById('submit-btn').addEventListener('click', submitCommand);
document.getElementById('submit-val').addEventListener('keydown', (e) => {
  if (e.key === 'Enter') submitCommand();
});

async function submitCommand() {
  const keyInput = document.getElementById('submit-key');
  const valInput = document.getElementById('submit-val');
  const key = parseInt(keyInput.value);
  const val = valInput.value.trim();

  if (isNaN(key) || !val) return;

  const leader = nodes.find(n => n.state === 'leader' && n.connected);
  if (!leader) {
    addEvent(-1, 'No leader available');
    return;
  }

  const port = PORTS[leader.id];
  console.log(JSON.stringify({ key, val }))
  try {
    const res = await fetch(`http://localhost:${port}/submit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ key, val })
    });
    if (res.ok) {
      addEvent(leader.id, `Accepted: {${key}: ${val}}`);
    } else {
      addEvent(leader.id, `Rejected: ${res.status}`);
    }
  } catch (e) {
    addEvent(leader.id, `Submit failed: ${e.message}`);
  }

  keyInput.value = '';
  valInput.value = '';
  keyInput.focus();
}

async function togglePauseNode(id) {
    const node = nodes[id];
    const port = PORTS[id];
    const action = node.paused ? 'resume' : 'pause';

    try {
        await fetch(`http://localhost:${port}/${action}`, { method: 'POST' });
        node.paused = !node.paused;
    } catch (e) {
        console.error(e);
    }
}


// Init
for (let i = 0; i < NUM_NODES; i++) {
  updateNodeUI(i);
  connectNode(i);
}
updateClusterInfo();
renderEvents();
renderControls();