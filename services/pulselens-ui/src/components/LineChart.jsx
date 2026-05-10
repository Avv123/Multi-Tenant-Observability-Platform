export default function LineChart({ data, valueKey = "value", labelKey = "label", color = "#38bdf8", height = 180 }) {
  if (!data?.length) {
    return <div className="chart-empty">No data</div>;
  }

  const width = 520;
  const padding = 24;
  const values = data.map((item) => Number(item[valueKey]) || 0);
  const maxValue = Math.max(...values, 1);
  const stepX = data.length === 1 ? 0 : (width - padding * 2) / (data.length - 1);
  const points = data.map((item, index) => {
    const x = padding + index * stepX;
    const ratio = (Number(item[valueKey]) || 0) / maxValue;
    const y = height - padding - ratio * (height - padding * 2);
    return { x, y, value: Number(item[valueKey]) || 0, label: item[labelKey] };
  });
  const path = points.map((point, index) => `${index === 0 ? "M" : "L"}${point.x},${point.y}`).join(" ");

  return (
    <div className="chart-shell">
      <svg viewBox={`0 0 ${width} ${height}`} className="chart-svg" role="img" aria-label="line chart">
        <line x1={padding} y1={height - padding} x2={width - padding} y2={height - padding} stroke="#334155" strokeWidth="1" />
        <path d={path} fill="none" stroke={color} strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" />
        {points.map((point) => (
          <g key={`${point.label}-${point.x}`}>
            <circle cx={point.x} cy={point.y} r="4" fill={color} />
            <title>{`${point.label}: ${point.value}`}</title>
          </g>
        ))}
      </svg>
      <div className="chart-label-row">
        {data.map((item) => (
          <span key={`${item[labelKey]}`}>{item[labelKey]}</span>
        ))}
      </div>
    </div>
  );
}

