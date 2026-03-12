import { useState, useEffect } from "react";
import { Outlet } from "@tanstack/react-router";
import { Sidebar } from "./sidebar";
import { Header } from "./header";
import { useGlobalShortcuts } from "@/hooks/use-keyboard-shortcuts";
import { ShortcutHelpDialog } from "@/components/ui/shortcut-help-dialog";

export function AppShell() {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [shortcutHelpOpen, setShortcutHelpOpen] = useState(false);

  useGlobalShortcuts();

  useEffect(() => {
    const handler = () => setShortcutHelpOpen((prev) => !prev);
    document.addEventListener("shortcut:show-help", handler);
    return () => document.removeEventListener("shortcut:show-help", handler);
  }, []);

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <Sidebar
        collapsed={sidebarCollapsed}
        onToggle={() => setSidebarCollapsed((prev) => !prev)}
      />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header />
        <main className="flex-1 overflow-y-auto p-6">
          <Outlet />
        </main>
      </div>
      <ShortcutHelpDialog
        open={shortcutHelpOpen}
        onOpenChange={setShortcutHelpOpen}
      />
    </div>
  );
}
