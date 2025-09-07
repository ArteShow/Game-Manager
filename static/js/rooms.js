const token = localStorage.getItem("jwt");
if (!token) window.location.href = "../index.html";

const baseUrl = "http://192.168.31.239:8079";

function parseJwt(token) {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(c =>
      '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
    ).join(''));
    return JSON.parse(jsonPayload);
  } catch {
    return {};
  }
}
const payload = parseJwt(token);
const userId = payload.userID;

function getSeason(month) {
  if ([11,0,1].includes(month)) return "winter";
  if ([2,3,4].includes(month)) return "spring";
  if ([5,6,7].includes(month)) return "summer";
  if ([8,9,10].includes(month)) return "autumn";
}
const month = new Date().getMonth();
const season = getSeason(month);
const seasonLink = document.getElementById("season-css");
if (seasonLink) seasonLink.href = `../css/${season}.css`;

const wsConnections = {};
const pendingRooms = {};
let lastKnownServerRooms = [];

function escapeHtml(s) {
  return String(s || "").replace(/[&<>"']/g, m => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[m]));
}

function addPendingRoom(tempId, name) {
  pendingRooms[tempId] = { room_id: tempId, room_name: name, users: [], pending: true };
  renderRoomsWithPendingLastKnown();
}

function removePendingRoom(tempId) {
  delete pendingRooms[tempId];
  renderRoomsWithPendingLastKnown();
}

function attachRoomCardEvents(card, room) {
  const statusEl = card.querySelector(".join-status");
  const usersList = card.querySelector(".games-list");
  const currentProfileId = localStorage.getItem("currentProfileId");

  async function updateUsersList(userIDs) {
    usersList.innerHTML = "";
    for (const uID of userIDs) {
      try {
        const res = await fetch(`http://192.168.31.239:8080/getUsername`, {
          method: "POST",
          headers: { "Content-Type": "application/json", "Authorization": "Bearer " + token },
          body: JSON.stringify({ userID: uID })
        });
        if (!res.ok) throw new Error("Failed to fetch username");
        const data = await res.json();
        const li = document.createElement("li");
        li.className = "game-item";
    
        li.innerHTML = `<span class="game-name">${escapeHtml(data.username)}</span>`;
        usersList.appendChild(li);
      } catch (err) {
        console.error(err);
        const li = document.createElement("li");
        li.className = "game-item";
        li.innerHTML = `<span class="game-name">Unknown (ID: ${uID})</span>`;
        usersList.appendChild(li);
      }
    }
  }

  card.querySelector(".join-btn")?.addEventListener("click", () => {
    if (wsConnections[room.room_id]) {
      alert("Already joined!");
      return;
    }
    if (!currentProfileId) {
      alert("No profile selected!");
      return;
    }

    const wsUrl = `${baseUrl.replace("http","ws")}/ws?roomID=${encodeURIComponent(room.room_id)}&profileID=${encodeURIComponent(currentProfileId)}&token=${encodeURIComponent(token)}`;
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      statusEl.textContent = "You joined the room!";
      ws.send(JSON.stringify({ message: "Hello from client!" }));
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === "user_joined") {
          statusEl.textContent = `User ${data.user} joined. Total: ${data.users.length}`;
          console.log(data.users, "this is teh list")
          updateUsersList(data.users); 
        }
        else if (data.type === "user_left") {
          statusEl.textContent = `User ${data.user} left. Total: ${data.users.length}`;
          updateUsersList(data.users); 
        }
        else if (data.type === "task_chosen") {
          const taskDisplay = document.getElementById("taskDisplay");
          if (taskDisplay) {
            taskDisplay.textContent = `Task: ${data.task}`;
            taskDisplay.classList.add("glow-task");
          }
        }
        else if (data.type === "game_chosen") {
          const gameName = data.game?.name || "unknown";
          localStorage.setItem("currentGame", gameName);

          const gameScreen = document.getElementById("gameScreen");
          const gameTitle = document.getElementById("gameTitle");
          const playersList = document.getElementById("playersList");
          const taskDisplay = document.getElementById("taskDisplay");
          const gameEndedBtn = document.getElementById("gameEndedBtn");
          const closeBtn = document.getElementById("closeGameBtn");

          if (gameScreen && gameTitle && playersList && taskDisplay && gameEndedBtn && closeBtn) {
            document.getElementById("roomsContainer")?.parentElement?.classList.add("hidden");

            gameScreen.style.backgroundImage = 'url("../assest/unnamed.png")';
            gameScreen.style.backgroundSize = 'cover';
            gameScreen.style.backgroundPosition = 'center';

            gameTitle.textContent = `Game started: ${gameName}`;
            gameTitle.style.fontSize = "32px";

            playersList.style.display = "none";
            statusEl.style.display = "none";

            taskDisplay.textContent = "";
            taskDisplay.style.fontSize = "36px";

            gameScreen.style.display = "flex";

            gameEndedBtn.style.display = "inline-block";
            gameEndedBtn.onclick = () => {
              ws.send(JSON.stringify({ message: "TASK" }));
              taskDisplay.textContent = "Waiting for task from server...";
              gameEndedBtn.style.display = "none";
            };

            closeBtn.style.display = "inline-block";
            closeBtn.onclick = () => {
              ws.send(JSON.stringify({ message: "STOP_ROOM" }));
              if (wsConnections[room.room_id]) {
                wsConnections[room.room_id].close();
                delete wsConnections[room.room_id];
              }
              window.location.reload();
            };
          }
        }

      } catch (e) {
        console.warn("Non-JSON WS message", event.data);
      }
    };

    ws.onclose = () => {
      statusEl.textContent = "Disconnected from room.";
      delete wsConnections[room.room_id];
    };

    ws.onerror = (err) => console.error(err);

    wsConnections[room.room_id] = ws;
  });

  card.querySelector(".leave-btn")?.addEventListener("click", () => {
    const ws = wsConnections[room.room_id];
    if (!ws) {
      alert("Not in this room!");
      return;
    }
    ws.send(JSON.stringify({ message: "LEAVE" }));
    ws.close();
    statusEl.textContent = "You left the room.";
    delete wsConnections[room.room_id];
  });

  if (room.creator_id == userId) {
    card.querySelector(".start-btn")?.addEventListener("click", () => {
      const ws = wsConnections[room.room_id];
      if (!ws) {
        alert("Join the room first!");
        return;
      }
      ws.send(JSON.stringify({ message: "START" }));
      if (statusEl) statusEl.textContent = "Start message sent! Waiting for game...";
    });
  }

  card.querySelector(".delete-btn")?.addEventListener("click", async (e) => {
    e.stopPropagation();
    if (!confirm(`Delete room "${room.room_name}"?`)) return;
    try {
      const res = await fetch(`${baseUrl}/deleteRoom`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "Authorization": "Bearer " + token },
        body: JSON.stringify({ room_id: room.room_id })
      });
      if (!res.ok) throw new Error("Failed to delete room");
      if (wsConnections[room.room_id]) wsConnections[room.room_id].close();
      await loadRooms();
    } catch (err) {
      console.error(err);
      alert("Failed to delete room");
    }
  });
}



async function loadRooms() {
  try {
    const res = await fetch(`${baseUrl}/getRooms`, { headers: { "Authorization": "Bearer "+token } });
    if (!res.ok) throw new Error("Failed to fetch rooms");
    const rooms = await res.json();
    lastKnownServerRooms = Array.isArray(rooms) ? rooms : [];
    renderRoomsWithPending(rooms);
  } catch (err) { console.error(err); renderRoomsWithPending(lastKnownServerRooms || []); }
}

async function renderRoomsWithPending(servers = null) {
  const serverRooms = Array.isArray(servers) ? servers : lastKnownServerRooms || [];
  const container = document.getElementById("roomsContainer");
  if (!container) return;
  container.innerHTML = "";

  if (serverRooms.length === 0 && Object.keys(pendingRooms).length === 0) {
    container.innerHTML = "<p>No rooms available.</p>";
    return;
  }

  for (const room of serverRooms) {
    const isCreator = room.creator_id == userId;
    const card = document.createElement("div");
    card.className = "card profileCard";
    card.setAttribute("data-room-id", room.room_id);

    const usersListItems = await Promise.all((Array.isArray(room.users) ? room.users : []).map(async uID => {
      try {
        const res = await fetch(`http://192.168.31.239:8080/getUsername`, {
          method: "POST",
          headers: { "Content-Type": "application/json", "Authorization": "Bearer " + token },
          body: JSON.stringify({ userID: uID })
        });
        if (!res.ok) throw new Error("Failed to fetch username");
        const data = await res.json();
        return `<li class="game-item"><span class="game-name">${escapeHtml(data.username)}</span></li>`;
      } catch (err) {
        console.error(err);
        return `<li class="game-item"><span class="game-name">Unknown (ID: ${uID})</span></li>`;
      }
    }));

    card.innerHTML = `
      <h2>${escapeHtml(room.room_name)}</h2>
      <div class="games-panel">
        <div class="games-header"><span>Users in room</span></div>
        <ul class="games-list">
          ${usersListItems.join("")}
        </ul>
      </div>
      <button class="choose-btn join-btn">Join Room</button>
      <button class="leave-btn">Leave Room</button>
      ${isCreator ? '<button class="start-btn">Start</button>' : ''}
      <button class="delete-btn">×</button>
      <div class="join-status"></div>
    `;

    attachRoomCardEvents(card, room);
    container.appendChild(card);
  }

  Object.values(pendingRooms).forEach(p => {
    if (serverRooms.some(r => r.room_name === p.room_name)) return;
    const card = document.createElement("div");
    card.className = "card profileCard pending";
    card.setAttribute("data-room-id", p.room_id);
    card.innerHTML = `
      <h2>${escapeHtml(p.room_name)} <small style="font-size:12px;opacity:.9">(creating...)</small></h2>
      <div class="games-panel">
        <div class="games-header"><span>Users in room</span></div>
        <ul class="games-list"></ul>
      </div>
      <button class="choose-btn join-btn" disabled>Join Room</button>
      <button class="leave-btn" disabled>Leave Room</button>
      <button class="delete-btn" disabled>×</button>
      <div class="join-status">Pending creation...</div>
    `;
    container.appendChild(card);
  });
}


function renderRoomsWithPendingLastKnown() {
  renderRoomsWithPending(lastKnownServerRooms || []);
}

async function createRoomWithTimeout(roomName) {
  const tempId = `pending-${Date.now()}`;
  addPendingRoom(tempId, roomName);

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 5000);

  try {
    const res = await fetch(`${baseUrl}/createRoom`, {
      method: "POST",
      headers: { "Content-Type":"application/json", "Authorization":"Bearer "+token },
      body: JSON.stringify({ room_name: roomName }),
      signal: controller.signal
    });
    clearTimeout(timer);
    if (!res.ok) throw new Error(await res.text().catch(()=> "error"));
    await loadRooms();
    removePendingRoom(tempId);
  } catch (err) {
    clearTimeout(timer);
    const maxPollMs = 30000, pollInterval=2000, start=Date.now();
    const poll = setInterval(async ()=>{
      try {
        const r = await fetch(`${baseUrl}/getRooms`, { headers:{ "Authorization":"Bearer "+token } });
        if (r.ok) {
          const rooms = await r.json().catch(()=>[]);
          const found = rooms.find(x=>x.room_name===roomName);
          if (found) { lastKnownServerRooms=rooms; removePendingRoom(tempId); renderRoomsWithPending(rooms); clearInterval(poll); }
        }
      } catch(e){}
      if (Date.now()-start>maxPollMs) {
        clearInterval(poll);
        const pendingEl = document.querySelector(`[data-room-id="${tempId}"] .join-status`);
        if (pendingEl) pendingEl.textContent = "Creation timed out — try refresh later";
      }
    }, pollInterval);
  }
}

document.getElementById("createRoomBtn")?.addEventListener("click", async () => {
  const name = prompt("Enter room name:"); 
  if (!name) return; 
  createRoomWithTimeout(name);
});
document.getElementById("refreshRoomsBtn")?.addEventListener("click", ()=>loadRooms());

(function openAdminWS(){
  const wsUrl = `${baseUrl.replace("http","ws")}/admin/ws?token=${encodeURIComponent(token)}`;
  function connect(){
    const ws = new WebSocket(wsUrl);
    ws.onopen = ()=>console.log("Admin WS connected");
    ws.onmessage = ev => {
      try{ const upd=JSON.parse(ev.data); if (upd.action==="create"||upd.action==="delete") loadRooms(); } 
      catch(e){console.warn("Invalid admin WS data", ev.data);}
    };
    ws.onclose = ()=>{ console.log("Admin WS closed, reconnect in 2s"); setTimeout(connect,2000); };
    ws.onerror = err=>{ console.error("Admin WS error",err); ws.close(); };
  }
  connect();
})();

loadRooms();
