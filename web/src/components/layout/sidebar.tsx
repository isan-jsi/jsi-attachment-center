import { Link, useMatchRoute } from "@tanstack/react-router";
import {
  FileText,
  Search,
  RefreshCw,
  Settings,
  PanelLeftClose,
  PanelLeft,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { FolderTreeComponent } from "@/components/folders/folder-tree";

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
}

const navItems = [
  { to: "/documents", label: "Documents", icon: FileText },
  { to: "/search", label: "Search", icon: Search },
  { to: "/sync", label: "Sync Status", icon: RefreshCw },
  { to: "/settings", label: "Settings", icon: Settings },
] as const;

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const matchRoute = useMatchRoute();

  return (
    <aside
      className={cn(
        "flex h-full flex-col border-r bg-card transition-all duration-200",
        collapsed ? "w-14" : "w-64",
      )}
    >
      <div className="flex h-14 items-center justify-between border-b px-3">
        {!collapsed && (
          <span className="text-sm font-semibold">IBS Doc Engine</span>
        )}
        <Button variant="ghost" size="icon" onClick={onToggle}>
          {collapsed ? <PanelLeft size={18} /> : <PanelLeftClose size={18} />}
        </Button>
      </div>

      <nav className="flex-1 space-y-1 overflow-y-auto p-2">
        {navItems.map(({ to, label, icon: Icon }) => {
          const isActive = matchRoute({ to, fuzzy: true });
          return (
            <Link
              key={to}
              to={to}
              className={cn(
                "flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors",
                isActive
                  ? "bg-accent text-accent-foreground"
                  : "text-muted-foreground hover:bg-accent/50",
              )}
            >
              <Icon size={18} />
              {!collapsed && <span>{label}</span>}
            </Link>
          );
        })}

        {!collapsed && (
          <div className="mt-4 border-t pt-4">
            <p className="mb-2 px-3 text-xs font-medium uppercase text-muted-foreground">
              Folders
            </p>
            <FolderTreeComponent />
          </div>
        )}
      </nav>
    </aside>
  );
}
