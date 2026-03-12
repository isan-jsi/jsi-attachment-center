import { useState, useCallback } from "react";
import { useSearch } from "@/hooks/use-search";
import { SearchBar } from "@/components/search/search-bar";
import { SearchFilters } from "@/components/search/search-filters";
import { DocumentGrid } from "@/components/documents/document-grid";
import { DocumentPreview } from "@/components/documents/document-preview";
import { Button } from "@/components/ui/button";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { Document } from "@/types/document";

export default function SearchPage() {
  const [query, setQuery] = useState("");
  const [folderId, setFolderId] = useState<number | undefined>();
  const [ownerClass, setOwnerClass] = useState<string | undefined>();
  const [documentType, setDocumentType] = useState<string | undefined>();
  const [page, setPage] = useState(1);
  const pageSize = 24;

  const [previewDoc, setPreviewDoc] = useState<Document | null>(null);
  const [previewOpen, setPreviewOpen] = useState(false);

  const { data, isLoading, isFetching } = useSearch({
    q: query,
    folder_id: folderId,
    owner_class_name: ownerClass,
    document_type: documentType,
    page,
    page_size: pageSize,
  });

  const documents = data?.data ?? [];
  const meta = data?.meta;
  const totalPages = meta?.total_pages ?? 1;

  const handleSearch = useCallback((q: string) => {
    setQuery(q);
    setPage(1);
  }, []);

  const handlePreview = useCallback((doc: Document) => {
    setPreviewDoc(doc);
    setPreviewOpen(true);
  }, []);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">Search</h1>
        <p className="text-sm text-muted-foreground">
          Search across all documents.
        </p>
      </div>

      <SearchBar value={query} onSearch={handleSearch} />

      <SearchFilters
        folderId={folderId}
        onFolderChange={(id) => {
          setFolderId(id);
          setPage(1);
        }}
        ownerClass={ownerClass}
        onOwnerChange={(o) => {
          setOwnerClass(o);
          setPage(1);
        }}
        documentType={documentType}
        onDocumentTypeChange={(t) => {
          setDocumentType(t);
          setPage(1);
        }}
      />

      {/* Results */}
      {!query ? (
        <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
          <p className="text-lg font-medium">Enter a search query</p>
          <p className="text-sm">
            Type in the search bar above to find documents.
          </p>
        </div>
      ) : isLoading ? (
        <div className="flex items-center justify-center py-16">
          <p className="text-muted-foreground">Searching...</p>
        </div>
      ) : (
        <>
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              {meta
                ? `${meta.total_count} result(s) found`
                : `${documents.length} result(s)`}
              {isFetching && " (updating...)"}
            </p>
          </div>

          <DocumentGrid
            documents={documents}
            selectedIds={new Set()}
            onToggleSelect={() => {}}
            onPreview={handlePreview}
          />

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
        </>
      )}

      <DocumentPreview
        document={previewDoc}
        open={previewOpen}
        onOpenChange={setPreviewOpen}
      />
    </div>
  );
}
