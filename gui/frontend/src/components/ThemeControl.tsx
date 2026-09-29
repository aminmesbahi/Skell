import { useEffect, useState } from "react";
export function ThemeControl() {
  const [theme, setTheme] = useState(() => localStorage.getItem("skell-theme") || "system");
  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const apply = () => { document.documentElement.dataset.theme = theme === "system" ? media.matches ? "dark" : "light" : theme; };
    apply(); localStorage.setItem("skell-theme", theme);
    media.addEventListener("change", apply);
    return () => media.removeEventListener("change", apply);
  }, [theme]);
  return (
    <label className="flex items-center gap-2 text-xs text-slate-400">
      <span className="font-medium uppercase tracking-wide text-slate-500">Theme</span>
      <select
        aria-label="Theme"
        className="input h-9 w-auto min-w-28 px-2"
        value={theme}
        onChange={(e) => setTheme(e.target.value)}
      >
        <option value="system">System</option>
        <option value="light">Light</option>
        <option value="dark">Dark</option>
      </select>
    </label>
  );
}
