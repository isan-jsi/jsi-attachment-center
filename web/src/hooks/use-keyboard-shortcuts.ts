import { useHotkeys } from "react-hotkeys-hook";
import { useNavigate } from "@tanstack/react-router";

export interface ShortcutDef {
  key: string;
  description: string;
  scope?: string;
}

export function useGlobalShortcuts() {
  const navigate = useNavigate();

  useHotkeys(
    "mod+k",
    (e) => {
      e.preventDefault();
      const searchInput =
        document.querySelector<HTMLInputElement>("[data-search-input]");
      searchInput?.focus();
    },
    { enableOnFormTags: false }
  );

  useHotkeys(
    "mod+n",
    (e) => {
      e.preventDefault();
      void navigate({ to: "/documents" });
    },
    { enableOnFormTags: false }
  );

  useHotkeys("escape", () => {
    const event = new CustomEvent("shortcut:escape");
    document.dispatchEvent(event);
  });

  useHotkeys(
    "shift+/",
    (e) => {
      e.preventDefault();
      const event = new CustomEvent("shortcut:show-help");
      document.dispatchEvent(event);
    },
    { enableOnFormTags: false }
  );
}

export const SHORTCUT_REGISTRY: ShortcutDef[] = [
  { key: "Ctrl+K", description: "Focus search", scope: "global" },
  { key: "Ctrl+N", description: "Go to documents", scope: "global" },
  { key: "Escape", description: "Close panel / clear selection", scope: "global" },
  { key: "?", description: "Show keyboard shortcuts", scope: "global" },
];
