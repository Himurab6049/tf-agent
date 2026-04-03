import { useEffect, useState } from "react";
import { getSettings, saveSettings, updateMe } from "../lib/api";
import { AdminPage } from "./AdminPage";
import type { UserInfo, UserSettings } from "../types";

type Section = "profile" | "github" | "jira" | "users";

interface Props {
  userInfo: UserInfo | null;
  onUserUpdate: (info: UserInfo) => void;
}

const SS_SECTION = "tf_settings_section";

export function SettingsPage({ userInfo, onUserUpdate }: Props) {
  const [section, setSection] = useState<Section>(
    () => (sessionStorage.getItem(SS_SECTION) as Section) ?? "profile"
  );
  const [settings, setSettings] = useState<UserSettings | null>(null);

  const navSection = (s: Section) => {
    setSection(s);
    sessionStorage.setItem(SS_SECTION, s);
  };

  useEffect(() => {
    getSettings().then(setSettings).catch(() => {});
  }, []);

  const reload = () => getSettings().then(setSettings).catch(() => {});

  return (
    <div>
      {/* Page header */}
      <div style={{ marginBottom: 28 }}>
        <div style={{ fontSize: 22, fontWeight: 700, color: "var(--text)", letterSpacing: "-.02em" }}>
          Settings
        </div>
      </div>

      <div style={{ display: "flex", gap: 32, alignItems: "flex-start" }}>

        {/* Left nav */}
        <div style={{ width: 180, flexShrink: 0 }}>
          <SectionNav label="Profile" active={section === "profile"} onClick={() => navSection("profile")} />
          <SectionNav label="GitHub Key" active={section === "github"} onClick={() => navSection("github")} configured={settings?.github_token_set} />
          <SectionNav label="Jira Key" active={section === "jira"} onClick={() => navSection("jira")} configured={settings?.atlassian_token_set} />
          {userInfo?.role === "admin" && (
            <>
              <div style={{ height: 1, background: "var(--border)", margin: "10px 0" }} />
              <SectionNav label="Users" active={section === "users"} onClick={() => navSection("users")} />
            </>
          )}
        </div>

        {/* Right content */}
        <div style={{ flex: 1, minWidth: 0 }}>
          {section === "profile" && (
            <ProfileSection userInfo={userInfo} onUserUpdate={onUserUpdate} />
          )}
          {section === "github" && (
            <GitHubSection isSet={settings?.github_token_set ?? false} onSaved={reload} />
          )}
          {section === "jira" && (
            <JiraSection
              isSet={settings?.atlassian_token_set ?? false}
              domain={settings?.atlassian_domain ?? ""}
              email={settings?.atlassian_email ?? ""}
              onSaved={reload}
            />
          )}
          {section === "users" && <AdminPage />}
        </div>

      </div>
    </div>
  );
}

// ── Left nav item ──────────────────────────────────────────────────────────────

function SectionNav({ label, active, onClick, configured }: {
  label: string;
  active: boolean;
  onClick: () => void;
  configured?: boolean;
}) {
  return (
    <button
      onClick={onClick}
      style={{
        display: "flex", alignItems: "center", justifyContent: "space-between",
        width: "100%", padding: "8px 12px", borderRadius: 7, marginBottom: 2,
        background: active ? "var(--surface-2, #f5f4f0)" : "transparent",
        border: "none", cursor: "pointer", textAlign: "left",
        fontSize: "var(--text-sm)", fontWeight: active ? 600 : 400,
        color: active ? "var(--text)" : "var(--text-2)",
        transition: "background .12s, color .12s",
      }}
      onMouseEnter={e => { if (!active) e.currentTarget.style.background = "var(--surface-2, #f5f4f0)"; }}
      onMouseLeave={e => { if (!active) e.currentTarget.style.background = "transparent"; }}
    >
      {label}
      {configured === false && (
        <span style={{
          width: 7, height: 7, borderRadius: "50%", flexShrink: 0,
          background: "#d0d0d0",
        }} />
      )}
    </button>
  );
}

// ── Profile section ────────────────────────────────────────────────────────────

function ProfileSection({ userInfo, onUserUpdate }: { userInfo: UserInfo | null; onUserUpdate: (info: UserInfo) => void }) {
  const [username, setUsername] = useState(userInfo?.username ?? "");
  const [saving, setSaving] = useState(false);
  const [feedback, setFeedback] = useState<{ ok: boolean; msg: string } | null>(null);

  // Sync input when userInfo loads or changes (e.g. first load, after save)
  useEffect(() => {
    if (userInfo?.username) setUsername(userInfo.username);
  }, [userInfo?.username]);

  const handleSave = async () => {
    const trimmed = username.trim();
    if (!trimmed || trimmed === userInfo?.username) return;
    setSaving(true);
    setFeedback(null);
    try {
      const updated = await updateMe(trimmed);
      onUserUpdate(updated);
      setFeedback({ ok: true, msg: "Display name updated." });
      setTimeout(() => setFeedback(null), 3000);
    } catch (e) {
      setFeedback({ ok: false, msg: e instanceof Error ? e.message : "Failed to update." });
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>
      <SectionHeader
        title="Profile"
        description="Update your display name shown in the sidebar and team views."
      />

      <div className="card" style={{ marginTop: 20 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 14, marginBottom: 20, padding: "12px 16px", background: "var(--surface-2)", borderRadius: 8 }}>
          <div style={{
            width: 44, height: 44, borderRadius: "50%", flexShrink: 0,
            background: "var(--accent)", color: "#fff",
            display: "flex", alignItems: "center", justifyContent: "center",
            fontSize: 18, fontWeight: 700,
          }}>
            {(username || userInfo?.username || "U")[0].toUpperCase()}
          </div>
          <div>
            <div style={{ fontSize: "var(--text-base)", fontWeight: 600, color: "var(--text)" }}>
              {username || userInfo?.username || "—"}
            </div>
            <div style={{ fontSize: "var(--text-xs)", color: "var(--text-3)", marginTop: 2 }}>
              {userInfo?.role ?? ""}
            </div>
          </div>
        </div>

        <div className="form-row" style={{ marginBottom: 16 }}>
          <label className="form-label">Display name</label>
          <input
            type="text"
            className="form-input"
            placeholder="Your name"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") handleSave(); }}
            autoComplete="off"
          />
          <span className="form-hint">This is how you appear in the sidebar and task history.</span>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <button
            className="btn btn-primary"
            onClick={handleSave}
            disabled={saving || !username.trim() || username.trim() === userInfo?.username}
          >
            {saving ? "Saving…" : "Save"}
          </button>
          {feedback && (
            <span style={{ fontSize: "var(--text-sm)", color: feedback.ok ? "var(--green)" : "var(--red)" }}>
              {feedback.msg}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

// ── GitHub section ─────────────────────────────────────────────────────────────

function GitHubSection({ isSet, onSaved }: { isSet: boolean; onSaved: () => void }) {
  const [token, setToken] = useState("");
  const [saving, setSaving] = useState(false);
  const [feedback, setFeedback] = useState<{ ok: boolean; msg: string } | null>(null);

  const handleSave = async () => {
    if (!token.trim()) return;
    setSaving(true);
    setFeedback(null);
    try {
      await saveSettings({ github_token: token.trim() });
      setToken("");
      onSaved();
      setFeedback({ ok: true, msg: "GitHub key saved." });
      setTimeout(() => setFeedback(null), 3000);
    } catch (e) {
      setFeedback({ ok: false, msg: e instanceof Error ? e.message : "Failed to save." });
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>
      <SectionHeader
        title="GitHub Key"
        description="Personal access token used to open pull requests on your behalf. Requires the repo scope."
      />

      <div className="card" style={{ marginTop: 20 }}>
        <div style={{ marginBottom: 16, fontSize: "var(--text-sm)", color: isSet ? "#2d6a4f" : "var(--text-3)" }}>
          {isSet ? "Token configured — enter a new value below to replace it." : "No token set."}
        </div>
        <div className="form-row" style={{ marginBottom: 16 }}>
          <label className="form-label">Personal Access Token</label>
          <input
            type="password"
            className="form-input"
            placeholder="ghp_…"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") handleSave(); }}
            autoComplete="off"
          />
          <span className="form-hint">
            Generate one at <strong>github.com → Settings → Developer settings → Personal access tokens</strong>
          </span>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <button
            className="btn btn-primary"
            onClick={handleSave}
            disabled={saving || !token.trim()}
          >
            {saving ? "Saving…" : "Save"}
          </button>
          {feedback && (
            <span style={{ fontSize: "var(--text-sm)", color: feedback.ok ? "var(--green)" : "var(--red)" }}>
              {feedback.msg}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

// ── Jira section ───────────────────────────────────────────────────────────────

function JiraSection({ isSet, domain: savedDomain, email: savedEmail, onSaved }: {
  isSet: boolean;
  domain: string;
  email: string;
  onSaved: () => void;
}) {
  const [token, setToken] = useState("");
  const [domain, setDomain] = useState(savedDomain);
  const [email, setEmail] = useState(savedEmail);
  const [saving, setSaving] = useState(false);
  const [feedback, setFeedback] = useState<{ ok: boolean; msg: string } | null>(null);

  // Sync if parent reloads saved values
  useEffect(() => { setDomain(savedDomain); }, [savedDomain]);
  useEffect(() => { setEmail(savedEmail); }, [savedEmail]);

  const handleSave = async () => {
    setSaving(true);
    setFeedback(null);
    try {
      await saveSettings({
        atlassian_token: token.trim() || undefined,
        atlassian_domain: domain.trim() || undefined,
        atlassian_email: email.trim() || undefined,
      });
      setToken("");
      onSaved();
      setFeedback({ ok: true, msg: "Jira settings saved." });
      setTimeout(() => setFeedback(null), 3000);
    } catch (e) {
      setFeedback({ ok: false, msg: e instanceof Error ? e.message : "Failed to save." });
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>
      <SectionHeader
        title="Jira Key"
        description="Atlassian API token used to fetch Jira ticket context when running tasks with a ticket number."
      />

      <div className="card" style={{ marginTop: 20 }}>
        <div style={{ marginBottom: 16, fontSize: "var(--text-sm)", color: isSet ? "#2d6a4f" : "var(--text-3)" }}>
          {isSet ? "Token configured — enter a new value below to replace it." : "No token set."}
        </div>
        <div className="form-row" style={{ marginBottom: 16 }}>
          <label className="form-label">Atlassian API Token</label>
          <input
            type="password"
            className="form-input"
            placeholder="ATATT3x…"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            autoComplete="off"
          />
          <span className="form-hint">
            Generate one at <strong>id.atlassian.com → Security → API tokens</strong>
          </span>
        </div>

        <div style={{ display: "flex", gap: 12, marginBottom: 16 }}>
          <div className="form-row" style={{ flex: 1, marginBottom: 0 }}>
            <label className="form-label">Domain</label>
            <input
              type="text"
              className="form-input"
              placeholder="mycompany.atlassian.net"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
            />
          </div>
          <div className="form-row" style={{ flex: 1, marginBottom: 0 }}>
            <label className="form-label">Email</label>
            <input
              type="email"
              className="form-input"
              placeholder="you@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <button
            className="btn btn-primary"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? "Saving…" : "Save"}
          </button>
          {feedback && (
            <span style={{ fontSize: "var(--text-sm)", color: feedback.ok ? "var(--green)" : "var(--red)" }}>
              {feedback.msg}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

// ── Shared section header ──────────────────────────────────────────────────────

function SectionHeader({ title, description }: { title: string; description: string }) {
  return (
    <div>
      <h2 style={{ fontSize: 18, fontWeight: 700, color: "var(--text)", letterSpacing: "-.01em", margin: "0 0 6px" }}>
        {title}
      </h2>
      <p style={{ fontSize: "var(--text-sm)", color: "var(--text-3)", margin: 0, lineHeight: 1.6 }}>
        {description}
      </p>
    </div>
  );
}
