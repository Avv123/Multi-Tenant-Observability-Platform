import { useState, useRef, useEffect, useCallback } from "react";

const STARTERS = [
  "How do I set up an alert rule?",
  "What does the Cooldown field mean?",
  "How do I ingest logs from my app?",
  "Why am I getting 401 on my API key?",
  "Explain what Aggregation does in alert rules.",
];

function TypingDots() {
  return (
    <div style={{ display: "flex", gap: "4px", padding: "0.1rem 0", alignItems: "center" }}>
      {[0, 1, 2].map(i => (
        <span key={i} style={{
          width: "6px", height: "6px", borderRadius: "50%",
          background: "var(--primary)", opacity: 0.7,
          animation: `typingBounce 1.2s ease ${i * 0.2}s infinite`,
          display: "inline-block",
        }} />
      ))}
    </div>
  );
}

function Message({ msg }) {
  const isUser = msg.role === "user";
  return (
    <div style={{
      display: "flex", flexDirection: isUser ? "row-reverse" : "row",
      gap: "0.5rem", alignItems: "flex-end", marginBottom: "0.75rem",
    }}>
      {/* Avatar */}
      <div style={{
        width: "28px", height: "28px", borderRadius: "50%", flexShrink: 0,
        background: isUser ? "var(--primary)" : "linear-gradient(135deg,#6366f1,#06b6d4)",
        display: "flex", alignItems: "center", justifyContent: "center",
        fontSize: "0.7rem", fontWeight: 700, color: "white",
      }}>
        {isUser ? "U" : "✦"}
      </div>

      {/* Bubble */}
      <div style={{
        maxWidth: "82%",
        background: isUser ? "var(--primary)" : "var(--surface-active)",
        color: isUser ? "white" : "var(--text)",
        padding: "0.6rem 0.85rem",
        borderRadius: isUser ? "14px 14px 4px 14px" : "14px 14px 14px 4px",
        fontSize: "0.82rem", lineHeight: 1.6,
        border: isUser ? "none" : "1px solid var(--border)",
      }}>
        {msg.typing ? <TypingDots /> : msg.content}
      </div>
    </div>
  );
}

export default function AiWidget({ state }) {
  const [open, setOpen] = useState(false);
  const [messages, setMessages] = useState([
    {
      role: "assistant",
      content: "Hi! I'm your PulseLens AI assistant. I can help you with alert rules, ingestion, incidents, and anything else about this platform. What would you like to know?",
    }
  ]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [aiStatus, setAiStatus] = useState("unknown"); // unknown | ready | warming
  const bottomRef = useRef(null);
  const inputRef = useRef(null);

  // Check AI status on mount
  useEffect(() => {
    fetch("http://localhost:8085/api/v1/chat/status")
      .then(r => r.json())
      .then(d => setAiStatus(d?.ready ? "ready" : "warming"))
      .catch(() => setAiStatus("warming"));
  }, []);

  // Scroll to bottom on new messages
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // Focus input when opened
  useEffect(() => {
    if (open) setTimeout(() => inputRef.current?.focus(), 200);
  }, [open]);

  const sendMessage = useCallback(async (text) => {
    const userMsg = text || input.trim();
    if (!userMsg || loading) return;
    setInput("");

    const newMessages = [...messages, { role: "user", content: userMsg }];
    setMessages(newMessages);
    setLoading(true);

    // Add typing indicator
    setMessages(m => [...m, { role: "assistant", content: "", typing: true }]);

    try {
      const res = await fetch("http://localhost:8085/api/v1/chat", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(state?.token ? { Authorization: `Bearer ${state.token}` } : {}),
        },
        body: JSON.stringify({
          message: userMsg,
          context: {
            tenant_id: state?.tenantId,
            email: state?.email,
          }
        }),
      });

      const data = await res.json();
      const reply = data?.reply || data?.message || "Sorry, I couldn't get a response. Is Ollama running?";

      setMessages(m => m.filter(x => !x.typing).concat({ role: "assistant", content: reply }));
    } catch {
      setMessages(m => m.filter(x => !x.typing).concat({
        role: "assistant",
        content: "⚠️ Couldn't reach the AI service. Make sure the `pulselens-ai-service` container is running.",
      }));
    } finally {
      setLoading(false);
    }
  }, [input, messages, loading, state]);

  const handleKey = (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  return (
    <>
      <style>{`
        @keyframes typingBounce {
          0%, 60%, 100% { transform: translateY(0); }
          30% { transform: translateY(-5px); }
        }
        @keyframes widgetPop {
          from { transform: scale(0.85) translateY(12px); opacity: 0; }
          to   { transform: scale(1) translateY(0); opacity: 1; }
        }
      `}</style>

      {/* Floating trigger button */}
      <button
        onClick={() => setOpen(o => !o)}
        title="Ask AI"
        style={{
          position: "fixed", bottom: "1.5rem", right: "1.5rem", zIndex: 800,
          width: "56px", height: "56px", borderRadius: "50%",
          background: "linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)",
          border: "none", cursor: "pointer",
          boxShadow: "0 4px 24px rgba(99,102,241,0.5)",
          display: "flex", alignItems: "center", justifyContent: "center",
          flexDirection: "column", gap: "2px",
          transition: "transform 0.2s, box-shadow 0.2s",
          color: "white",
        }}
        onMouseEnter={e => { e.currentTarget.style.transform = "scale(1.08)"; e.currentTarget.style.boxShadow = "0 6px 30px rgba(99,102,241,0.65)"; }}
        onMouseLeave={e => { e.currentTarget.style.transform = "scale(1)"; e.currentTarget.style.boxShadow = "0 4px 24px rgba(99,102,241,0.5)"; }}
      >
        {open ? (
          <span style={{ fontSize: "1.2rem", lineHeight: 1 }}>✕</span>
        ) : (
          <>
            <span style={{ fontSize: "1.3rem", lineHeight: 1 }}>✦</span>
            <span style={{ fontSize: "0.5rem", fontWeight: 700, letterSpacing: "0.03em", lineHeight: 1 }}>ASK AI</span>
          </>
        )}
      </button>

      {/* Chat panel */}
      {open && (
        <div style={{
          position: "fixed", bottom: "5rem", right: "1.5rem", zIndex: 800,
          width: "360px", height: "520px",
          background: "var(--surface)",
          border: "1px solid var(--border)",
          borderRadius: "16px",
          boxShadow: "0 16px 60px rgba(0,0,0,0.5)",
          display: "flex", flexDirection: "column", overflow: "hidden",
          animation: "widgetPop 0.22s cubic-bezier(0.34,1.56,0.64,1)",
        }}>
          {/* Header */}
          <div style={{
            padding: "0.875rem 1rem",
            background: "linear-gradient(135deg, rgba(99,102,241,0.15), rgba(6,182,212,0.08))",
            borderBottom: "1px solid var(--border)",
            display: "flex", alignItems: "center", gap: "0.75rem", flexShrink: 0,
          }}>
            <div style={{
              width: "32px", height: "32px", borderRadius: "50%",
              background: "linear-gradient(135deg,#6366f1,#06b6d4)",
              display: "flex", alignItems: "center", justifyContent: "center",
              fontSize: "0.9rem", color: "white",
            }}>✦</div>
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: "0.9rem", fontWeight: 700 }}>PulseLens AI</div>
              <div style={{ fontSize: "0.7rem", color: "var(--text-3)", display: "flex", alignItems: "center", gap: "4px" }}>
                <span style={{
                  width: "6px", height: "6px", borderRadius: "50%",
                  background: aiStatus === "ready" ? "var(--success)" : "var(--warning)",
                  display: "inline-block",
                }} />
                {aiStatus === "ready" ? "llama3.2 · Ready" : "Warming up model…"}
              </div>
            </div>
            <button
              onClick={() => setMessages([{ role: "assistant", content: "Conversation cleared. How can I help?" }])}
              style={{ background: "none", border: "none", color: "var(--text-3)", cursor: "pointer", fontSize: "0.75rem", padding: "0.25rem" }}
              title="Clear conversation"
            >
              ↺ Clear
            </button>
          </div>

          {/* Messages */}
          <div style={{ flex: 1, overflowY: "auto", padding: "1rem 0.75rem" }}>
            {messages.map((msg, i) => <Message key={i} msg={msg} />)}

            {/* Starter questions — only when only 1 message (initial) */}
            {messages.length === 1 && (
              <div style={{ marginTop: "0.5rem" }}>
                <div style={{ fontSize: "0.7rem", color: "var(--text-3)", marginBottom: "0.5rem", fontWeight: 600, textTransform: "uppercase", letterSpacing: "0.07em" }}>
                  Try asking
                </div>
                {STARTERS.map((s, i) => (
                  <button key={i} onClick={() => sendMessage(s)} style={{
                    display: "block", width: "100%", textAlign: "left",
                    background: "var(--surface-active)", border: "1px solid var(--border)",
                    borderRadius: "8px", padding: "0.45rem 0.65rem",
                    fontSize: "0.78rem", color: "var(--text-2)",
                    cursor: "pointer", marginBottom: "0.35rem",
                    transition: "border-color 0.15s, color 0.15s",
                  }}
                  onMouseEnter={e => { e.currentTarget.style.borderColor = "var(--primary)"; e.currentTarget.style.color = "var(--text)"; }}
                  onMouseLeave={e => { e.currentTarget.style.borderColor = "var(--border)"; e.currentTarget.style.color = "var(--text-2)"; }}
                  >
                    {s}
                  </button>
                ))}
              </div>
            )}

            <div ref={bottomRef} />
          </div>

          {/* Input */}
          <div style={{
            padding: "0.75rem",
            borderTop: "1px solid var(--border)",
            display: "flex", gap: "0.5rem", flexShrink: 0,
            background: "var(--bg)",
          }}>
            <input
              ref={inputRef}
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKey}
              placeholder="Ask anything about PulseLens…"
              disabled={loading}
              style={{
                flex: 1, background: "var(--surface-active)",
                border: "1px solid var(--border)", borderRadius: "8px",
                color: "var(--text)", padding: "0.5rem 0.75rem",
                fontSize: "0.82rem", outline: "none",
                transition: "border-color 0.15s",
              }}
              onFocus={e => e.target.style.borderColor = "var(--primary)"}
              onBlur={e => e.target.style.borderColor = "var(--border)"}
            />
            <button
              onClick={() => sendMessage()}
              disabled={loading || !input.trim()}
              style={{
                width: "36px", height: "36px", borderRadius: "8px",
                background: input.trim() && !loading ? "var(--primary)" : "var(--surface-active)",
                border: "1px solid var(--border)",
                color: "white", cursor: input.trim() && !loading ? "pointer" : "default",
                display: "flex", alignItems: "center", justifyContent: "center",
                fontSize: "1rem", transition: "all 0.15s", flexShrink: 0,
              }}
            >
              ↑
            </button>
          </div>
        </div>
      )}
    </>
  );
}
