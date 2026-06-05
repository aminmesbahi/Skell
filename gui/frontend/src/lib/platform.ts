// Platform / host environment helpers. Kept tiny to avoid pulling in heavy deps.

export const isMac =
  typeof navigator !== "undefined" &&
  /Mac|iPhone|iPad|iPod/i.test(navigator.platform || navigator.userAgent || "");