import type { Document } from "@/types/document";
import { DocumentCard } from "./document-card";

interface DocumentGridProps {
  documents: Document[];
  selectedIds: Set<number>;
  onToggleSelect: (id: number) => void;
  onPreview: (doc: Document) => void;
}

export function DocumentGrid({
  documents,
  selectedIds,
  onToggleSelect,
  onPreview,
}: DocumentGridProps) {
  if (documents.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <p className="text-lg font-medium">No documents found</p>
        <p className="text-sm">Upload a document or select a different folder.</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
      {documents.map((doc) => (
        <DocumentCard
          key={doc.id}
          document={doc}
          selected={selectedIds.has(doc.id)}
          onSelect={() => onToggleSelect(doc.id)}
          onClick={() => onPreview(doc)}
        />
      ))}
    </div>
  );
}
