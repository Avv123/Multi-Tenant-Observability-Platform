export default function BarChart({ data = [], color = "var(--primary-2)" }) {
  if (!data.length) return <p style={{color:"var(--text-3)",fontSize:".82rem"}}>No data</p>;
  const max = Math.max(...data.map(d => d.value), 1);
  const W=400, H=90;
  const bw = (W - (data.length-1)*5) / data.length;
  return (
    <div className="chart-wrap">
      <svg className="chart-svg" viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none">
        <defs>
          {data.map((d,i) => (
            <linearGradient key={i} id={`bg${i}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={d.color||color} stopOpacity="1"/>
              <stop offset="100%" stopColor={d.color||color} stopOpacity="0.4"/>
            </linearGradient>
          ))}
        </defs>
        {data.map((d,i) => {
          const h = Math.max(4, (d.value/max)*H);
          const x = i*(bw+5);
          return (
            <rect key={i} x={x} y={H-h} width={bw} height={h}
              fill={`url(#bg${i})`} rx={4} className="chart-bar">
              <title>{`${d.label}: ${d.value}`}</title>
            </rect>
          );
        })}
      </svg>
      <div className="chart-labels" style={{gridTemplateColumns:`repeat(${data.length},1fr)`}}>
        {data.map((d,i) => <span key={i} className="truncate" title={d.label}>{d.label}</span>)}
      </div>
    </div>
  );
}
