import {
  createRouter,
  createRootRoute,
  createRoute,
} from "@tanstack/react-router";
import { AppShell } from "@/components/layout/app-shell";
import LoginPage from "@/pages/login";
import { useAuth } from "@/context/auth-context";
import DocumentsPage from "@/pages/documents";
import SearchPage from "@/pages/search";
import SyncDashboard from "@/pages/sync-dashboard";

// Placeholder pages
function SettingsPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold">Settings</h1>
      <p className="text-muted-foreground mt-2">Settings coming soon.</p>
    </div>
  );
}

function RootComponent() {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) return <LoginPage />;
  return <AppShell />;
}

const rootRoute = createRootRoute({
  component: RootComponent,
});

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: DocumentsPage,
});

const documentsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/documents",
  component: DocumentsPage,
});

const documentsFolderRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/documents/$folderId",
  component: DocumentsPage,
});

const searchRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/search",
  component: SearchPage,
});

const syncRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/sync",
  component: SyncDashboard,
});

const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings",
  component: SettingsPage,
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  documentsRoute,
  documentsFolderRoute,
  searchRoute,
  syncRoute,
  settingsRoute,
]);

export const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
