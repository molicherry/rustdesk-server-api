import { useState } from "react";
import { Save } from "lucide-react";
import { useTranslation } from "react-i18next";

function Card({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-white rounded-lg p-5 mb-4" style={{ border: "1px solid #E2E8F0" }}>
      <h3 className="text-sm font-semibold text-slate-800 mb-4" style={{ fontFamily: "Inter, system-ui, sans-serif" }}>
        {title}
      </h3>
      {children}
    </div>
  );
}

function Toggle({ label, checked, onChange }: { label: string; checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="flex items-center justify-between py-2">
      <span className="text-sm text-slate-700">{label}</span>
      <button
        type="button"
        onClick={() => onChange(!checked)}
        className="relative inline-flex h-5 w-9 items-center rounded-full transition-colors"
        style={{ backgroundColor: checked ? "#E95B24" : "#cbd5e1" }}
      >
        <span
          className="inline-block h-4 w-4 rounded-full bg-white transition-transform"
          style={{ transform: checked ? "translateX(18px)" : "translateX(2px)" }}
        />
      </button>
    </label>
  );
}

function InputRow({ label, value, onChange, type = "text" }: { label: string; value: string; onChange: (v: string) => void; type?: string }) {
  return (
    <div className="py-2">
      <label className="block text-sm text-slate-700 mb-1">{label}</label>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full rounded-md px-3 py-2 text-sm border focus:outline-none focus:ring-2"
        style={{ borderColor: "#E2E8F0" }}
      />
    </div>
  );
}

export default function SettingsPage() {
  const { t } = useTranslation("settings");
  const [twoFA, setTwoFA] = useState(false);
  const [emailVerify, setEmailVerify] = useState(false);
  const [jwtExpiry, setJwtExpiry] = useState("24");
  const [ldapUrl, setLdapUrl] = useState("");
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="p-6 max-w-3xl">
      <h2 className="text-lg font-semibold text-slate-800 mb-4" style={{ fontFamily: "Inter, system-ui, sans-serif" }}>
        {t("title")}
      </h2>

      <Card title={t("authentication")}>
        <Toggle label={t("enable2FA")} checked={twoFA} onChange={setTwoFA} />
        <Toggle label={t("requireEmailVerification")} checked={emailVerify} onChange={setEmailVerify} />
        <InputRow label={t("jwtExpiryHours")} value={jwtExpiry} onChange={setJwtExpiry} type="number" />
      </Card>

      <Card title={t("securityPolicies")}>
        <InputRow label={t("passwordMinLength")} value="8" onChange={() => {}} />
        <InputRow label={t("sessionTimeout")} value="30" onChange={() => {}} />
      </Card>

      <Card title={t("ldapDirectory")}>
        <InputRow label={t("ldapServerUrl")} value={ldapUrl} onChange={setLdapUrl} />
        <InputRow label={t("bindDn")} value="" onChange={() => {}} />
      </Card>

      <Card title={t("organizations")}>
        <InputRow label={t("defaultOrgName")} value="Default" onChange={() => {}} />
      </Card>

      <Card title={t("localization")}>
        <InputRow label={t("defaultLanguage")} value="en" onChange={() => {}} />
        <InputRow label={t("defaultTimezone")} value="UTC" onChange={() => {}} />
      </Card>

      <div className="flex items-center gap-3">
        <button
          onClick={handleSave}
          className="flex items-center gap-2 text-sm font-medium text-white px-4 py-2 rounded-md transition-colors"
          style={{ backgroundColor: "#E95B24" }}
        >
          <Save size={16} />
          {t("saveChanges")}
        </button>
        {saved && <span className="text-sm text-green-600">{t("saved")}</span>}
      </div>
    </div>
  );
}
