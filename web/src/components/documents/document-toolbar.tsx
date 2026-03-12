import { LayoutGrid, List, Upload, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";

export type ViewMode = "grid" | "list";

interface DocumentToolbarProps {
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
  sortBy: string;
  onSortChange: (sort: string) => void;
  sortDir: "asc" | "desc";
  onSortDirChange: (dir: "asc" | "desc") => void;
  selectedCount: number;
  onBulkDelete: () => void;
  onUpload: () => void;
  isDeleting?: boolean;
}

export function DocumentToolbar({
  viewMode,
  onViewModeChange,
  sortBy,
  onSortChange,
  sortDir,
  onSortDirChange,
  selectedCount,
  onBulkDelete,
  onUpload,
  isDeleting,
}: DocumentToolbarProps) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      {/* View mode toggle — hidden on mobile since we always show grid there */}
      <div className="hidden md:flex items-center rounded-md border border-input">
        <Button
          variant={viewMode === "grid" ? "secondary" : "ghost"}
          size="sm"
          className="rounded-r-none touch-manipulation"
          onClick={() => onViewModeChange("grid")}
        >
          <LayoutGrid className="h-4 w-4" />
        </Button>
        <Button
          variant={viewMode === "list" ? "secondary" : "ghost"}
          size="sm"
          className="rounded-l-none touch-manipulation"
          onClick={() => onViewModeChange("list")}
        >
          <List className="h-4 w-4" />
        </Button>
      </div>

      {/* Sort controls */}
      <select
        value={sortBy}
        onChange={(e) => onSortChange(e.target.value)}
        className="flex h-9 rounded-md border border-input bg-background px-3 py-1 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 touch-manipulation"
      >
        <option value="file_name">Name</option>
        <option value="created_at">Date Created</option>
        <option value="updated_at">Date Modified</option>
        <option value="file_size">Size</option>
        <option value="document_type">Type</option>
      </select>

      <Button
        variant="outline"
        size="sm"
        onClick={() => onSortDirChange(sortDir === "asc" ? "desc" : "asc")}
        className="touch-manipulation"
      >
        {sortDir === "asc" ? "A-Z" : "Z-A"}
      </Button>

      <div className="flex-1" />

      {/* Bulk actions */}
      {selectedCount > 0 && (
        <Button
          variant="destructive"
          size="sm"
          onClick={onBulkDelete}
          disabled={isDeleting}
          className="touch-manipulation"
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete {selectedCount}
        </Button>
      )}

      {/* Upload */}
      <Button size="sm" onClick={onUpload} className="touch-manipulation">
        <Upload className="mr-2 h-4 w-4" />
        Upload
      </Button>
    </div>
  );
}
