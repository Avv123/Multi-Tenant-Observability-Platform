import { normalizeWidgetFilters } from "./dashboardState";

export function compactTimestamp(value) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  return `${date.getHours().toString().padStart(2, "0")}:${date.getMinutes().toString().padStart(2, "0")}`;
}

export function parseJSONSafe(value, fallback) {
  if (!value) {
    return fallback;
  }
  if (typeof value !== "string") {
    return value;
  }
  try {
    return JSON.parse(value);
  } catch {
    return fallback;
  }
}

export function widgetDataset(widget, data) {
  switch (widget.dataset) {
    case "logs":
      return data.logs || [];
    case "metrics":
      return data.metrics || [];
    case "traces":
      return data.traces || [];
    case "service_health":
      return data.health || [];
    case "log_severity":
      return data.logSeverity || [];
    case "metric_series":
      return data.metricSeries || [];
    case "trace_latency":
      return data.traceLatency || [];
    default:
      return [];
  }
}

function containsFilterValue(value, filterValue) {
  if (!filterValue) {
    return true;
  }
  return String(value || "").toLowerCase().includes(String(filterValue).toLowerCase());
}

export function applyWidgetFilters(rows, filters = {}) {
  return rows.filter((row) => {
    if (filters.service_id && !containsFilterValue(row.service_id || row.service_name, filters.service_id)) {
      return false;
    }
    if (filters.environment && !containsFilterValue(row.environment, filters.environment)) {
      return false;
    }
    if (filters.severity && !containsFilterValue(row.severity, filters.severity)) {
      return false;
    }
    if (filters.metric_name && !containsFilterValue(row.metric_name, filters.metric_name)) {
      return false;
    }
    if (filters.trace_id && !containsFilterValue(row.trace_id, filters.trace_id)) {
      return false;
    }
    if (filters.search) {
      const haystack = [
        row.message,
        row.operation,
        row.service_name,
        row.metric_name,
        row.severity,
      ].join(" ").toLowerCase();
      if (!haystack.includes(String(filters.search).toLowerCase())) {
        return false;
      }
    }
    return true;
  });
}

export function buildWidgetChartRows(widget, rows) {
  if (widget.type === "stat") {
    return [];
  }
  return rows.slice(0, 10).map((row) => ({
    label:
      row.severity ||
      row.metric_name ||
      row.service_name ||
      compactTimestamp(row.bucket_start || row.occurred_at || row.last_seen_at),
    value: Number(row.average_value || row.average_duration_ms || row.event_count || row.value || 0),
      }));
}

export function widgetRows(widget, data) {
  return applyWidgetFilters(widgetDataset(widget, data), normalizeWidgetFilters(widget.dataset, widget.filters || {}));
}
