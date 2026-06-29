/**
 * ServiceMapPage — Canvas-based force-directed service topology graph.
 * Nodes = unique services. Edges = shared trace_id co-occurrence.
 * Node color encodes health: green (<1% errors), yellow (1-10%), red (>10%).
 * Edge thickness encodes call volume.
 */
import { useEffect, useRef, useState, useCallback } from "react";
import { queryApi } from "../lib/api";

// ─── Force-directed layout (pure JS, no library) ─────────────────────────────
const REPULSION  = 8000;
const ATTRACTION = 0.04;
const DAMPING    = 0.82;
const MIN_DIST   = 60;

function initPositions(nodes, w, h) {
  return nodes.map((n, i) => {
    const angle = (2 * Math.PI * i) / nodes.length;
    const radius = Math.min(w, h) * 0.32;
    return { ...n, x: w / 2 + radius * Math.cos(angle), y: h / 2 + radius * Math.sin(angle), vx: 0, vy: 0 };
  });
}

function runForceStep(positions, edges) {
  const next = positions.map(n => ({ ...n, fx: 0, fy: 0 }));

  // Repulsion between all pairs
  for (let i = 0; i < next.length; i++) {
    for (let j = i + 1; j < next.length; j++) {
      const dx = next[j].x - next[i].x;
      const dy = next[j].y - next[i].y;
      const dist = Math.max(Math.sqrt(dx * dx + dy * dy), 1);
      const force = REPULSION / (dist * dist);
      const fx = force * (dx / dist);
      const fy = force * (dy / dist);
      next[i].fx -= fx; next[i].fy -= fy;
      next[j].fx += fx; next[j].fy += fy;
    }
  }

  // Attraction along edges
  const nameToIdx = Object.fromEntries(next.map((n, i) => [n.service_name, i]));
  for (const e of edges) {
    const si = nameToIdx[e.source];
    const ti = nameToIdx[e.target];
    if (si == null || ti == null) continue;
    const dx = next[ti].x - next[si].x;
    const dy = next[ti].y - next[si].y;
    const dist = Math.max(Math.sqrt(dx * dx + dy * dy), MIN_DIST);
    const restLen = 160 + Math.log1p(e.call_count) * 8;
    const force = ATTRACTION * (dist - restLen);
    const fx = force * (dx / dist);
    const fy = force * (dy / dist);
    next[si].fx += fx; next[si].fy += fy;
    next[ti].fx -= fx; next[ti].fy -= fy;
  }

  return next.map(n => ({
    ...n,
    vx: (n.vx + n.fx) * DAMPING,
    vy: (n.vy + n.fy) * DAMPING,
    x: n.x + n.vx + n.fx,
    y: n.y + n.vy + n.fy,
  }));
}

// ─── Color by error rate ──────────────────────────────────────────────────────
function nodeColor(errorRate) {
  if (errorRate > 10) return { fill: "rgba(239,68,68,0.85)", stroke: "#ef4444", glow: "rgba(239,68,68,0.4)" };
  if (errorRate > 1)  return { fill: "rgba(245,158,11,0.85)", stroke: "#f59e0b", glow: "rgba(245,158,11,0.35)" };
  return { fill: "rgba(99,102,241,0.85)", stroke: "#818cf8", glow: "rgba(99,102,241,0.35)" };
}

// ─── Canvas renderer ──────────────────────────────────────────────────────────
function drawGraph(canvas, positions, edges, hovered, selected) {
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  const { width, height } = canvas;
  ctx.clearRect(0, 0, width, height);

  const nameToPos = Object.fromEntries(positions.map(p => [p.service_name, p]));
  const maxCalls = Math.max(...edges.map(e => e.call_count), 1);

  // Draw edges first
  for (const e of edges) {
    const src = nameToPos[e.source];
    const tgt = nameToPos[e.target];
    if (!src || !tgt) continue;
    const thickness = 1 + (e.call_count / maxCalls) * 5;
    const alpha = 0.25 + (e.call_count / maxCalls) * 0.5;

    ctx.beginPath();
    ctx.moveTo(src.x, src.y);
    ctx.lineTo(tgt.x, tgt.y);
    ctx.strokeStyle = `rgba(148,163,184,${alpha})`;
    ctx.lineWidth = thickness;
    ctx.stroke();

    // Arrowhead at midpoint direction
    const mx = (src.x + tgt.x) / 2;
    const my = (src.y + tgt.y) / 2;
    const angle = Math.atan2(tgt.y - src.y, tgt.x - src.x);
    ctx.save();
    ctx.translate(mx, my);
    ctx.rotate(angle);
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.lineTo(-8, -4);
    ctx.lineTo(-8, 4);
    ctx.closePath();
    ctx.fillStyle = `rgba(148,163,184,${alpha + 0.2})`;
    ctx.fill();
    ctx.restore();

    // Edge label — call count
    if (e.call_count > 0) {
      ctx.save();
      ctx.font = "10px 'Inter', sans-serif";
      ctx.fillStyle = "rgba(148,163,184,0.8)";
      ctx.textAlign = "center";
      ctx.fillText(e.call_count.toLocaleString(), mx, my - 7);
      ctx.restore();
    }
  }

  // Draw nodes
  const R = 28;
  for (const p of positions) {
    const isHovered  = p.service_name === hovered;
    const isSelected = p.service_name === selected;
    const colors = nodeColor(p.error_rate || 0);
    const r = isHovered || isSelected ? R + 4 : R;

    // Glow
    if (isHovered || isSelected) {
      const grd = ctx.createRadialGradient(p.x, p.y, r * 0.5, p.x, p.y, r * 2.2);
      grd.addColorStop(0, colors.glow);
      grd.addColorStop(1, "transparent");
      ctx.beginPath();
      ctx.arc(p.x, p.y, r * 2.2, 0, Math.PI * 2);
      ctx.fillStyle = grd;
      ctx.fill();
    }

    // Node circle
    ctx.beginPath();
    ctx.arc(p.x, p.y, r, 0, Math.PI * 2);
    ctx.fillStyle = colors.fill;
    ctx.fill();
    ctx.strokeStyle = isSelected ? "#fff" : colors.stroke;
    ctx.lineWidth = isSelected ? 3 : 1.5;
    ctx.stroke();

    // Service name label
    ctx.font = `${isSelected ? 700 : 600} 11px 'Inter', sans-serif`;
    ctx.fillStyle = "#f1f5f9";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";

    // Wrap long names
    const words = p.service_name.split(/[-_]/);
    if (words.length > 1 && p.service_name.length > 12) {
      const mid = Math.ceil(words.length / 2);
      ctx.fillText(words.slice(0, mid).join("-"), p.x, p.y - 6);
      ctx.fillText(words.slice(mid).join("-"), p.x, p.y + 6);
    } else {
      ctx.fillText(p.service_name, p.x, p.y);
    }

    // Error rate badge
    if (p.error_rate > 0) {
      ctx.font = "bold 9px 'Inter', sans-serif";
      ctx.fillStyle = p.error_rate > 10 ? "#ef4444" : "#f59e0b";
      ctx.fillText(`${p.error_rate.toFixed(1)}% err`, p.x, p.y + r + 12);
    }
  }
}

// ─── Panel: node detail ───────────────────────────────────────────────────────
function NodePanel({ node, onClose }) {
  if (!node) return null;
  const colors = nodeColor(node.error_rate || 0);
  const health = node.error_rate > 10 ? "Critical" : node.error_rate > 1 ? "Degraded" : "Healthy";
  return (
    <div className="panel" style={{ padding: 0, overflow: "hidden", minWidth: "260px", flexShrink: 0 }}>
      <div style={{ padding: "0.875rem 1rem", borderBottom: "1px solid var(--border)", background: "rgba(99,102,241,0.08)", borderLeft: `4px solid ${colors.stroke}` }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
          <div>
            <div style={{ fontWeight: 700, fontSize: "0.9rem", marginBottom: "0.25rem" }}>{node.service_name}</div>
            <span style={{ fontSize: "0.7rem", fontWeight: 800, padding: "0.15rem 0.4rem", borderRadius: "3px", background: `${colors.stroke}30`, color: colors.stroke, letterSpacing: "0.05em" }}>{health.toUpperCase()}</span>
          </div>
          <button className="btn btn-ghost btn-sm" onClick={onClose}>✕</button>
        </div>
      </div>
      <div style={{ padding: "0.875rem 1rem", display: "flex", flexDirection: "column", gap: "0.6rem" }}>
        {[
          { label: "Total Calls", value: node.total_calls?.toLocaleString() || "—" },
          { label: "Error Rate", value: `${(node.error_rate || 0).toFixed(2)}%` },
          { label: "Avg Latency", value: node.avg_latency_ms ? `${Math.round(node.avg_latency_ms)}ms` : "—" },
        ].map(({ label, value }) => (
          <div key={label} style={{ display: "flex", justifyContent: "space-between", fontSize: "0.82rem" }}>
            <span style={{ color: "var(--text-3)" }}>{label}</span>
            <span style={{ fontWeight: 600 }}>{value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Main ─────────────────────────────────────────────────────────────────────
export default function ServiceMapPage({ state }) {
  const token = state.token;
  const canvasRef = useRef(null);
  const [topology, setTopology]   = useState({ nodes: [], edges: [] });
  const [positions, setPositions] = useState([]);
  const [hovered, setHovered]     = useState(null);
  const [selected, setSelected]   = useState(null);
  const [loading, setLoading]     = useState(true);
  const [error, setError]         = useState(null);
  const [lookback, setLookback]   = useState(60);
  const [simRunning, setSimRunning] = useState(false);
  const simRef = useRef(null);
  const posRef = useRef([]);
  const rafRef = useRef(null);

  // ─── Load topology ──────────────────────────────────────────────────────
  async function loadTopology(lb = lookback) {
    setLoading(true);
    setError(null);
    try {
      const data = await queryApi.serviceMap(token, lb);
      const topo = data || { nodes: [], edges: [] };
      setTopology(topo);
      const canvas = canvasRef.current;
      const w = canvas?.clientWidth || 800;
      const h = canvas?.clientHeight || 600;
      const initial = initPositions(topo.nodes, w, h);
      setPositions(initial);
      posRef.current = initial;
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }
  useEffect(() => { if (token) loadTopology(); }, [token]);

  // ─── Force simulation ────────────────────────────────────────────────────
  useEffect(() => {
    if (!posRef.current.length || !topology.edges) return;
    let step = 0;
    const MAX = 200;

    function tick() {
      if (step >= MAX) { setSimRunning(false); return; }
      posRef.current = runForceStep(posRef.current, topology.edges);
      setPositions([...posRef.current]);
      step++;
      rafRef.current = requestAnimationFrame(tick);
    }
    setSimRunning(true);
    rafRef.current = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(rafRef.current);
  }, [topology]);

  // ─── Draw ────────────────────────────────────────────────────────────────
  useEffect(() => {
    drawGraph(canvasRef.current, positions, topology.edges || [], hovered, selected);
  }, [positions, hovered, selected, topology.edges]);

  // ─── Resize canvas ───────────────────────────────────────────────────────
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ro = new ResizeObserver(() => {
      canvas.width  = canvas.clientWidth  * (window.devicePixelRatio || 1);
      canvas.height = canvas.clientHeight * (window.devicePixelRatio || 1);
      const ctx = canvas.getContext("2d");
      ctx.scale(window.devicePixelRatio || 1, window.devicePixelRatio || 1);
      drawGraph(canvas, positions, topology.edges || [], hovered, selected);
    });
    ro.observe(canvas);
    return () => ro.disconnect();
  }, [positions, hovered, selected, topology.edges]);

  // ─── Mouse interaction ───────────────────────────────────────────────────
  const hitTest = useCallback((e) => {
    const canvas = canvasRef.current;
    if (!canvas) return null;
    const rect = canvas.getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;
    const R = 32;
    for (const p of posRef.current) {
      const dx = p.x - mx;
      const dy = p.y - my;
      if (Math.sqrt(dx * dx + dy * dy) < R) return p.service_name;
    }
    return null;
  }, []);

  function handleMouseMove(e) {
    setHovered(hitTest(e));
  }
  function handleClick(e) {
    const hit = hitTest(e);
    setSelected(hit === selected ? null : hit);
  }

  const selectedNode = topology.nodes.find(n => n.service_name === selected);
  const hoveredNode  = topology.nodes.find(n => n.service_name === hovered);

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1rem", height: "100%" }}>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", flexShrink: 0 }}>
        <div>
          <h1 style={{ fontSize: "1.4rem", fontWeight: 700, marginBottom: "0.2rem" }}>Service Map</h1>
          <p style={{ color: "var(--text-2)", fontSize: "0.875rem" }}>Force-directed topology derived from distributed trace co-occurrence.</p>
        </div>
        <div style={{ display: "flex", gap: "0.5rem", alignItems: "flex-end" }}>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Lookback (mins)</label>
            <select className="form-input form-select" value={lookback} onChange={e => { const v = Number(e.target.value); setLookback(v); loadTopology(v); }} style={{ padding: "0.35rem 0.6rem", fontSize: "0.82rem" }}>
              {[15, 30, 60, 180, 360, 1440].map(m => (
                <option key={m} value={m}>{m < 60 ? `${m}m` : `${m/60}h`}</option>
              ))}
            </select>
          </div>
          <button className="btn btn-secondary btn-sm" onClick={() => loadTopology()}>↺ Refresh</button>
        </div>
      </div>

      {/* Legend */}
      <div style={{ display: "flex", gap: "1.25rem", flexShrink: 0 }}>
        {[
          { color: "#818cf8", label: "Healthy (<1% errors)" },
          { color: "#f59e0b", label: "Degraded (1-10%)" },
          { color: "#ef4444", label: "Critical (>10%)" },
        ].map(({ color, label }) => (
          <div key={label} style={{ display: "flex", alignItems: "center", gap: "0.4rem", fontSize: "0.78rem", color: "var(--text-2)" }}>
            <div style={{ width: "10px", height: "10px", borderRadius: "50%", background: color, flexShrink: 0 }} />
            {label}
          </div>
        ))}
        <div style={{ fontSize: "0.78rem", color: "var(--text-3)", marginLeft: "auto" }}>
          Edge thickness = call volume · Arrow = direction
        </div>
      </div>

      {error && (
        <div style={{ padding: "0.875rem", background: "rgba(239,68,68,0.1)", border: "1px solid rgba(239,68,68,0.25)", borderRadius: "var(--r-md)", color: "var(--danger)", fontSize: "0.875rem" }}>
          ⚠ {error}
        </div>
      )}

      {/* Graph area */}
      <div style={{ display: "flex", gap: "1rem", flex: 1, minHeight: 0 }}>
        {/* Canvas */}
        <div className="panel" style={{ flex: 1, position: "relative", padding: 0, overflow: "hidden" }}>
          {loading && (
            <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", gap: "0.75rem", color: "var(--text-2)" }}>
              <div style={{ width: "36px", height: "36px", border: "3px solid var(--border)", borderTopColor: "var(--primary-2)", borderRadius: "50%", animation: "spin 0.8s linear infinite" }} />
              <span style={{ fontSize: "0.875rem" }}>Loading topology…</span>
            </div>
          )}
          {!loading && !topology.nodes.length && (
            <div style={{ position: "absolute", inset: 0, display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", gap: "0.5rem", color: "var(--text-3)" }}>
              <div style={{ fontSize: "2.5rem" }}>🔍</div>
              <div style={{ fontWeight: 600 }}>No trace data in the selected lookback window</div>
              <div style={{ fontSize: "0.82rem" }}>Ingest some traces or increase the lookback period.</div>
            </div>
          )}
          {!loading && topology.nodes.length > 0 && (
            <>
              <canvas
                ref={canvasRef}
                style={{ width: "100%", height: "100%", display: "block", cursor: hovered ? "pointer" : "default" }}
                onMouseMove={handleMouseMove}
                onMouseLeave={() => setHovered(null)}
                onClick={handleClick}
              />
              {simRunning && (
                <div style={{ position: "absolute", top: "0.75rem", left: "0.75rem", fontSize: "0.7rem", color: "var(--text-3)", background: "rgba(0,0,0,0.4)", padding: "0.2rem 0.5rem", borderRadius: "4px" }}>
                  ⟳ Simulating…
                </div>
              )}
              {hoveredNode && !selected && (
                <div style={{ position: "absolute", top: "0.75rem", right: "0.75rem", background: "var(--bg-2)", border: "1px solid var(--border)", borderRadius: "var(--r-md)", padding: "0.5rem 0.75rem", fontSize: "0.8rem", minWidth: "160px", pointerEvents: "none" }}>
                  <div style={{ fontWeight: 700, marginBottom: "0.25rem" }}>{hoveredNode.service_name}</div>
                  <div style={{ color: "var(--text-3)" }}>{hoveredNode.total_calls?.toLocaleString()} calls · {(hoveredNode.error_rate || 0).toFixed(1)}% err</div>
                </div>
              )}
              <div style={{ position: "absolute", bottom: "0.75rem", left: "0.75rem", fontSize: "0.72rem", color: "var(--text-3)" }}>
                {topology.nodes.length} services · {topology.edges.length} connections
              </div>
            </>
          )}
        </div>

        {/* Detail panel */}
        {selectedNode && (
          <NodePanel node={selectedNode} onClose={() => setSelected(null)} />
        )}
      </div>
    </div>
  );
}
