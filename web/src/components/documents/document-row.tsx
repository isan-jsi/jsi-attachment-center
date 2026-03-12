import { FileText, FileImage, FileSpreadsheet, File } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import type { Document } from "@/types/document";
import { formatBytes, formatDate, cn } from "@/lib/utils";

interface DocumentRowProps {
  document: Document;
  selected?: boolean;
  onSelect?: (selected: boolean) => void;
  onClick?: () => void;
}

function getFileIcon(ext: string) {
  const lower = ext.toLowerCase().replace(".", "");
  if (["pdf", "doc", "docx", "txt", "rtf"].includes(lower)) return FileText;
  if (["jpg", "jpeg", "png", "gif", "bmp", "svg", "webp"].includes(lower))
    return FileImage;
  if (["xls", "xlsx", "csv"].includes(lower)) return FileSpreadsheet;
  return File;
}

export function DocumentRow({
  document,
  selected,
  onSelect,
  onClick,
}: DocumentRowProps) {
  const Icon = getFileIcon(document.file_extension);

  return (
    <tr
      className={cn(
        "border-b transition-colors hover:bg-muted/50 cursor-pointer touch-manipulation",
        selected && "bg-accent",
      )}
      onClick={onClick}
    >
      <td className="w-10 px-3 py-2">
        <Checkbox
          checked={selected}
          onCheckedChange={(checked) => onSelect?.(checked)}
          onClick={(e) => e.stopPropagation()}
        />
      </td>
      <td className="px-3 py-2">
        <div className="flex items-center gap-2">
          <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span className="truncate text-sm font-medium">
            {document.file_name}
          </span>
        </div>
      </td>
      {/* Content Type — hidden on mobile */}
      <td className="hidden md:table-cell px-3 py-2 text-sm text-muted-foreground">
        {document.document_type}
      </td>
      {/* Size — hidden on mobile and tablet */}
      <td className="hidden lg:table-cell px-3 py-2 text-sm text-muted-foreground">
        {formatBytes(document.file_size)}
      </td>
      {/* Owner — hidden on mobile */}
      <td className="hidden md:table-cell px-3 py-2 text-sm text-muted-foreground">
        {document.owner_class_name}
      </td>
      {/* Created — hidden on smallest screens */}
      <td className="hidden sm:table-cell px-3 py-2 text-sm text-muted-foreground">
        {formatDate(document.created_at)}
      </td>
    </tr>
  );
}
