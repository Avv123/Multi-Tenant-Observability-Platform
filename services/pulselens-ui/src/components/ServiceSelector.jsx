import { useState, useEffect, useRef } from "react";
import { queryApi } from "../lib/api";

const FALLBACK_SERVICES = [
  "api-gateway",
  "analytics-service",
  "tenant-service",
  "ingest-service",
  "processing-service",
  "alerting-service",
  "web-app"
];

export default function ServiceSelector({ token, selectedServices = [], onChange }) {
  const [isOpen, setIsOpen] = useState(false);
  const [services, setServices] = useState([]);
  const [searchQuery, setSearchQuery] = useState("");
  const containerRef = useRef(null);

  // Fetch active services with telemetry data
  useEffect(() => {
    if (!token) return;
    let active = true;
    queryApi.serviceHealth(token)
      .then(data => {
        if (!active) return;
        // Extract unique service names from health telemetry rows
        const names = Array.isArray(data) 
          ? [...new Set(data.map(row => row.service_name).filter(Boolean))]
          : [];
        
        // Merge with fallback services to ensure a robust, filled list
        const merged = [...new Set([...names, ...FALLBACK_SERVICES])].sort();
        setServices(merged);
      })
      .catch(() => {
        if (!active) return;
        setServices(FALLBACK_SERVICES);
      });

    return () => {
      active = false;
    };
  }, [token]);

  // Outside click listener to auto-close dropdown
  useEffect(() => {
    function handleOutsideClick(event) {
      if (containerRef.current && !containerRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleOutsideClick);
    return () => document.removeEventListener("mousedown", handleOutsideClick);
  }, []);

  const handleToggleSelectAll = () => {
    // Selecting "All" clears specific services
    onChange([]);
  };

  const handleToggleService = (service) => {
    let next;
    if (selectedServices.includes(service)) {
      next = selectedServices.filter(s => s !== service);
    } else {
      next = [...selectedServices, service];
    }
    onChange(next);
  };

  // Filter services by search query
  const filteredServices = services.filter(service =>
    service.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Determine button trigger label
  let triggerLabel = "All Services";
  if (selectedServices.length === 1) {
    triggerLabel = selectedServices[0];
  } else if (selectedServices.length > 1) {
    triggerLabel = `${selectedServices.length} Services Selected`;
  }

  const isAllSelected = selectedServices.length === 0;

  return (
    <div ref={containerRef} style={{ position: "relative", width: "100%" }}>
      {/* Dropdown Trigger Button */}
      <button
        type="button"
        className="form-input form-select"
        onClick={() => setIsOpen(!isOpen)}
        style={{
          textAlign: "left",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          cursor: "pointer",
          background: "var(--bg)",
          border: "1px solid var(--border)",
          color: "var(--text-1)",
          width: "100%",
          paddingRight: "2.5rem" // maintains dropdown arrow space
        }}
      >
        <span style={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
          {triggerLabel}
        </span>
      </button>

      {/* Dropdown List Panel */}
      {isOpen && (
        <div
          style={{
            position: "absolute",
            top: "100%",
            left: 0,
            right: 0,
            zIndex: 9999,
            marginTop: "0.25rem",
            background: "#0d1326", // Solid deep dark background to ensure absolute opacity
            border: "1px solid var(--border)",
            borderRadius: "var(--r-sm)",
            boxShadow: "0 10px 25px -5px rgba(0, 0, 0, 0.5), 0 8px 10px -6px rgba(0, 0, 0, 0.5)",
            padding: "0.5rem",
            display: "flex",
            flexDirection: "column",
            gap: "0.5rem",
            maxHeight: "300px",
            overflow: "hidden"
          }}
        >
          {/* Search Box */}
          <input
            type="text"
            className="form-input"
            placeholder="Search services..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            style={{
              padding: "0.4rem 0.6rem",
              fontSize: "0.82rem",
              background: "rgba(0, 0, 0, 0.3)",
              border: "1px solid rgba(255, 255, 255, 0.08)",
              borderRadius: "4px",
              color: "#fff"
            }}
          />

          {/* Service Options List */}
          <div
            style={{
              overflowY: "auto",
              display: "flex",
              flexDirection: "column",
              gap: "2px",
              maxHeight: "200px",
              paddingRight: "2px"
            }}
          >
            {/* "All Services" Option */}
            <label
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.5rem",
                padding: "0.4rem 0.5rem",
                borderRadius: "4px",
                cursor: "pointer",
                background: isAllSelected ? "rgba(255,255,255,0.06)" : "transparent",
                transition: "background 0.15s",
                fontSize: "0.82rem",
                userSelect: "none"
              }}
              onMouseEnter={e => e.currentTarget.style.background = "rgba(255,255,255,0.08)"}
              onMouseLeave={e => e.currentTarget.style.background = isAllSelected ? "rgba(255,255,255,0.06)" : "transparent"}
            >
              <input
                type="checkbox"
                checked={isAllSelected}
                onChange={handleToggleSelectAll}
                style={{
                  cursor: "pointer",
                  accentColor: "var(--primary)",
                  width: "14px",
                  height: "14px"
                }}
              />
              <span style={{ fontWeight: isAllSelected ? 600 : 400, color: isAllSelected ? "var(--primary)" : "var(--text-2)" }}>
                ✨ All Services
              </span>
            </label>

            {/* Individual Services */}
            {filteredServices.map(service => {
              const isChecked = selectedServices.includes(service);
              return (
                <label
                  key={service}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "0.5rem",
                    padding: "0.4rem 0.5rem",
                    borderRadius: "4px",
                    cursor: "pointer",
                    background: isChecked ? "rgba(59, 130, 246, 0.12)" : "transparent",
                    transition: "background 0.15s",
                    fontSize: "0.82rem",
                    userSelect: "none"
                  }}
                  onMouseEnter={e => e.currentTarget.style.background = isChecked ? "rgba(59, 130, 246, 0.18)" : "rgba(255,255,255,0.04)"}
                  onMouseLeave={e => e.currentTarget.style.background = isChecked ? "rgba(59, 130, 246, 0.12)" : "transparent"}
                >
                  <input
                    type="checkbox"
                    checked={isChecked}
                    onChange={() => handleToggleService(service)}
                    style={{
                      cursor: "pointer",
                      accentColor: "var(--primary)",
                      width: "14px",
                      height: "14px"
                    }}
                  />
                  <span style={{ fontWeight: isChecked ? 600 : 400, color: isChecked ? "#fff" : "var(--text-2)" }}>
                    {service}
                  </span>
                </label>
              );
            })}

            {filteredServices.length === 0 && (
              <div style={{ padding: "0.5rem", color: "var(--text-3)", fontSize: "0.78rem", textAlign: "center" }}>
                No services match search
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
