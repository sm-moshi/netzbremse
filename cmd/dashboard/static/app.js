const bytesToMbps = (value) => `${(value / 1_000_000).toFixed(1)} Mbps`;
const formatMs = (value) => `${value.toFixed(1)} ms`;

const chartColors = {
  download: "#e20074",
  upload: "#059669",
  latency: "#e11d48",
  jitter: "#8b5cf6",
  grid: "rgba(0, 0, 0, 0.08)",
  label: "#6b7280",
};

const state = {
  overview: null,
  measurements: [],
};

async function load() {
  const [overview, measurements] = await Promise.all([
    fetchJson("/api/overview"),
    fetchJson("/api/measurements?limit=72"),
  ]);

  state.overview = overview;
  state.measurements = measurements.reverse();

  renderOverview();
  renderTable();
  renderThroughputChart();
  renderLatencyChart();
}

async function fetchJson(url) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status}`);
  }
  return response.json();
}

function renderOverview() {
  const overview = state.overview;
  document.getElementById("success-rate").textContent = `${(overview.successRate * 100).toFixed(1)}%`;
  document.getElementById("avg-download").textContent = bytesToMbps(overview.averageDownload);
  document.getElementById("avg-upload").textContent = bytesToMbps(overview.averageUpload);
  document.getElementById("avg-latency").textContent = formatMs(overview.averageLatency);
  document.getElementById("last-seen").textContent = formatRelativeDate(overview.lastMeasuredAt);
  document.getElementById("last-endpoint").textContent = overview.lastEndpoint || "No endpoint recorded";
}

function renderTable() {
  const body = document.getElementById("measurements-body");
  document.getElementById("recent-count").textContent = String(state.measurements.length);
  document.getElementById("throughput-meta").textContent = `${state.measurements.length} recent measurements`;

  if (!state.measurements.length) {
    body.innerHTML = `<tr><td colspan="7">No measurements yet.</td></tr>`;
    return;
  }

  body.innerHTML = state.measurements
    .slice()
    .reverse()
    .map((item) => `
      <tr>
        <td>${formatAbsoluteDate(item.measuredAt)}</td>
        <td>${escapeHtml(item.endpoint)}</td>
        <td><span class="pill ${item.success ? "ok" : "fail"}">${item.success ? "OK" : "Failed"}</span></td>
        <td>${bytesToMbps(item.downloadBPS)}</td>
        <td>${bytesToMbps(item.uploadBPS)}</td>
        <td>${formatMs(item.latencyMS)}</td>
        <td>${formatMs(item.jitterMS)}</td>
      </tr>
    `)
    .join("");
}

function renderThroughputChart() {
  const points = state.measurements;
  const series = [
    { key: "downloadBPS", color: chartColors.download, formatter: bytesToMbps },
    { key: "uploadBPS", color: chartColors.upload, formatter: bytesToMbps },
  ];
  drawChart(document.getElementById("throughput-chart"), points, series, bytesToMbps);
}

function renderLatencyChart() {
  const points = state.measurements;
  const series = [
    { key: "latencyMS", color: chartColors.latency, formatter: formatMs },
    { key: "jitterMS", color: chartColors.jitter, formatter: formatMs },
  ];
  drawChart(document.getElementById("latency-chart"), points, series, formatMs);
}

function drawChart(svg, points, series, labelFormatter) {
  const width = 800;
  const height = 320;
  const padding = { top: 24, right: 18, bottom: 40, left: 56 };

  if (!points.length) {
    svg.innerHTML = "";
    return;
  }

  const values = points.flatMap((point) => series.map((entry) => Number(point[entry.key]) || 0));
  const maxValue = Math.max(...values, 1);
  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;

  const x = (index) => padding.left + (chartWidth * index) / Math.max(points.length - 1, 1);
  const y = (value) => padding.top + chartHeight - (value / maxValue) * chartHeight;

  const grid = Array.from({ length: 5 }, (_, idx) => {
    const fraction = idx / 4;
    const value = maxValue * (1 - fraction);
    const lineY = padding.top + chartHeight * fraction;
    return `
      <line x1="${padding.left}" y1="${lineY}" x2="${width - padding.right}" y2="${lineY}" stroke="${chartColors.grid}" stroke-width="1" />
      <text x="${padding.left - 12}" y="${lineY + 4}" fill="${chartColors.label}" font-size="11" text-anchor="end">${labelFormatter(value)}</text>
    `;
  }).join("");

  const labels = points
    .filter((_, idx) => idx === 0 || idx === points.length - 1 || idx % Math.ceil(points.length / 6) === 0)
    .map((point, idx, filtered) => {
      const originalIndex = points.indexOf(point);
      return `<text x="${x(originalIndex)}" y="${height - 12}" fill="${chartColors.label}" font-size="11" text-anchor="${idx === 0 ? "start" : idx === filtered.length - 1 ? "end" : "middle"}">${formatShortDate(point.measuredAt)}</text>`;
    })
    .join("");

  const paths = series.map((entry) => {
    const d = points.map((point, idx) => `${idx === 0 ? "M" : "L"} ${x(idx)} ${y(Number(point[entry.key]) || 0)}`).join(" ");
    return `<path d="${d}" fill="none" stroke="${entry.color}" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />`;
  }).join("");

  svg.innerHTML = `
    <rect x="0" y="0" width="${width}" height="${height}" fill="transparent"></rect>
    ${grid}
    ${paths}
    ${labels}
  `;
}

function formatRelativeDate(input) {
  if (!input) {
    return "No data";
  }
  const date = new Date(input);
  const minutes = Math.round((Date.now() - date.getTime()) / 60000);
  if (minutes < 1) {
    return "Just now";
  }
  if (minutes < 60) {
    return `${minutes} min ago`;
  }
  const hours = Math.round(minutes / 60);
  if (hours < 24) {
    return `${hours} h ago`;
  }
  const days = Math.round(hours / 24);
  return `${days} d ago`;
}

function formatAbsoluteDate(input) {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(input));
}

function formatShortDate(input) {
  return new Intl.DateTimeFormat("en-GB", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
  }).format(new Date(input));
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

load().catch((error) => {
  console.error(error);
  document.getElementById("measurements-body").innerHTML = `<tr><td colspan="7">${escapeHtml(error.message)}</td></tr>`;
});
