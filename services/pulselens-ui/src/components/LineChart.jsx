export default function LineChart({ data = [], color = "var(--primary-2)" }) {
  if (!data.length) return <p style={{color:"var(--text-3)",fontSize:".82rem"}}>No data</p>;
  const W=400, H=90;
  const max = Math.max(...data.map(d=>d.value),1);
  const min = Math.min(...data.map(d=>d.value),0);
  const range = max-min||1;
  const tx = i => (i/(data.length-1||1))*W;
  const ty = v => H - ((v-min)/range)*(H-10) - 5;
  const pts = data.map((d,i) => `${tx(i)},${ty(d.value)}`).join(" ");
  const id = `lg${Math.random().toString(36).slice(2,8)}`;
  // Show at most 6 labels — thin them out to prevent concatenation overflow
  const step = Math.max(1, Math.ceil(data.length / 6));
  return (
    <div className="chart-wrap">
      <svg className="chart-svg" viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none">
        <defs>
          <linearGradient id={id} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={color} stopOpacity="0.3"/>
            <stop offset="100%" stopColor={color} stopOpacity="0"/>
          </linearGradient>
        </defs>
        <polygon points={`0,${H} ${pts} ${W},${H}`} fill={`url(#${id})`}/>
        <polyline points={pts} fill="none" stroke={color} strokeWidth="2.5"
          strokeLinejoin="round" strokeLinecap="round"/>
        {data.map((d,i) => (
          <circle key={i} cx={tx(i)} cy={ty(d.value)} r={4}
            fill={color} stroke="var(--bg)" strokeWidth="2" className="chart-bar">
            <title>{`${d.label}: ${d.value}`}</title>
          </circle>
        ))}
      </svg>
      <div className="chart-labels" style={{gridTemplateColumns:`repeat(${data.length},1fr)`}}>
        {data.map((d,i) => (
          <span key={i} style={{
            overflow:"hidden", whiteSpace:"nowrap", textOverflow:"ellipsis",
            visibility: i % step === 0 ? "visible" : "hidden"
          }} title={d.label}>{d.label}</span>
        ))}
      </div>
    </div>
  );
}

