const windowSelect = document.getElementById("window-hours");
const refreshBtn = document.getElementById("refresh-btn");

const metricsEl = document.getElementById("metrics");
const trendEl = document.getElementById("trend-chart");
const hotspotsBody = document.getElementById("hotspots-body");
const alertsBody = document.getElementById("alerts-body");
const risksBody = document.getElementById("risks-body");
const updatedEl = document.getElementById("last-updated");

const API_BASE = "";

async function fetchJSON(path) {
  const response = await fetch(`${API_BASE}${path}`);
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed (${response.status})`);
  }
  return response.json();
}

function formatTime(value) {
  try {
    return new Date(value).toLocaleString();
  } catch {
    return value;
  }
}

function metricCard(label, value) {
  return `
    <article class="metric">
      <div class="label">${label}</div>
      <div class="value">${value}</div>
    </article>
  `;
}

function levelClass(score) {
  if (score >= 90) return "critical";
  if (score >= 75) return "high";
  if (score >= 60) return "medium";
  return "low";
}

function renderSummary(summary) {
  metricsEl.innerHTML = [
    metricCard("Open Alerts", summary.open_alerts),
    metricCard("Acknowledged", summary.acknowledged_alerts),
    metricCard("Resolved (24h)", summary.resolved_last_24h),
    metricCard("Avg Risk (24h)", Number(summary.avg_risk_score_24h || 0).toFixed(2)),
  ].join("");
}

function renderTrend(points) {
  if (!points.length) {
    trendEl.innerHTML = "<p>No data in selected window.</p>";
    return;
  }

  const max = Math.max(...points.map((p) => p.avg_risk_score), 1);
  trendEl.innerHTML = points
    .map((p) => {
      const percent = Math.max(2, (p.avg_risk_score / max) * 100);
      const tip = `${new Date(p.bucket_start).toLocaleString()} | avg ${Number(p.avg_risk_score).toFixed(1)} | events ${p.risk_events}`;
      return `<div class="bar" style="height:${percent}%" data-tip="${tip}"></div>`;
    })
    .join("");
}

function renderHotspots(items) {
  hotspotsBody.innerHTML = items
    .map(
      (h) => `
      <tr>
        <td>${h.country}</td>
        <td>${h.region}</td>
        <td>${h.commodity}</td>
        <td>${Number(h.avg_risk_score).toFixed(2)}</td>
        <td>${h.active_alerts}</td>
      </tr>
    `,
    )
    .join("");
}

async function patchAlert(id, action) {
  await fetchJSON(`/v1/alerts/${id}/${action}`);
  await loadDashboard();
}

function renderAlerts(items) {
  alertsBody.innerHTML = items
    .map((a) => {
      const severity = levelClass(a.risk_score);
      return `
      <tr>
        <td>${formatTime(a.created_at)}</td>
        <td>${a.country}/${a.region}</td>
        <td>${a.commodity}</td>
        <td><span class="severity ${severity}">${Number(a.risk_score).toFixed(1)}</span></td>
        <td>
          <button class="action-btn ack" data-id="${a.id}" data-action="ack">Ack</button>
          <button class="action-btn resolve" data-id="${a.id}" data-action="resolve">Resolve</button>
        </td>
      </tr>
    `;
    })
    .join("");

  alertsBody.querySelectorAll("button").forEach((btn) => {
    btn.addEventListener("click", async () => {
      btn.disabled = true;
      try {
        await patchAlert(btn.dataset.id, btn.dataset.action);
      } catch (err) {
        alert(`Failed to update alert: ${err.message}`);
      } finally {
        btn.disabled = false;
      }
    });
  });
}

function renderRisks(items) {
  risksBody.innerHTML = items
    .map(
      (r) => `
      <tr>
        <td>${formatTime(r.timestamp)}</td>
        <td>${r.country}</td>
        <td>${r.region}</td>
        <td>${r.commodity}</td>
        <td>${Number(r.risk_score).toFixed(2)}</td>
      </tr>
    `,
    )
    .join("");
}

async function loadDashboard() {
  const hours = Number(windowSelect.value || 24);

  try {
    const [summary, series, hotspots, alerts, risks] = await Promise.all([
      fetchJSON("/v1/dashboard/summary"),
      fetchJSON(`/v1/dashboard/timeseries?hours=${hours}`),
      fetchJSON(`/v1/dashboard/hotspots?hours=${hours}&limit=15`),
      fetchJSON("/v1/alerts?status=open&limit=20"),
      fetchJSON("/v1/risks?limit=20"),
    ]);

    renderSummary(summary);
    renderTrend(series.items || []);
    renderHotspots(hotspots.items || []);
    renderAlerts(alerts.items || []);
    renderRisks(risks.items || []);
    updatedEl.textContent = `Updated: ${new Date().toLocaleString()}`;
  } catch (err) {
    console.error(err);
    alert(`Dashboard load failed: ${err.message}`);
  }
}

refreshBtn.addEventListener("click", loadDashboard);
windowSelect.addEventListener("change", loadDashboard);

loadDashboard();
setInterval(loadDashboard, 30000);
