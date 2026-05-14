import BarChart from "../BarChart";
import DataTable from "../DataTable";
import LineChart from "../LineChart";
import StatCard from "../StatCard";
import { buildWidgetChartRows, widgetRows } from "../../lib/widgets";

export default function WidgetRenderer({ widget, data, overview, title, color = "#34d399" }) {
  const rows = widgetRows(widget, data);

  if (widget.type === "stat") {
    return (
      <StatCard
        label={widget.metric || title || "value"}
        value={overview?.[widget.metric] ?? rows.length}
      />
    );
  }

  if (widget.type === "table") {
    const columns = Object.keys(rows[0] || {})
      .slice(0, 4)
      .map((key) => ({ key, label: key }));
    return <DataTable columns={columns} rows={rows.slice(0, 5)} />;
  }

  const chartRows = buildWidgetChartRows(widget, rows);
  if (widget.chart_type === "line") {
    return <LineChart data={chartRows} color={color} />;
  }
  return <BarChart data={chartRows} />;
}
