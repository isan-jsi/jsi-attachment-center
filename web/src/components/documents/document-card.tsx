import { FileText, FileImage, FileSpreadsheet, File } from "lucide-react";
import type { Document } from "@/types/document";
import { formatBytes, formatDate, cn } from "@/lib/utils";

interface DocumentCardProps {
  document: Document;
  selected?: boolean;
  onSelect?: () => void;
  onClick?: () => void;
}

function getFileIcon(ext: string) {
  const lower = ext.toLowerCase().replace(".", "");
  if (["pdf", "doc", "docx", "txt", "rtf"].includes(lower))
    return FileText;
  if (["jpg", "jpeg", "png", "gif", "bmp", "svg", "webp"].includes(lower))
    return FileImage;
  if (["xls", "xlsx", "csv"].includes(lower))
    return FileSpreadsheet;
  return File;
}

function getFileColor(ext: string): string {
  const lower = ext.toLowerCase().replace(".", "");
  if (["pdf"].includes(lower)) return "text-red-500";
  if (["doc", "docx"].includes(lower)) return "text-blue-500";
  if (["xls", "xlsx", "csv"].includes(lower)) return "text-green-500";
  if (["jpg", "jpeg", "png", "gif", "bmp", "svg", "webp"].includes(lower))
    return "text-purple-500";
  return "text-muted-foreground";
}

export function DocumentCard({
  document,
  selected,
  onSelect,
  onClick,
}: DocumentCardProps) {
  const Icon = getFileIcon(document.file_extension);
  const color = getFileColor(document.file_extension);

  return (
    <div
      className={cn(
        "group relative flex flex-col items-center gap-2 rounded-lg border p-4 transition-colors hover:bg-accent cursor-pointer",
        selected && "border-primary bg-accent"
      )}
      onClick={onClick}
    >
      {/* Selection checkbox area */}
      {onSelect && (
        <input
          type="checkbox"
          checked={selected}
          onChange={(e) => {
            e.stopPropagation();
            onSelect();
          }}
          onClick={(e) => e.stopPropagation()}
          className="absolute left-2 top-2 h-4 w-4 opacity-0 group-hover:opacity-100 data-[checked]:opacity-100"
          data-checked={selected || undefined}
        />
      )}

      <Icon className={cn("h-10 w-10", color)} />

      <div className="w-full text-center">
        <p className="truncate text-sm font-medium" title={document.file_name}>
          {document.file_name}
        </p>
        <p className="text-xs text-muted-foreground">
          {formatBytes(document.file_size)} &middot;{" "}
          {formatDate(document.created_at)}
        </p>
      </div>
    </div>
  );
}
