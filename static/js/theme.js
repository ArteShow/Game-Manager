document.addEventListener("DOMContentLoaded", () => {
    const month = new Date().getMonth(); 
    const themeLink = document.getElementById("theme-style");

    let season;
    if (month >= 2 && month <= 4) season = "spring";
    else if (month >= 5 && month <= 7) season = "summer";
    else if (month >= 8 && month <= 10) season = "autumn";
    else season = "winter";

    themeLink.href = `css/${season}.css`;
});
