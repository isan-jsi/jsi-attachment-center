import { LogOut } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { Button } from "@/components/ui/button";
import { Breadcrumbs } from "./breadcrumbs";

export function Header() {
  const { logout } = useAuth();

  return (
    <header className="flex h-14 items-center justify-between border-b bg-card px-4">
      <Breadcrumbs />
      <Button variant="ghost" size="sm" onClick={logout}>
        <LogOut size={16} className="mr-2" />
        Sign out
      </Button>
    </header>
  );
}
