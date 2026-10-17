import { isRouteErrorResponse, useRouteError } from "react-router";

/**
 * Shown when a page throws while rendering or a lazily loaded page fails to
 * load. Without it the window would go blank with no way back.
 */
export function RouteError() {
  const error = useRouteError();
  const detail = isRouteErrorResponse(error)
    ? `${error.status} ${error.statusText}`
    : error instanceof Error
      ? error.message
      : String(error);
  return (
    <div role="alert" className="mx-auto mt-16 max-w-xl card space-y-3">
      <h1 className="text-lg font-semibold text-slate-100">Something went wrong</h1>
      <p className="text-sm text-slate-400">
        This page hit an unexpected error. Your projects and skills are not affected.
      </p>
      <details className="text-xs text-slate-500">
        <summary className="cursor-pointer">Technical details</summary>
        <pre className="mt-2 whitespace-pre-wrap break-words">{detail}</pre>
      </details>
      <div className="flex gap-2">
        <button className="btn-primary" onClick={() => window.location.reload()}>Reload</button>
        <a className="btn-ghost" href="/">Go to Home</a>
      </div>
    </div>
  );
}
