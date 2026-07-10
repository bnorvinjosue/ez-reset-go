// Frontend logic for ez-reset. Talks to the Go backend via window.go bindings.

const $ = (sel) => document.querySelector(sel);

const state = {
  printers: [],
  selected: null,
};

function toast(message, isError = false) {
  const el = $("#toast");
  el.textContent = message;
  el.className = "toast" + (isError ? " error" : "");
  setTimeout(() => el.classList.add("hidden"), 3500);
}

async function refreshPrinters() {
  const list = $("#printerList");
  list.innerHTML = '<div class="empty">Scanning…</div>';
  try {
    const printers = await window.go.app.App.ListPrinters();
    state.printers = printers;
    renderPrinters();
  } catch (err) {
    list.innerHTML = "";
    const div = document.createElement("div");
    div.className = "empty";
    div.textContent = String(err);
    list.appendChild(div);
    toast(String(err), true);
  }
}

function renderPrinters() {
  const list = $("#printerList");
  list.innerHTML = "";
  if (state.printers.length === 0) {
    const div = document.createElement("div");
    div.className = "empty";
    div.textContent = "No USB printers found. Connect an Epson printer over USB.";
    list.appendChild(div);
    return;
  }
  for (const p of state.printers) {
    const item = document.createElement("div");
    item.className = "printer-item" + (state.selected === p.path ? " active" : "");
    const name = p.model || p.des || "Unknown printer";
    item.innerHTML = `<div class="name">${escapeHtml(name)}</div><div class="path">${escapeHtml(p.path)}</div>`;
    item.addEventListener("click", () => selectPrinter(p));
    list.appendChild(item);
  }
}

async function selectPrinter(p) {
  state.selected = p.path;
  renderPrinters();
  $("#detailEmpty").classList.add("hidden");
  $("#detail").classList.remove("hidden");
  $("#printerName").textContent = p.model || p.des || "Unknown printer";
  $("#printerMeta").textContent = p.path;
  $("#inkLevels").innerHTML = '<div class="empty">Reading…</div>';
  $("#wasteCounters").innerHTML = "";
  $("#resetMsg").textContent = "";
  try {
    const status = await window.go.app.App.GetStatus(p.path);
    renderStatus(status);
  } catch (err) {
    $("#inkLevels").innerHTML = "";
    toast(String(err), true);
  }
}

function renderStatus(status) {
  const badge = $("#stateBadge");
  badge.textContent = status.state;
  badge.className = "state " + status.state.toLowerCase().replace(/[^a-z]/g, "");

  const ink = $("#inkLevels");
  ink.innerHTML = "";
  if (!status.inkLevels || status.inkLevels.length === 0) {
    ink.innerHTML = '<div class="empty">No ink data.</div>';
  }
  for (const lvl of status.inkLevels) {
    const color = inkColor(lvl.color);
    const el = document.createElement("div");
    el.className = "ink";
    el.innerHTML = `
      <div class="swatch" style="border-color:${color}">
        <div class="fill" style="height:${lvl.level}%;background:${color}"></div>
        <div class="pct">${lvl.level}%</div>
      </div>
      <div class="label">${lvl.color.replace(/_/g, " ")}</div>`;
    ink.appendChild(el);
  }

  const waste = $("#wasteCounters");
  waste.innerHTML = "";
  for (const w of status.wasteCounters) {
    const pct = Math.round(w.ratio * 100);
    const high = w.ratio > 0.8 ? " high" : "";
    const el = document.createElement("div");
    el.className = "waste";
    el.innerHTML = `
      <div class="top">
        <span>Counter ${w.index}</span>
        <span class="val">${w.value} / ${w.max} (${pct}%)</span>
      </div>
      <div class="bar${high}"><span style="width:${pct}%"></span></div>`;
    waste.appendChild(el);
  }
}

async function resetWaste() {
  if (!state.selected) return;
  const btn = $("#resetBtn");
  btn.disabled = true;
  try {
    const msg = await window.go.app.App.ResetWaste(state.selected);
    $("#resetMsg").textContent = msg;
    toast(msg);
    // Refresh status after reset.
    const status = await window.go.app.App.GetStatus(state.selected);
    renderStatus(status);
  } catch (err) {
    $("#resetMsg").textContent = "";
    toast(String(err), true);
  } finally {
    btn.disabled = false;
  }
}

async function loadModels() {
  try {
    const models = await window.go.app.App.Models();
    const sel = $("#modelSelect");
    sel.innerHTML = "";
    for (const m of models) {
      const opt = document.createElement("option");
      opt.value = m;
      opt.textContent = m;
      sel.appendChild(opt);
    }
  } catch (err) {
    console.error(err);
  }
}

function inkColor(name) {
  const map = {
    BLACK: "#1f2937",
    CYAN: "#06b6d4",
    MAGENTA: "#db2777",
    YELLOW: "#eab308",
    LIGHT_CYAN: "#67e8f9",
    LIGHT_MAGENTA: "#f9a8d4",
    DARK_YELLOW: "#ca8a04",
    GRAY: "#9ca3af",
    LIGHT_BLACK: "#4b5563",
    RED: "#ef4444",
    BLUE: "#3b82f6",
    GLOSS_OPTIMIZER: "#a78bfa",
    LIGHT_GRAY: "#cbd5e1",
    ORANGE: "#f97316",
  };
  return map[name] || "#38bdf8";
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, (c) => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
  }[c]));
}

function setPlatform() {
  const badge = $("#platformBadge");
  const isWin = navigator.userAgent.includes("Windows") || window.go?.app?.App === undefined ? false : false;
  // We can't reliably detect OS from the frontend; ask the backend indirectly.
  badge.textContent = "ready";
}

window.addEventListener("DOMContentLoaded", () => {
  $("#refreshBtn").addEventListener("click", refreshPrinters);
  $("#resetBtn").addEventListener("click", resetWaste);
  setPlatform();
  loadModels();
  refreshPrinters();
});
