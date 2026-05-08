import { createBrowserRouter, RouterProvider, Navigate } from "react-router-dom";
import { Layout } from "./components/Layout";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Layout />,
    children: [
      {
        index: true,
        lazy: async () => {
          const mod = await import("./pages/Dashboard");
          return { Component: mod.Dashboard };
        },
      },
      {
        path: "repositories",
        lazy: async () => {
          const mod = await import("./pages/Repositories");
          return { Component: mod.Repositories };
        },
      },
      {
        path: "skills",
        lazy: async () => {
          const mod = await import("./pages/InstalledSkills");
          return { Component: mod.InstalledSkills };
        },
      },
      {
        path: "skills/:skillName",
        lazy: async () => {
          const mod = await import("./pages/SkillDetail");
          return { Component: mod.SkillDetail };
        },
      },
      {
        path: "registry",
        lazy: async () => {
          const mod = await import("./pages/Registry");
          return { Component: mod.Registry };
        },
      },
      {
        path: "sync",
        lazy: async () => {
          const mod = await import("./pages/Sync");
          return { Component: mod.Sync };
        },
      },
      {
        path: "doctor",
        lazy: async () => {
          const mod = await import("./pages/Doctor");
          return { Component: mod.Doctor };
        },
      },
      {
        path: "cache",
        lazy: async () => {
          const mod = await import("./pages/Cache");
          return { Component: mod.Cache };
        },
      },
      {
        path: "audit",
        lazy: async () => {
          const mod = await import("./pages/AuditLog");
          return { Component: mod.AuditLog };
        },
      },
      {
        path: "settings",
        lazy: async () => {
          const mod = await import("./pages/Settings");
          return { Component: mod.Settings };
        },
      },
      {
        path: "contribute-info",
        lazy: async () => {
          const mod = await import("./pages/ContributeInfo");
          return { Component: mod.ContributeInfo };
        },
      },
      {
        path: "contribute/:skillName",
        lazy: async () => {
          const mod = await import("./pages/ContributeMetadata");
          return { Component: mod.ContributeMetadataPage };
        },
      },
      { path: "*", element: <Navigate to="/" replace /> },
    ],
  },
]);

export function AppRouter() {
  return <RouterProvider router={router} />;
}
