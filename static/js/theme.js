const url = "";

async function BuildURL() {
    const res = await fetch("../url.json");
    const config = await res.json();

    const port = config.app_port;
    const demoURL = config.url;

    url = demoURL + port;
}

const month = new Date().getMonth() + 1; 
let theme = "spring.css"; 
if (month >= 3 && month <= 5) theme = "css/spring.css"; 
else if (month >= 6 && month <= 8) theme = "css/summer.css"; 
else if (month >= 9 && month <= 11) theme = "css/autumn.css"; 
else theme = "css/winter.css"; 

document.getElementById("theme-style").href = theme;

document.getElementById("register-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const username = document.getElementById("reg-username").value;
    const password = document.getElementById("reg-password").value;

    try {
        const res = await fetch(`${url}/reg`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, password }),
        });
        const data = await res.text();
        document.getElementById("reg-message").textContent = res.ok 
            ? "✅ Registered successfully" 
            : `❌ Error: ${data}`;
    } catch {
        document.getElementById("reg-message").textContent = "❌ Network error";
    }
});

document.getElementById("login-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const username = document.getElementById("login-username").value;
    const password = document.getElementById("login-password").value;

    try {
        const res = await fetch(`${url}/login`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, password }),
        });

        const token = await res.text(); 

        if (res.ok && token) {
            localStorage.setItem("jwt", token);
            document.getElementById("login-message").textContent = "✅ Logged in successfully";
            console.log(token)
            window.location.href = "../profiles/profile.html"; 
        } else {
            document.getElementById("login-message").textContent = `❌ Error: ${token || "Unknown error"}`;
        }
    } catch (err) {
        console.error(err);
        document.getElementById("login-message").textContent = "❌ Network error";
    }
});

BuildURL();