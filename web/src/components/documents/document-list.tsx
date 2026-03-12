import type { Document } from "@/types/document";
import { DocumentRow } from "./document-row";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";

interface DocumentListProps {
  documents: Document[];
  selectedIds: Set<number>;
  onToggleSelect: (id: number) => void;
  onSelectAll: (selected: boolean) => void;
  onPreview: (doc: Document) => void;
  sortBy: string;
  sortDir: "asc" | "desc";
  onSort: (column: string) => void;
}

function SortHeader({
  label,
  column,
  sortBy,
  sortDir,
  onSort,
  className,
}: {
  label: string;
  column: string;
  sortBy: string;
  sortDir: "asc" | "desc";
  onSort: (col: string) => void;
  className?: string;
}) {
  const isActive = sortBy === column;
  return (
    <th
      className={cn(
        "cursor-pointer px-3 py-2 text-left text-sm font-medium text-muted-foreground hover:text-foreground touch-manipulation",
        className,
      )}
      onClick={() => onSort(column)}
    >
      <span className="inline-flex items-center gap-1">
        {label}
        {isActive && (
          <span className="text-xs">{sortDir === "asc" ? "\u25B2" : "\u25BC"}</span>
        )}
      </span>
    </th>
  );
}

export function DocumentList({
  documents,
  selectedIds,
  onToggleSelect,
  onSelectAll,
  onPreview,
  sortBy,
  sortDir,
  onSort,
}: DocumentListProps) {
  if (documents.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <p className="text-lg font-medium">No documents found</p>
        <p className="text-sm">Upload a document or select a different folder.</p>
      </div>
    );
  }

  const allSelected =
    documents.length > 0 && documents.every((d) => selectedIds.has(d.id));

  return (
    <div className="overflow-auto rounded-md border">
      <table className="w-full">
        <thead>
          <tr className={cn("border-b bg-muted/50")}>
            <th className="w-10 px-3 py-2">
              <Checkbox
                checked={allSelected}
                onCheckedChange={onSelectAll}
              />
            </th>
            <SortHeader
              label="Name"
              column="file_name"
              sortBy={sortBy}
              sortDir={sortDir}
              onSort={onSort}
            />
            {/* Content Type hidden on mobile */}
            <SortHeader
              label="Type"
              column="document_type"
              sortBy={sortBy}
              sortDir={sortDir}
              onSort={onSort}
              className="hidden md:table-cell"
            />
            {/* Size hidden on mobile */}
            <SortHeader
              label="Size"
              column="file_size"
              sortBy={sortBy}
              sortDir={sortDir}
              onSort={onSort}
              className="hidden lg:table-cell"
            />
            {/* Owner hidden on mobile */}
            <SortHeader
              label="Owner"
              column="owner_class_name"
              sortBy={sortBy}
              sortDir={sortDir}
              onSort={onSort}
              className="hidden md:table-cell"
            />
            <SortHeader
              label="Created"
              column="created_at"
              sortBy={sortBy}
              sortDir={sortDir}
              onSort={onSort}
              className="hidden sm:table-cell"
            />
          </tr>
        </thead>
        <tbody>
          {documents.map((doc) => (
            <DocumentRow
              key={doc.id}
              document={doc}
              selected={selectedIds.has(doc.id)}
              onSelect={() => onToggleSelect(doc.id)}
              onClick={() => onPreview(doc)}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}
