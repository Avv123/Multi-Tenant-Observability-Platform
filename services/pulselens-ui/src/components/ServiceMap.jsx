import { useMemo, useState } from "react";

export default function ServiceMap({ traces }) {
  const [hoveredNode, setHoveredNode] = useState(null);

  const { nodes, edges } = useMemo(() => {
    if (!traces || !traces.length) return { nodes: [], edges: [] };

    const spanMap = {};
    const serviceCounts = {};
    const serviceErrors = {};

    // 1. Build span mapping and collect node stats
    traces.forEach(t => {
      const p = t.payload || {};
      const sid = p.span_id || t.span_id;
      const svc = t.service_name || p.service_name || "unknown-service";
      const isError = t.status === "error" || p.status === "error";

      if (sid) {
        spanMap[sid] = { svc, parent_id: p.parent_span_id || t.parent_span_id };
      }

      serviceCounts[svc] = (serviceCounts[svc] || 0) + 1;
      if (isError) serviceErrors[svc] = (serviceErrors[svc] || 0) + 1;
    });

    // 2. Build edges (caller -> callee)
    const edgeMap = {};
    Object.values(spanMap).forEach(span => {
      if (span.parent_id && spanMap[span.parent_id]) {
        const parentSvc = spanMap[span.parent_id].svc;
        if (parentSvc !== span.svc) {
          const key = `${parentSvc}->${span.svc}`;
          edgeMap[key] = (edgeMap[key] || 0) + 1;
        }
      }
    });

    // 3. Layout the graph (simple circle or layered layout)
    const uniqueServices = Object.keys(serviceCounts);
    const nodes = [];
    const centerX = 300;
    const centerY = 150;
    const radius = Math.min(100, Math.max(50, uniqueServices.length * 20));

    uniqueServices.forEach((svc, i) => {
      const angle = (i / uniqueServices.length) * 2 * Math.PI - Math.PI / 2;
      // If there's only 1 node, put it in the center
      const isCenter = uniqueServices.length === 1;
      nodes.push({
        id: svc,
        x: isCenter ? centerX : centerX + radius * Math.cos(angle),
        y: isCenter ? centerY : centerY + radius * Math.sin(angle),
        calls: serviceCounts[svc],
        errors: serviceErrors[svc] || 0,
        errorRate: ((serviceErrors[svc] || 0) / serviceCounts[svc]) * 100,
      });
    });

    const edges = Object.entries(edgeMap).map(([key, weight]) => {
      const [src, dst] = key.split("->");
      const srcNode = nodes.find(n => n.id === src);
      const dstNode = nodes.find(n => n.id === dst);
      return { id: key, src: srcNode, dst: dstNode, weight };
    }).filter(e => e.src && e.dst);

    return { nodes, edges };
  }, [traces]);

  if (!nodes.length) {
    return (
      <div style={{ display: "flex", alignItems: "center", justifyItems: "center", height: "300px", color: "var(--text-3)", fontSize: "0.85rem" }}>
        <div style={{ width: "100%", textAlign: "center" }}>Not enough trace data to build topology.</div>
      </div>
    );
  }

  return (
    <div style={{ width: "100%", height: "320px", position: "relative", overflow: "hidden", background: "var(--surface-active)", borderRadius: "8px", border: "1px solid var(--border)" }}>
      <svg width="100%" height="100%" viewBox="0 0 600 320" preserveAspectRatio="xMidYMid meet">
        <defs>
          <marker id="arrow" viewBox="0 0 10 10" refX="22" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
            <path d="M 0 0 L 10 5 L 0 10 z" fill="var(--text-3)" />
          </marker>
        </defs>

        {/* Edges */}
        {edges.map(e => {
          const isHovered = hoveredNode === e.src.id || hoveredNode === e.dst.id;
          return (
            <g key={e.id}>
              <line
                x1={e.src.x} y1={e.src.y} x2={e.dst.x} y2={e.dst.y}
                stroke={isHovered ? "var(--primary)" : "var(--border)"}
                strokeWidth={Math.min(4, 1 + Math.log10(e.weight))}
                markerEnd="url(#arrow)"
                style={{ transition: "stroke 0.2s" }}
              />
              {/* Edge label */}
              {isHovered && (
                <text
                  x={(e.src.x + e.dst.x) / 2} y={(e.src.y + e.dst.y) / 2 - 5}
                  fill="var(--text-2)" fontSize="10" textAnchor="middle"
                  style={{ pointerEvents: "none" }}
                >
                  {e.weight} calls
                </text>
              )}
            </g>
          );
        })}

        {/* Nodes */}
        {nodes.map(n => {
          const hasError = n.errors > 0;
          return (
            <g key={n.id}
              transform={`translate(${n.x}, ${n.y})`}
              onMouseEnter={() => setHoveredNode(n.id)}
              onMouseLeave={() => setHoveredNode(null)}
              style={{ cursor: "pointer", transition: "transform 0.2s" }}
            >
              <circle
                r="18"
                fill={hasError ? "rgba(239, 68, 68, 0.15)" : "var(--surface)"}
                stroke={hasError ? "var(--danger)" : "var(--primary)"}
                strokeWidth={hoveredNode === n.id ? 3 : 2}
                style={{ transition: "stroke-width 0.2s" }}
              />
              <text y="32" fill="var(--text)" fontSize="12" fontWeight="600" textAnchor="middle">{n.id}</text>
              <text y="46" fill={hasError ? "var(--danger)" : "var(--text-3)"} fontSize="9" textAnchor="middle">
                {hasError ? `${n.errorRate.toFixed(1)}% err` : `${n.calls} reqs`}
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
}
