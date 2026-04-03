import { useEffect, useRef, useState } from "react";
import { getSettings, saveSettings, updateMe } from "../lib/api";
import type { UserInfo, UserSettings } from "../types";

interface Props {
  userInfo: UserInfo | null;
  onClose: () => void;
  onLogout: () => void;
  onUserUpdate: (info: UserInfo) => void;
}

export function SettingsPanel({ userInfo, onClose, onLogout, onUserUpdate }: Props) {
  const [settings, setSettings] = useState<UserSettings | null>(null);
  const [displayName, setDisplayName] = useState(userInfo?.username ?? "");
  const [ghToken, setGhToken] = useState("");
  const [jiraToken, setJiraToken] = useState("");
  const [jiraDomain, setJiraDomain] = useState("");
  const [jiraEmail, setJiraEmail] = useState("");
  const [saving, setSaving] = useState(false);
  const [feedback, setFeedback] = useState<{ ok: boolean; msg: string } | null>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    getSettings()
      .then((s) => {
        setSettings(s);
        setJiraDomain(s.atlassian_domain ?? "");
        setJiraEmail(s.atlassian_email ?? "");
      })
      .catch(() => {});
  }, []);

  // Close on Escape
  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === "Escape") onClose(); };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onClose]);

  const handleSave = async () => {
    setSaving(true);
    setFeedback(null);
    try {
      const trimmed = displayName.trim();
      if (trimmed && trimmed !== userInfo?.username) {
        const updated = await updateMe(trimmed);
        onUserUpdate(updated);
      }
      await saveSettings({
        github_token: ghToken.trim() || undefined,
        atlassian_token: jiraToken.trim() || undefined,
        atlassian_domain: jiraDomain.trim() || undefined,
        atlassian_email: jiraEmail.trim() || undefined,
      });
      setGhToken("");
      setJiraToken("");
      const refreshed = await getSettings();
      setSettings(refreshed);
      setFeedback({ ok: true, msg: "Saved." });
      setTimeout(() => setFeedback(null), 2500);
    } catch {
      setFeedback({ ok: false, msg: "Failed to save." });
    } finally {
      setSaving(false);
    }
  };

  const avatarLetter = userInfo?.username?.[0]?.toUpperCase() ?? "U";

  return (
    <>
      {/* Backdrop */}
      <div
        onClick={onClose}
        style={{ position: "fixed", inset: 0, zIndex: 49 }}
      />

      {/* Panel — slides up from the sidebar bottom */}
      <div
        ref={panelRef}
        style={{
          position: "fixed",
          left: 0,
          bottom: 0,
          width: 240,
          background: "var(--surface)",
          borderTop: "1px solid var(--border)",
          borderRight: "1px solid var(--border)",
          borderRadius: "12px 12px 0 0",
          boxShadow: "0 -8px 32px rgba(0,0,0,.14)",
          zIndex: 50,
          animation: "slideUp .18s ease",
          overflow: "hidden",
        }}
      >
        {/* User header */}
        <div style={{
          display: "flex", alignItems: "center", gap: 10,
          padding: "16px 16px 14px",
          borderBottom: "1px solid var(--border)",
        }}>
          <div style={{
            width: 34, height: 34, borderRadius: "50%", flexShrink: 0,
            background: "var(--accent)", color: "#fff",
            display: "flex", alignItems: "center", justifyContent: "center",
            fontSize: 14, fontWeight: 700,
          }}>
            {avatarLetter}
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: "var(--text-sm)", fontWeight: 600, color: "var(--text)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
              {userInfo?.username ?? "—"}
            </div>
            <div style={{ fontSize: "var(--text-xs)", color: "var(--text-3)" }}>
              {userInfo?.role ?? ""}
            </div>
          </div>
          <button
            onClick={onClose}
            style={{ background: "none", border: "none", cursor: "pointer", color: "var(--text-3)", padding: 4, borderRadius: 4, display: "flex", alignItems: "center" }}
          >
            <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        {/* Token fields */}
        <div style={{ padding: "14px 16px", display: "flex", flexDirection: "column", gap: 12 }}>

          {/* Display name */}
          <div>
            <div style={{ display: "flex", alignItems: "center", gap: 6, marginBottom: 5 }}>
              <svg width="13" height="13" fill="none" stroke="var(--text-3)" strokeWidth="2" viewBox="0 0 24 24">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
              </svg>
              <span style={{ fontSize: "var(--text-xs)", fontWeight: 600, color: "var(--text-2)" }}>Display name</span>
            </div>
            <input
              className="form-input"
              style={{ fontSize: 11, padding: "6px 9px" }}
              placeholder="Your name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              autoComplete="off"
            />
          </div>

          <div style={{ height: 1, background: "var(--border)", margin: "2px 0" }} />

          {/* GH Key */}
          <TokenField
            label="GitHub Key"
            isSet={settings?.github_token_set ?? false}
            value={ghToken}
            onChange={setGhToken}
            placeholder="ghp_…"
            icon={
              <svg width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"/>
              </svg>
            }
          />

          {/* Jira Key */}
          <TokenField
            label="Jira Key"
            isSet={settings?.atlassian_token_set ?? false}
            value={jiraToken}
            onChange={setJiraToken}
            placeholder="ATATT3x…"
            icon={
              <svg width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                <rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8M12 17v4"/>
              </svg>
            }
          />

          {/* Jira domain + email — shown only when entering a jira token */}
          {(jiraToken || settings?.atlassian_token_set) && (
            <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
              <input
                className="form-input"
                style={{ fontSize: 11, padding: "6px 9px" }}
                placeholder="Domain (myco.atlassian.net)"
                value={jiraDomain}
                onChange={(e) => setJiraDomain(e.target.value)}
              />
              <input
                className="form-input"
                style={{ fontSize: 11, padding: "6px 9px" }}
                placeholder="Email (you@company.com)"
                value={jiraEmail}
                onChange={(e) => setJiraEmail(e.target.value)}
              />
            </div>
          )}

          {/* Save row */}
          <div style={{ display: "flex", alignItems: "center", gap: 8, paddingTop: 2 }}>
            <button
              className="btn btn-primary"
              onClick={handleSave}
              disabled={saving}
              style={{ flex: 1, justifyContent: "center", padding: "7px 12px", fontSize: 12 }}
            >
              {saving ? "Saving…" : "Save"}
            </button>
            {feedback && (
              <span style={{ fontSize: 11, color: feedback.ok ? "var(--green)" : "var(--red)", whiteSpace: "nowrap" }}>
                {feedback.msg}
              </span>
            )}
          </div>
        </div>

        {/* Sign out */}
        <div style={{ padding: "0 16px 14px" }}>
          <button
            onClick={onLogout}
            style={{
              width: "100%", padding: "7px 12px",
              background: "none", border: "1px solid var(--border)",
              borderRadius: 6, cursor: "pointer",
              fontSize: "var(--text-xs)", color: "var(--text-3)",
              display: "flex", alignItems: "center", justifyContent: "center", gap: 6,
              transition: "background .12s, color .12s",
            }}
            onMouseEnter={e => { e.currentTarget.style.background = "var(--red-bg, #fdf0ee)"; e.currentTarget.style.color = "var(--red)"; e.currentTarget.style.borderColor = "var(--red)"; }}
            onMouseLeave={e => { e.currentTarget.style.background = "none"; e.currentTarget.style.color = "var(--text-3)"; e.currentTarget.style.borderColor = "var(--border)"; }}
          >
            <svg width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>
            </svg>
            Sign out
          </button>
        </div>
      </div>
    </>
  );
}

function TokenField({ label, isSet, value, onChange, placeholder, icon }: {
  label: string;
  isSet: boolean;
  value: string;
  onChange: (v: string) => void;
  placeholder: string;
  icon: React.ReactNode;
}) {
  return (
    <div>
      <div style={{ display: "flex", alignItems: "center", gap: 6, marginBottom: 5 }}>
        <span style={{ color: "var(--text-3)" }}>{icon}</span>
        <span style={{ fontSize: "var(--text-xs)", fontWeight: 600, color: "var(--text-2)" }}>{label}</span>
        <span style={{
          marginLeft: "auto",
          fontSize: 10, fontWeight: 600,
          padding: "1px 6px", borderRadius: 8,
          background: isSet ? "var(--green-bg, #edf7f1)" : "var(--surface-2, #f5f4f0)",
          color: isSet ? "var(--green)" : "var(--text-3)",
          border: isSet ? "1px solid #a7d9bc" : "1px solid var(--border)",
        }}>
          {isSet ? "set" : "not set"}
        </span>
      </div>
      <input
        type="password"
        className="form-input"
        style={{ fontSize: 11, padding: "6px 9px" }}
        placeholder={isSet ? "Enter new value to replace…" : placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        autoComplete="off"
      />
    </div>
  );
}
