import * as React from "react";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import { Keyboard, X } from "lucide-react";
import { SHORTCUT_REGISTRY } from "@/hooks/use-keyboard-shortcuts";
import { cn } from "@/lib/utils";

interface ShortcutHelpDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ShortcutHelpDialog({
  open,
  onOpenChange,
}: ShortcutHelpDialogProps) {
  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/50 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0" />
        <DialogPrimitive.Content
          className={cn(
            "fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2",
            "rounded-lg border bg-card p-6 shadow-lg",
            "data-[state=open]:animate-in data-[state=closed]:animate-out",
            "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
            "data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95",
            "data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%]",
            "data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]"
          )}
        >
          <div className="flex items-center gap-2 mb-4">
            <Keyboard size={18} className="text-muted-foreground" />
            <DialogPrimitive.Title className="text-lg font-semibold">
              Keyboard Shortcuts
            </DialogPrimitive.Title>
          </div>

          <DialogPrimitive.Description className="sr-only">
            A list of all available keyboard shortcuts in the application.
          </DialogPrimitive.Description>

          <div className="space-y-1">
            {SHORTCUT_REGISTRY.map((shortcut) => (
              <div
                key={shortcut.key}
                className="flex items-center justify-between rounded-md px-2 py-2 text-sm hover:bg-accent/50"
              >
                <span className="text-foreground">{shortcut.description}</span>
                <kbd className="ml-4 inline-flex h-6 items-center gap-1 rounded border border-border bg-muted px-2 font-mono text-xs text-muted-foreground">
                  {shortcut.key}
                </kbd>
              </div>
            ))}
          </div>

          <p className="mt-4 text-xs text-muted-foreground">
            Press <kbd className="rounded border border-border bg-muted px-1 font-mono">?</kbd> to toggle this dialog.
          </p>

          <DialogPrimitive.Close className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2">
            <X className="h-4 w-4" />
            <span className="sr-only">Close</span>
          </DialogPrimitive.Close>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}
