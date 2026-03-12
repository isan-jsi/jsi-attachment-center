import { useState, useCallback, useRef } from "react";
import { useParams } from "@tanstack/react-router";
import {
  useDocuments,
  useDeleteDocuments,
  useCreateDocument,
} from "@/hooks/use-documents";
import {
  DocumentToolbar,
  type ViewMode,
} from "@/components/documents/document-toolbar";
import { DocumentGrid } from "@/components/documents/document-grid";
import { DocumentList } from "@/components/documents/document-list";
import { DocumentPreview } from "@/components/documents/document-preview";
import type { Document, DocumentListParams } from "@/types/document";
import { Button } from "@/components/ui/button";
import { ChevronLeft, ChevronRight } from "lucide-react";

export default function DocumentsPage() {
  // Get folderId from route params if present
  const params = useParams({ strict: false }) as { folderId?: string };
  const folderId = params.folderId ? Number(params.folderId) : undefined;

  // View state
  const [viewMode, setViewMode] = useState<ViewMode>("grid");
  const [sortBy, setSortBy] = useState("file_name");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("asc");
  const [page, setPage] = useState(1);
  const pageSize = 24;

  // Selection state
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

  // Preview state
  const [previewDoc, setPreviewDoc] = useState<Document | null>(null);
  const [previewOpen, setPreviewOpen] = useState(false);

  // Upload ref
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Query params
  const queryParams: DocumentListParams = {
    folder_id: folderId,
    page,
    page_size: pageSize,
    sort_by: sortBy,
    sort_dir: sortDir,
  };

  const { data, isLoading } = useDocuments(queryParams);
  const documents = data?.data ?? [];
  const meta = data?.meta;
  const totalPages = meta?.total_pages ?? 1;

  const deleteMutation = useDeleteDocuments();
  const createMutation = useCreateDocument();

  // Handlers
  const handleToggleSelect = useCallback((id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const handleSelectAll = useCallback(
    (selected: boolean) => {
      if (selected) {
        setSelectedIds(new Set(documents.map((d) => d.id)));
      } else {
        setSelectedIds(new Set());
      }
    },
    [documents]
  );

  const handleBulkDelete = useCallback(async () => {
    const ids = Array.from(selectedIds);
    if (ids.length === 0) return;
    if (!confirm(`Delete ${ids.length} document(s)?`)) return;
    await deleteMutation.mutateAsync(ids);
    setSelectedIds(new Set());
  }, [selectedIds, deleteMutation]);

  const handlePreview = useCallback((doc: Document) => {
    setPreviewDoc(doc);
    setPreviewOpen(true);
  }, []);

  const handleSort = useCallback(
    (column: string) => {
      if (column === sortBy) {
        setSortDir((d) => (d === "asc" ? "desc" : "asc"));
      } else {
        setSortBy(column);
        setSortDir("asc");
      }
    },
    [sortBy]
  );

  const handleUpload = useCallback(() => {
    fileInputRef.current?.click();
  }, []);

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file) return;

      await createMutation.mutateAsync({
        folder_id: folderId ?? 1,
        owner_class_library: "IBS",
        owner_class_name: "Default",
        document_type: file.type || "application/octet-stream",
        title: file.name,
        file,
      });

      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    },
    [folderId, createMutation]
  );

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">Documents</h1>
        <p className="text-sm text-muted-foreground">
          {meta
            ? `${meta.total_count} document(s)`
            : "Browse and manage documents."}
        </p>
      </div>

      <DocumentToolbar
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        sortBy={sortBy}
        onSortChange={setSortBy}
        sortDir={sortDir}
        onSortDirChange={setSortDir}
        selectedCount={selectedIds.size}
        onBulkDelete={handleBulkDelete}
        onUpload={handleUpload}
        isDeleting={deleteMutation.isPending}
      />

      {/* Hidden file input */}
      <input
        ref={fileInputRef}
        type="file"
        className="hidden"
        onChange={handleFileChange}
      />

      {/* Loading state */}
      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <p className="text-muted-foreground">Loading documents...</p>
        </div>
      ) : (
        <>
          {/* Mobile: always show grid/card layout */}
          <div className="md:hidden">
            <DocumentGrid
              documents={documents}
              selectedIds={selectedIds}
              onToggleSelect={handleToggleSelect}
              onPreview={handlePreview}
            />
          </div>

          {/* Desktop: respect the user's chosen view mode */}
          <div className="hidden md:block">
            {viewMode === "grid" ? (
              <DocumentGrid
                documents={documents}
                selectedIds={selectedIds}
                onToggleSelect={handleToggleSelect}
                onPreview={handlePreview}
              />
            ) : (
              <DocumentList
                documents={documents}
                selectedIds={selectedIds}
                onToggleSelect={handleToggleSelect}
                onSelectAll={handleSelectAll}
                onPreview={handlePreview}
                sortBy={sortBy}
                sortDir={sortDir}
                onSort={handleSort}
              />
            )}
          </div>
        </>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 pt-4">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => Math.max(1, p - 1))}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Preview panel */}
      <DocumentPreview
        document={previewDoc}
        open={previewOpen}
        onOpenChange={setPreviewOpen}
      />
    </div>
  );
}
