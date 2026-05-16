import { useEffect, useRef } from "react";

export default function Drawer({ open, onClose, title, description, children, width = "440px" }) {
  const drawerRef = useRef(null);

  // Close on Escape
  useEffect(() => {
    if (!open) return;
    const fn = (e) => { if (e.key === "Escape") onClose(); };
    window.addEventListener("keydown", fn);
    return () => window.removeEventListener("keydown", fn);
  }, [open, onClose]);

  // Trap focus when open
  useEffect(() => {
    if (open && drawerRef.current) {
      const el = drawerRef.current.querySelector("input,select,textarea,button");
      el?.focus();
    }
  }, [open]);

  return (
    <>
      {/* Backdrop */}
      <div
        onClick={onClose}
        style={{
          position: "fixed", inset: 0, zIndex: 900,
          background: "rgba(0,0,0,0.55)",
          backdropFilter: "blur(3px)",
          opacity: open ? 1 : 0,
          pointerEvents: open ? "all" : "none",
          transition: "opacity 0.25s",
        }}
      />

      {/* Drawer panel */}
      <div
        ref={drawerRef}
        role="dialog"
        aria-modal="true"
        aria-label={title}
        style={{
          position: "fixed", top: 0, right: 0, bottom: 0, zIndex: 901,
          width, maxWidth: "100vw",
          background: "var(--surface)",
          borderLeft: "1px solid var(--border)",
          boxShadow: "-12px 0 40px rgba(0,0,0,0.45)",
          display: "flex", flexDirection: "column",
          transform: open ? "translateX(0)" : "translateX(100%)",
          transition: "transform 0.28s cubic-bezier(0.32,0.72,0,1)",
          willChange: "transform",
        }}
      >
        {/* Header */}
        <div style={{
          display: "flex", alignItems: "flex-start", justifyContent: "space-between",
          padding: "1.5rem 1.5rem 1rem",
          borderBottom: "1px solid var(--border)",
          flexShrink: 0,
        }}>
          <div>
            <h2 style={{ fontSize: "1.05rem", fontWeight: 700, marginBottom: "0.25rem" }}>{title}</h2>
            {description && (
              <p style={{ fontSize: "0.8rem", color: "var(--text-2)", lineHeight: 1.5 }}>{description}</p>
            )}
          </div>
          <button
            onClick={onClose}
            aria-label="Close"
            style={{
              width: "30px", height: "30px", borderRadius: "var(--r-sm)",
              border: "1px solid var(--border)", background: "var(--surface-active)",
              color: "var(--text-2)", cursor: "pointer",
              display: "flex", alignItems: "center", justifyContent: "center",
              fontSize: "1rem", flexShrink: 0, marginLeft: "0.75rem",
              transition: "all 0.15s",
            }}
            onMouseEnter={e => e.currentTarget.style.color = "var(--text)"}
            onMouseLeave={e => e.currentTarget.style.color = "var(--text-2)"}
          >
            ✕
          </button>
        </div>

        {/* Scrollable body */}
        <div style={{ flex: 1, overflowY: "auto", padding: "1.5rem" }}>
          {children}
        </div>
      </div>
    </>
  );
}
