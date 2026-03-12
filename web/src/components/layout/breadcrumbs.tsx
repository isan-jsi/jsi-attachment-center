import { Link, useMatches } from "@tanstack/react-router";
import { ChevronRight } from "lucide-react";

export function Breadcrumbs() {
  const matches = useMatches();

  const crumbs = matches
    .filter((m) => m.pathname !== "/" && m.pathname !== "")
    .map((m) => ({
      label: m.pathname.split("/").filter(Boolean).pop() || "Home",
      path: m.pathname,
    }));

  return (
    <nav className="flex items-center gap-1 text-sm text-muted-foreground">
      <Link to="/documents" className="hover:text-foreground">
        Home
      </Link>
      {crumbs.map((crumb, i) => (
        <span key={crumb.path} className="flex items-center gap-1">
          <ChevronRight size={14} />
          {i === crumbs.length - 1 ? (
            <span className="text-foreground capitalize">{crumb.label}</span>
          ) : (
            <Link to={crumb.path} className="capitalize hover:text-foreground">
              {crumb.label}
            </Link>
          )}
        </span>
      ))}
    </nav>
  );
}
