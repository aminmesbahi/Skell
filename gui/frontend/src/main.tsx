import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App";
import "./index.css";

// Tag the document with the host OS so platform-specific styles (e.g. macOS
// traffic-light spacing in the sidebar) can be applied via CSS.
if (typeof navigator !== "undefined" && /Mac|iPhone|iPad/.test(navigator.userAgent)) {
  document.documentElement.classList.add("is-mac");
}

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
