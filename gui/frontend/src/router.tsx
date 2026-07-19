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
          const mod = await import("./pages/Home");
          return { Component: mod.Home };
        },
      },
      {
        path: "projects",
        lazy: async () => {
          const mod = await import("./pages/Repositories");
          return { Component: mod.Repositories };
        },
      },
      {
        path: "projects/:projectId",
        lazy: async () => {
          const mod = await import("./pages/projects/ProjectOverview");
          return { Component: mod.ProjectOverview };
        },
      },
      {
        path: "projects/:projectId/skills",
        lazy: async () => {
          const mod = await import("./pages/projects/ProjectSkillsPage");
          return { Component: mod.ProjectSkillsPage };
        },
      },
      {
        path: "projects/:projectId/agents",
        lazy: async () => {
          const mod = await import("./pages/projects/ProjectAgentsPage");
          return { Component: mod.ProjectAgentsPage };
        },
      },
      {
        path: "projects/:projectId/health",
        lazy: async () => {
          const mod = await import("./pages/projects/ProjectHealthPage");
          return { Component: mod.ProjectHealthPage };
        },
      },
      {
        path: "projects/:projectId/activity",
        lazy: async () => {
          const mod = await import("./pages/projects/ProjectActivityPage");
          return { Component: mod.ProjectActivityPage };
        },
      },
      {
        path: "catalog",
        lazy: async () => {
          const mod = await import("./pages/Catalog");
          return { Component: mod.Catalog };
        },
      },
      {
        path: "activity",
        lazy: async () => {
          const mod = await import("./pages/Activity");
          return { Component: mod.Activity };
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
        path: "repositories",
        element: <Navigate to="/projects" replace />,
      },
      {
        path: "registry",
        element: <Navigate to="/catalog" replace />,
      },
      {
        path: "audit",
        element: <Navigate to="/activity" replace />,
      },
      {
        path: "skills",
        element: <Navigate to="/projects" replace />,
      },
      {
        path: "skills/:skillName",
        lazy: async () => {
          const mod = await import("./pages/SkillDetail");
          return { Component: mod.SkillDetail };
        },
      },
      {
        path: "sync",
        element: <Navigate to="/projects" replace />,
      },
      {
        path: "doctor",
        element: <Navigate to="/projects" replace />,
      },
      {
        path: "cache",
        element: <Navigate to="/settings" replace />,
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
