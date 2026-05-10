const DEFAULT_COLORS = ["#60a5fa", "#34d399", "#f59e0b", "#f87171", "#a78bfa", "#f472b6"];

export default function BarChart({ data, valueKey = "value", labelKey = "label", height = 180 }) {
  if (!data?.length) {
    return <div className="chart-empty">No data</div>;
  }

  const width = 520;
  const padding = 24;
  const values = data.map((item) => Number(item[valueKey]) || 0);
  const maxValue = Math.max(...values, 1);
  const gap = 14;
  const barWidth = Math.max(18, (width - padding * 2 - gap * (data.length - 1)) / data.length);

  return (
    <div className="chart-shell">
      <svg viewBox={`0 0 ${width} ${height}`} className="chart-svg" role="img" aria-label="bar chart">
        <line x1={padding} y1={height - padding} x2={width - padding} y2={height - padding} stroke="#334155" strokeWidth="1" />
        {data.map((item, index) => {
          const value = Number(item[valueKey]) || 0;
          const barHeight = (value / maxValue) * (height - padding * 2);
          const x = padding + index * (barWidth + gap);
          const y = height - padding - barHeight;
          return (
            <g key={`${item[labelKey]}-${index}`}>
              <rect x={x} y={y} width={barWidth} height={barHeight} rx="6" fill={item.color || DEFAULT_COLORS[index % DEFAULT_COLORS.length]} />
              <title>{`${item[labelKey]}: ${value}`}</title>
            </g>
          );
        })}
      </svg>
      <div className="chart-label-row">
        {data.map((item) => (
          <span key={`${item[labelKey]}`}>{item[labelKey]}</span>
        ))}
      </div>
    </div>
  );
}

