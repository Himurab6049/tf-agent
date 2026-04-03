import { useEffect, useRef, useState } from "react";
import type { OutputLine } from "../types";
import type { RunState } from "../hooks/useTaskRunner";

const SPINNER_FRAMES = ["⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"];


interface Props {
  output: OutputLine[];
  state: RunState;
  prUrl: string | null;
  pendingQuestion?: string | null;
  onAnswer?: (answer: string) => void;
  onRetry?: () => void;
  onCancel?: () => void;
}

export function OutputPanel({ output, state, prUrl, pendingQuestion, onAnswer, onRetry, onCancel }: Props) {
  const bottomRef = useRef<HTMLDivElement>(null);
  const [spinnerFrame, setSpinnerFrame] = useState(0);

  useEffect(() => {
    if (state !== "streaming") return;
    const id = setInterval(() => setSpinnerFrame((f) => (f + 1) % SPINNER_FRAMES.length), 100);
    return () => clearInterval(id);
  }, [state]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [output]);

  const [answerDraft, setAnswerDraft] = useState("");

  const toolLines = output.filter((l) => l.kind === "tool_start" || l.kind === "tool_end");
  const textLines = output.filter((l) => l.kind === "text");
  const errorLines = output.filter((l) => l.kind === "error");

  const submitAnswer = () => {
    const a = answerDraft.trim();
    if (!a || !onAnswer) return;
    setAnswerDraft("");
    onAnswer(a);
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16, animation: "slideUp .22s ease" }}>

      {/* Status bar */}
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <span style={{ fontSize: "var(--text-xs)", fontWeight: 600, letterSpacing: ".08em", textTransform: "uppercase", color: "var(--text-3)" }}>
            Output
          </span>
          {state === "streaming" && <span className="chip chip-running">Running</span>}
          {state === "submitting" && <span className="chip chip-running">Submitting</span>}
          {state === "waiting" && <span className="chip" style={{ background: "var(--surface-2)", color: "var(--text-muted)", border: "1px solid var(--border)" }}>Waiting for you</span>}
          {state === "done" && <span className="chip chip-done">Done</span>}
          {state === "error" && <span className="chip chip-failed">Failed</span>}
        </div>
        <div style={{ display: "flex", gap: 8 }}>
          {(state === "streaming" || state === "submitting" || state === "waiting") && onCancel && (
            <button className="btn btn-sm" style={{ color: "var(--red)", borderColor: "#f5c6c2" }} onClick={onCancel}>
              ⏹ Stop
            </button>
          )}
          {(state === "done" || state === "error") && onRetry && (
            <button className="btn btn-sm" onClick={onRetry}>
              ↺ Retry
            </button>
          )}
        </div>
      </div>


      {/* Agent steps timeline */}
      {toolLines.length > 0 && (
        <div className="card" style={{ padding: "16px 20px" }}>
          <div style={{ fontSize: "var(--text-xs)", fontWeight: 600, textTransform: "uppercase", letterSpacing: ".08em", color: "var(--text-3)", marginBottom: 12 }}>
            Agent steps
          </div>
          <ul className="timeline">
            {buildToolTimeline(output).map((item, i, arr) => {
              const isLast = i === arr.length - 1;
              const showSpinner = isLast && state === "streaming";
              return (
              <li key={i} className="timeline-item" style={{ background: "none" }}>
                {showSpinner ? (
                  <span style={{ fontFamily: "system-ui", fontSize: 15, color: "var(--accent)", width: 16, display: "inline-block", userSelect: "none" }}>
                    {SPINNER_FRAMES[spinnerFrame]}
                  </span>
                ) : item.done ? (
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="var(--green)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ flexShrink: 0 }}>
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                ) : (
                  <div className="tl-dot muted" />
                )}
                <div style={{ flex: 1, minWidth: 0 }}>
                  <span style={{ fontSize: "var(--text-sm)", fontWeight: 500, color: "var(--text-2)" }}>
                    {item.label}
                  </span>
                  {item.count > 1 && (
                    <span style={{ fontSize: "var(--text-xs)", color: "var(--text-3)", marginLeft: 6 }}>
                      ×{item.count}
                    </span>
                  )}
                  {!item.done && state === "streaming" && (
                    <span style={{ fontSize: "var(--text-xs)", color: "var(--accent)", marginLeft: 8, fontStyle: "italic" }}>
                      running…
                    </span>
                  )}
                </div>
              </li>
              );
            })}
          </ul>
        </div>
      )}

      {/* Text output */}
      {textLines.length > 0 && (
        <div className="card" style={{ padding: "16px 20px" }}>
          {textLines.map((line) => (
            <TextOutput key={line.id} content={line.content} />
          ))}
          {state === "streaming" && (
            <span style={{
              display: "inline-block", width: 7, height: "1em",
              background: "var(--accent)", marginLeft: 2, verticalAlign: "text-bottom",
              animation: "blink 1s step-end infinite",
            }} />
          )}
        </div>
      )}

      {/* Errors */}
      {errorLines.map((line) => (
        <div key={line.id} className="error-box">{line.content}</div>
      ))}

      {/* Waiting for input */}
      {state === "waiting" && pendingQuestion && (
        <div className="card" style={{ borderColor: "var(--border)", animation: "slideUp .2s ease" }}>
          <div style={{ marginBottom: 16 }}>
            <div style={{ fontSize: "var(--text-sm)", fontWeight: 600, color: "var(--text)", marginBottom: 4 }}>
              <span style={{ color: "#f97316", marginRight: 6 }}>?</span>Agent needs clarification
            </div>
            <div style={{ fontSize: "var(--text-sm)", color: "var(--text-2)", whiteSpace: "pre-wrap", lineHeight: 1.6 }}>
              {pendingQuestion}
            </div>
          </div>
          <div style={{ display: "flex", gap: 8 }}>
            <textarea
              className="form-input"
              rows={2}
              placeholder="Type your answer…"
              value={answerDraft}
              onChange={(e) => setAnswerDraft(e.target.value)}
              onKeyDown={(e) => { if ((e.metaKey || e.ctrlKey) && e.key === "Enter") { e.preventDefault(); submitAnswer(); } }}
              style={{ flex: 1, resize: "none" }}
              autoFocus
            />
            <button
              className="btn btn-primary"
              onClick={submitAnswer}
              disabled={!answerDraft.trim()}
              style={{ alignSelf: "flex-end", whiteSpace: "nowrap" }}
            >
              Send →
            </button>
          </div>
          <div style={{ fontSize: "var(--text-xs)", color: "var(--text-3)", marginTop: 6 }}>
            ⌘ + Enter to send
          </div>
        </div>
      )}

      {/* PR banner */}
      {prUrl && (
        <div className="pr-card">
          <div>
            <div className="pr-card-title">
              <span style={{ marginRight: 6 }}>✓</span> Pull request opened
            </div>
            <div className="pr-card-url">{prUrl}</div>
          </div>
          <a href={prUrl} target="_blank" rel="noopener noreferrer" className="btn btn-sm">
            View PR ↗
          </a>
        </div>
      )}

      <div ref={bottomRef} />
    </div>
  );
}

/* ── Render text, detecting fenced code blocks ── */
function TextOutput({ content }: { content: string }) {
  const parts = content.split(/(```[\w]*\n[\s\S]*?```)/g);
  return (
    <>
      {parts.map((part, i) => {
        const m = part.match(/^```([\w]*)\n([\s\S]*?)```$/);
        if (m) {
          return <InlineCodeBlock key={i} lang={m[1] || "text"} code={m[2]} />;
        }
        if (!part) return null;
        return (
          <span key={i} style={{
            display: "block",
            fontFamily: "var(--font-mono)",
            fontSize: "var(--text-code)",
            color: "var(--text-2)",
            lineHeight: 1.75,
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
            marginBottom: 2,
          }}>
            {part}
          </span>
        );
      })}
    </>
  );
}

function InlineCodeBlock({ lang, code }: { lang: string; code: string }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard.writeText(code).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };
  return (
    <div style={{ margin: "10px 0", border: "1px solid var(--border)", borderRadius: 6, overflow: "hidden" }}>
      <div style={{
        display: "flex", alignItems: "center", justifyContent: "space-between",
        background: "var(--surface-2)", padding: "6px 12px", borderBottom: "1px solid var(--border)",
      }}>
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--text-3)", textTransform: "uppercase", letterSpacing: ".06em", fontWeight: 500 }}>
          {lang}
        </span>
        <button className={`copy-btn${copied ? " copied" : ""}`} onClick={copy} style={{ fontSize: 10, padding: "2px 7px" }}>
          {copied ? "✓" : "Copy"}
        </button>
      </div>
      <pre style={{
        margin: 0, padding: "12px 14px",
        fontFamily: "var(--font-mono)", fontSize: "var(--text-code)",
        color: "var(--text)", lineHeight: 1.7,
        overflowX: "auto", whiteSpace: "pre-wrap", wordBreak: "break-word",
        background: "#ffffff",
      }}>
        {code}
      </pre>
    </div>
  );
}

/* ── Map raw tool names to friendly labels ── */
function friendlyLabel(toolName: string): string {
  const n = toolName.toLowerCase();
  if (n === "repo_scan")                                    return "Comprehending";
  if (n === "generate_terraform")                           return "Generating code";
  if (n === "validate_terraform")                           return "Validating";
  if (n === "securityscan")                                 return "Security scanning";
  if (n === "createpr")                                     return "Opening PR";
  if (n === "clarifier")                                    return "Clarifying";
  if (n === "detect_drift")                                 return "Detecting drift";
  if (n === "jira_fetch")                                   return "Fetching Jira";
  if (n === "task")                                         return "Task created";
  if (n === "agent")                                        return "Orchestrating";
  if (n === "ask_user")                                     return "Asking you";
  if (n === "read" || n === "ls")                           return "Reading files";
  if (n === "glob" || n === "grep")                         return "Searching codebase";
  if (n === "bash")                                         return "Running commands";
  if (n === "write")                                        return "Writing code";
  if (n === "edit")                                         return "Editing code";
  if (n === "web_fetch" || n === "web_search")              return "Researching";
  return "Working";
}

/* ── Build timeline, deduplicating consecutive same-label steps ── */
function buildToolTimeline(lines: OutputLine[]): Array<{ label: string; done: boolean; count: number }> {
  const result: Array<{ label: string; done: boolean; count: number; rawName: string }> = [];

  for (const line of lines) {
    if (line.kind === "tool_start") {
      const label = friendlyLabel(line.toolName ?? "");
      // Skip duplicate "Task created" — already injected synthetically
      if (label === "Task created" && result.some((r) => r.label === "Task created")) continue;
      const last = result[result.length - 1];
      if (last && last.label === label && !last.done) {
        // same step already pending — don't duplicate
      } else {
        result.push({ label, done: false, count: 1, rawName: line.toolName ?? "" });
      }
    } else if (line.kind === "tool_end") {
      const rawName = line.toolName ?? "";
      const existing = [...result].reverse().find((r) => r.rawName === rawName && !r.done);
      if (existing) existing.done = true;
    }
  }

  // Collapse consecutive done entries with the same label into one with a count
  const collapsed: typeof result = [];
  for (const item of result) {
    const prev = collapsed[collapsed.length - 1];
    if (prev && prev.label === item.label && prev.done && item.done) {
      prev.count += 1;
    } else {
      collapsed.push({ ...item });
    }
  }
  return collapsed;
}
