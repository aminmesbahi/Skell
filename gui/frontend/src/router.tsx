import { createBrowserRouter, RouterProvider, Navigate } from "react-router-dom";
import { Layout } from "./components/Layout";
import { Dashboard } from "./pages/Dashboard";
import { Repositories } from "./pages/Repositories";
import { InstalledSkills } from "./pages/InstalledSkills";
import { Registry } from "./pages/Registry";
import { SkillDetail } from "./pages/SkillDetail";
import { Sync } from "./pages/Sync";
import { Doctor } from "./pages/Doctor";
import { Cache } from "./pages/Cache";
import { Settings } from "./pages/Settings";
import { AuditLog } from "./pages/AuditLog";
import { ContributeMetadataPage } from "./pages/ContributeMetadata";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Layout />,
    children: [
      { index: true, element: <Dashboard /> },
      { path: "repositories", element: <Repositories /> },
      { path: "skills", element: <InstalledSkills /> },
      { path: "skills/:skillName", element: <SkillDetail /> },
      { path: "registry", element: <Registry /> },
      { path: "sync", element: <Sync /> },
      { path: "doctor", element: <Doctor /> },
      { path: "cache", element: <Cache /> },
      { path: "audit", element: <AuditLog /> },
      { path: "settings", element: <Settings /> },
      { path: "contribute/:skillName", element: <ContributeMetadataPage /> },
      { path: "*", element: <Navigate to="/" replace /> },
    ],
  },
]);

export function AppRouter() {
  return <RouterProvider router={router} />;
}
