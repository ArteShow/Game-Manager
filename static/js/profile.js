const token = localStorage.getItem("jwt");
if (!token) window.location.href = "../index.html";

let currentProfileId = localStorage.getItem("currentProfileId") || null;

let url = "";

async function BuildURL() {
  const res = await fetch("../url.json");
  const config = await res.json();

  //Build the url
  let port = config.app_port;
  let demoURL = config.url;

  url = demoURL + port;
}

const month = new Date().getMonth() + 1;
let theme = "../css/autumn.css";
if (month >= 3 && month <= 5) theme = "../css/spring.css";
else if (month >= 6 && month <= 8) theme = "../css/summer.css";
else if (month >= 9 && month <= 11) theme = "../css/autumn.css";
else theme = "../css/winter.css";

const existingLink = document.querySelector('link[href*="autumn.css"], link[href*="spring.css"], link[href*="summer.css"], link[href*="winter.css"]');
if (existingLink) existingLink.href = theme;
else {
  const linkEl = document.createElement("link");
  linkEl.rel = "stylesheet";
  linkEl.href = theme;
  document.head.appendChild(linkEl);
}

const profilesGrid = document.getElementById("profilesContainer");
const createBtn = document.getElementById("createProfileBtn");

function cssEscapeSafe(s) {
  if (window.CSS && CSS.escape) return CSS.escape(String(s));
  return String(s).replace(/(["\\])/g, "\\$1");
}

function escapeHtml(s) {
  return String(s || "").replace(/[&<>"']/g, function(m){ return {"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[m]; });
}

function getAuthHeader() {
  return { "Content-Type": "application/json", "Authorization": "Bearer " + localStorage.getItem("jwt") };
}

async function chooseProfileOnServer(profileId) {
  const numeric = Number(profileId);
  const payloadProfileId = Number.isFinite(numeric) && !Number.isNaN(numeric) ? numeric : profileId;
  try {
    await fetch(url + "/chooseProfile", {
      method: "POST",
      headers: getAuthHeader(),
      body: JSON.stringify({ profile_id: payloadProfileId })
    });
  } catch (err) {
    console.error("chooseProfile error", err);
  }
}

function selectProfileById(id) {
  if (!id) return;
  currentProfileId = String(id);
  localStorage.setItem("currentProfileId", currentProfileId);
  document.querySelectorAll(".profileCard").forEach(c => c.classList.remove("selected"));
  const sel = `.profileCard[data-profile-id="${cssEscapeSafe(currentProfileId)}"]`;
  const el = document.querySelector(sel);
  if (el) el.classList.add("selected");
  chooseProfileOnServer(currentProfileId).catch(e => console.warn("chooseProfile failed", e));
}

async function getProfiles() {
  try {
    const res = await fetch(url + "/getAllUsersProflies", {
      method: "GET",
      headers: getAuthHeader()
    });
    if (!res.ok) {
      console.error("Failed to fetch profiles:", res.status, await res.text());
      return;
    }
    const profiles = await res.json().catch(() => []);
    profilesGrid.innerHTML = "";
    if (!profiles || profiles.length === 0) return;
    profiles.forEach(p => renderProfileCard(p));
    if (!currentProfileId) {
      const first = profiles[0];
      const firstId = first && (first.profile_id || first.profileId || first.id);
      if (firstId) selectProfileById(firstId);
    } else {
      selectProfileById(currentProfileId);
    }
  } catch (err) {
    console.error("Error fetching profiles:", err);
  }
}

async function createProfile() {
  const name = document.getElementById("profileName").value.trim();
  const description = document.getElementById("profileDesc").value.trim();
  if (!name || !description) {
    alert("Please enter both name and description.");
    return;
  }
  try {
    const res = await fetch(url + "/createProfile", {
      method: "POST",
      headers: getAuthHeader(),
      body: JSON.stringify({ name, description })
    });
    if (!res.ok) {
      const text = await res.text();
      alert("Error creating profile: " + text);
      return;
    }
    await getProfiles();
    document.getElementById("profileName").value = "";
    document.getElementById("profileDesc").value = "";
  } catch (err) {
    console.error("Error creating profile:", err);
    alert("Network error while creating profile.");
  }
}

async function deletProfile(profileId) {
  try {
    const numeric = Number(profileId);
    const payload = Number.isFinite(numeric) && !Number.isNaN(numeric) ? numeric : profileId;
    const res = await fetch(url + "/deletProfile", {
      method: "POST",
      headers: getAuthHeader(),
      body: JSON.stringify({ profile_id: payload })
    });
    if (!res.ok) {
      console.error("deletProfile failed", res.status, await res.text());
      return;
    }
    if (String(currentProfileId) === String(profileId)) {
      localStorage.removeItem("currentProfileId");
      currentProfileId = null;
    }
    await getProfiles();
  } catch (err) {
    console.error("deletProfile error", err);
  }
}

async function deletGame(profile_id, game_id) {
  try {
    const payload = { profile_id: Number(profile_id) || profile_id, game_id: Number(game_id) || game_id };
    const res = await fetch(url + "/deletGame", {
      method: "POST",
      headers: getAuthHeader(),
      body: JSON.stringify(payload)
    });
    if (!res.ok) {
      console.error("deletGame failed", res.status, await res.text());
      return;
    }
    await getProfiles();
  } catch (err) {
    console.error("deletGame error", err);
  }
}

async function submitCreateGameInline(name, profileId) {
  if (!profileId) {
    alert("No profile selected.");
    return;
  }
  const numeric = Number(profileId);
  const payloadProfileId = Number.isFinite(numeric) && !Number.isNaN(numeric) ? numeric : profileId;
  try {
    const res = await fetch(url + "/createGame", {
      method: "POST",
      headers: getAuthHeader(),
      body: JSON.stringify({ name, profile_id: payloadProfileId })
    });
    if (!res.ok) {
      const text = await res.text();
      alert("Error creating game: " + text);
      return;
    }
    await getProfiles();
  } catch (err) {
    console.error("Error creating game:", err);
    alert("Network error while creating game.");
  }
}

async function loadGamesForProfile(profileId) {
  try {
    const numeric = Number(profileId);
    const payload = Number.isFinite(numeric) && !Number.isNaN(numeric) ? numeric : profileId;
    const res = await fetch(url + "/getGames", {
      method: "POST",
      headers: getAuthHeader(),
      body: JSON.stringify({ profile_id: payload })
    });
    if (!res.ok) return [];
    const games = await res.json().catch(() => []);
    return Array.isArray(games) ? games : [];
  } catch (err) {
    console.error("Error loading games for profile:", err);
    return [];
  }
}

function renderProfileCard(p) {
  const card = document.createElement("div");
  card.className = "card profileCard";
  const id = String(p.profile_id || p.profileId || p.id || "");
  card.dataset.profileId = id;

  card.innerHTML = `
    <h2>${escapeHtml(p.name || "Unnamed")}</h2>
    <p>${escapeHtml(p.description || "")}</p>
  `;

  const deleteBtn = document.createElement("button");
  deleteBtn.className = "delete-btn";
  deleteBtn.textContent = "Delete";
  deleteBtn.addEventListener("click", async (e) => {
    e.stopPropagation();
    await deletProfile(id);
  });
  card.appendChild(deleteBtn);

  const gamesPanel = document.createElement("div");
  gamesPanel.className = "games-panel";

  const gamesHeader = document.createElement("div");
  gamesHeader.className = "games-header";

  const gamesTitle = document.createElement("span");
  gamesTitle.textContent = "Games";

  const toggleBtn = document.createElement("button");
  toggleBtn.textContent = "Create Game";
  toggleBtn.addEventListener("click", () => {
    inlineForm.classList.toggle("open");
    toggleBtn.textContent = inlineForm.classList.contains("open") ? "Close" : "Create Game";
  });

  gamesHeader.appendChild(gamesTitle);
  gamesHeader.appendChild(toggleBtn);
  gamesPanel.appendChild(gamesHeader);

  const inlineForm = document.createElement("form");
  inlineForm.className = "game-add-inline";
  inlineForm.innerHTML = `
    <input type="text" placeholder="Game name" required>
    <button type="submit">Add</button>
  `;
  const inputGame = inlineForm.querySelector("input");
  inlineForm.addEventListener("submit", async (ev) => {
    ev.preventDefault();
    const name = inputGame.value.trim();
    if (!name) return;
    await submitCreateGameInline(name, id);
  });
  gamesPanel.appendChild(inlineForm);

  const gamesListHolder = document.createElement("div");
  gamesPanel.appendChild(gamesListHolder);
  card.appendChild(gamesPanel);

  const chooseBtn = document.createElement("button");
  chooseBtn.className = "choose-btn";
  chooseBtn.textContent = "Choose Profile";
  chooseBtn.addEventListener("click", async (e) => {
    e.stopPropagation();
    localStorage.setItem("selectedProfileId", id);
    localStorage.setItem("currentProfileId", id);
    currentProfileId = String(id);
    try { await chooseProfileOnServer(id); } catch (err) { console.warn("chooseProfileOnServer failed", err); }
    window.location.href = "../roomsFolder/rooms.html";
  });
  card.appendChild(chooseBtn);

  profilesGrid.appendChild(card);

  loadGamesForProfile(id).then(games => {
    gamesListHolder.innerHTML = "";
    const ul = document.createElement("ul");
    ul.className = "games-list";
    games.forEach(g => {
      const li = document.createElement("li");
      li.className = "game-item";
      li.textContent = g.name;
      const del = document.createElement("button");
      del.className = "game-delete-btn";
      del.textContent = "Delete";
      del.addEventListener("click", async (e) => {
        e.stopPropagation();
        await deletGame(id, g.game_id);
      });
      li.appendChild(del);
      ul.appendChild(li);
    });
    gamesListHolder.appendChild(ul);
  });
}

createBtn && createBtn.addEventListener("click", createProfile);

function getCurrentProfileId() {
  return currentProfileId;
}
window.getCurrentProfileId = getCurrentProfileId;

getProfiles();

async function loginAndStoreToken(username, password) {
  try {
    const res = await fetch(url + "/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password })
    });
    if (!res.ok) {
      const t = await res.text();
      throw new Error("Login failed: " + t);
    }
    const contentType = res.headers.get("content-type") || "";
    let tok = null;
    if (contentType.includes("application/json")) {
      const j = await res.json().catch(() => null);
      if (j && typeof j === "object") tok = j.token || j.jwt || j.accessToken || j;
    } else {
      tok = await res.text().catch(() => null);
    }
    if (!tok) throw new Error("No token returned by server");
    localStorage.setItem("jwt", String(tok).trim());
    return tok;
  } catch (err) {
    console.error("loginAndStoreToken error", err);
    throw err;
  }
}

window.loginAndStoreToken = loginAndStoreToken;
BuildURL();