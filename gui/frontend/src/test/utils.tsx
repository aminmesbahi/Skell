import { type ReactElement } from "react";
import { render, type RenderOptions } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import { vi } from "vitest";
import type { Mock } from "vitest";

// ---------------------------------------------------------------------------
// renderWithRouter — wraps the component in MemoryRouter for route-aware tests
// ---------------------------------------------------------------------------
interface RenderWithRouterOptions extends Omit<RenderOptions, "wrapper"> {
  initialEntries?: string[];
}

export function renderWithRouter(
  ui: ReactElement,
  { initialEntries = ["/"], ...opts }: RenderWithRouterOptions = {}
) {
  return render(ui, {
    wrapper: ({ children }) => (
      <MemoryRouter
        initialEntries={initialEntries}
        future={{ v7_startTransition: true, v7_relativeSplatPath: true }}
      >
        {children}
      </MemoryRouter>
    ),
    ...opts,
  });
}

// ---------------------------------------------------------------------------
// Helper to type-safely cast vi.fn() mocks
// ---------------------------------------------------------------------------
export function asMock<T>(fn: T): Mock {
  return fn as unknown as Mock;
}

// ---------------------------------------------------------------------------
// Flush all pending promises / microtasks
// ---------------------------------------------------------------------------
export async function flush() {
  await vi.runAllTimersAsync().catch(() => undefined);
  await new Promise((r) => setTimeout(r, 0));
}

// ---------------------------------------------------------------------------
// renderRoute — wraps component in Routes/Route so useParams() works
// Use this for components that read URL params via useParams().
// ---------------------------------------------------------------------------
export function renderRoute(
  routePath: string,
  element: ReactElement,
  initialEntry: string
) {
  return render(
    <MemoryRouter
      initialEntries={[initialEntry]}
      future={{ v7_startTransition: true, v7_relativeSplatPath: true }}
    >
      <Routes>
        <Route path={routePath} element={element} />
      </Routes>
    </MemoryRouter>
  );
}
