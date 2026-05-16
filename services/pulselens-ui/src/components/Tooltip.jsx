import { useState, useRef, useEffect } from "react";

/**
 * Tooltip — wraps any element and shows a help tip on hover.
 *
 * Usage:
 *   <Tooltip text="Generates a new secret key immediately.">
 *     <button>Rotate</button>
 *   </Tooltip>
 *
 * Or as an info icon inline with a label:
 *   <label>Cooldown <Tooltip text="Min time between repeated firings." /></label>
 */
export default function Tooltip({ text, children, position = "top", maxWidth = 240 }) {
  const [visible, setVisible] = useState(false);
  const [coords, setCoords] = useState({ top: 0, left: 0 });
  const triggerRef = useRef(null);
  const tooltipRef = useRef(null);

  function show() {
    if (!triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    const gap = 8;
    let top, left;
    if (position === "top") {
      top = rect.top - gap;
      left = rect.left + rect.width / 2;
    } else if (position === "bottom") {
      top = rect.bottom + gap;
      left = rect.left + rect.width / 2;
    } else if (position === "right") {
      top = rect.top + rect.height / 2;
      left = rect.right + gap;
    } else {
      top = rect.top + rect.height / 2;
      left = rect.left - gap;
    }
    setCoords({ top, left });
    setVisible(true);
  }

  // If no children, render a ⓘ icon as the trigger
  const trigger = children ?? (
    <span style={{
      display: "inline-flex", alignItems: "center", justifyContent: "center",
      width: "14px", height: "14px", borderRadius: "50%",
      fontSize: "0.65rem", fontWeight: 700,
      color: "var(--text-3)",
      border: "1.5px solid var(--text-3)",
      cursor: "default", lineHeight: 1, flexShrink: 0,
      verticalAlign: "middle", marginLeft: "4px",
    }}>i</span>
  );

  const transformMap = {
    top: "translate(-50%, -100%)",
    bottom: "translate(-50%, 0)",
    right: "translate(0, -50%)",
    left: "translate(-100%, -50%)",
  };

  return (
    <>
      <span
        ref={triggerRef}
        onMouseEnter={show}
        onMouseLeave={() => setVisible(false)}
        onFocus={show}
        onBlur={() => setVisible(false)}
        style={{ display: "inline-flex", alignItems: "center" }}
      >
        {trigger}
      </span>

      {/* Portal-style fixed tooltip */}
      {visible && (
        <div
          ref={tooltipRef}
          role="tooltip"
          style={{
            position: "fixed",
            top: coords.top,
            left: coords.left,
            transform: transformMap[position],
            zIndex: 9999,
            background: "#1e2030",
            color: "rgba(255,255,255,0.9)",
            padding: "0.45rem 0.7rem",
            borderRadius: "6px",
            fontSize: "0.75rem",
            lineHeight: 1.55,
            maxWidth: `${maxWidth}px`,
            boxShadow: "0 4px 20px rgba(0,0,0,0.5)",
            border: "1px solid rgba(255,255,255,0.08)",
            pointerEvents: "none",
            animation: "tooltipFade 0.12s ease",
          }}
        >
          {text}
        </div>
      )}
    </>
  );
}
