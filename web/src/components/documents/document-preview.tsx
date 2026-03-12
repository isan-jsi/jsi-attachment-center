import { Download, Trash2 } from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import type { Document } from "@/types/document";
import { formatBytes, formatDate } from "@/lib/utils";
import { getDownloadUrl, useDeleteDocument } from "@/hooks/use-documents";

interface DocumentPreviewProps {
  document: Document | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function isPreviewable(mime: string): boolean {
  return (
    mime.startsWith("image/") ||
    mime === "application/pdf" ||
    mime.startsWith("text/")
  );
}

export function DocumentPreview({
  document: doc,
  open,
  onOpenChange,
}: DocumentPreviewProps) {
  const deleteMutation = useDeleteDocument();

  if (!doc) return null;

  const downloadUrl = getDownloadUrl(doc.id);

  const handleDelete = async () => {
    if (!confirm(`Delete "${doc.file_name}"?`)) return;
    await deleteMutation.mutateAsync(doc.id);
    onOpenChange(false);
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-[440px] sm:w-[540px]">
        <SheetHeader>
          <SheetTitle className="pr-8 truncate">{doc.file_name}</SheetTitle>
          <SheetDescription>{doc.title}</SheetDescription>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Preview area */}
          {isPreviewable(doc.mime_type) && (
            <div className="rounded-md border bg-muted/30 p-2">
              {doc.mime_type.startsWith("image/") ? (
                <img
                  src={downloadUrl}
                  alt={doc.file_name}
                  className="max-h-64 w-full object-contain"
                  loading="lazy"
                />
              ) : doc.mime_type === "application/pdf" ? (
                <iframe
                  src={downloadUrl}
                  title={doc.file_name}
                  className="h-64 w-full"
                />
              ) : (
                <p className="text-xs text-muted-foreground">
                  Text preview not available inline.
                </p>
              )}
            </div>
          )}

          {/* Metadata */}
          <div className="space-y-3">
            <MetaRow label="Type" value={doc.document_type} />
            <MetaRow label="Extension" value={doc.file_extension} />
            <MetaRow label="MIME Type" value={doc.mime_type} />
            <MetaRow label="Size" value={formatBytes(doc.file_size)} />
            <MetaRow label="Version" value={String(doc.version)} />
            <MetaRow
              label="Owner"
              value={`${doc.owner_class_library}.${doc.owner_class_name}`}
            />
            <MetaRow label="Folder" value={doc.folder_name ?? "—"} />
            <MetaRow label="Created" value={formatDate(doc.created_at)} />
            <MetaRow label="Updated" value={formatDate(doc.updated_at)} />
            <MetaRow label="Checksum" value={doc.checksum} />
            <MetaRow
              label="Active"
              value={doc.is_active ? "Yes" : "No"}
            />
          </div>

          {/* Actions */}
          <div className="flex gap-2 pt-2">
            <Button asChild className="flex-1">
              <a href={downloadUrl} download>
                <Download className="mr-2 h-4 w-4" />
                Download
              </a>
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleteMutation.isPending}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}

function MetaRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className="text-right font-medium truncate max-w-[60%]" title={value}>
        {value}
      </span>
    </div>
  );
}
