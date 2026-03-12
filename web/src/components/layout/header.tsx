import { LogOut, Menu } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { Breadcrumbs } from "./breadcrumbs";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { SidebarContent } from "./sidebar";
import { useState } from "react";

export function Header() {
  const { logout } = useAuth();
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  return (
    <header className="flex h-14 items-center justify-between border-b bg-card px-4">
      <div className="flex items-center gap-2">
        {/* Hamburger — visible only on mobile */}
        <Sheet open={mobileNavOpen} onOpenChange={setMobileNavOpen}>
          <Button
            variant="ghost"
            size="icon"
            className="md:hidden touch-manipulation"
            onClick={() => setMobileNavOpen(true)}
            aria-label="Open navigation"
          >
            <Menu className="h-5 w-5" />
          </Button>
          <SheetContent side="left" className="w-64 p-0">
            <SheetHeader className="flex h-14 items-center border-b px-4 py-0 space-y-0">
              <SheetTitle className="text-sm font-semibold">
                IBS Doc Engine
              </SheetTitle>
            </SheetHeader>
            <SidebarContent onNavigate={() => setMobileNavOpen(false)} />
          </SheetContent>
        </Sheet>

        <Breadcrumbs />
      </div>

      <div className="flex items-center gap-1">
        <ThemeToggle />
        <Button
          variant="ghost"
          size="sm"
          onClick={logout}
          className="touch-manipulation"
        >
          <LogOut size={16} className="mr-2" />
          Sign out
        </Button>
      </div>
    </header>
  );
}
