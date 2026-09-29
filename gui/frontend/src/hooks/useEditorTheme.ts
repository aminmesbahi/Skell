import { useEffect, useState } from "react";
export function useEditorTheme() {
  const read = () => document.documentElement.dataset.theme === "light" ? "vs" : "vs-dark";
  const [theme, setTheme] = useState(read);
  useEffect(() => {
    const observer = new MutationObserver(() => setTheme(read()));
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });
    return () => observer.disconnect();
  }, []);
  return theme;
}
