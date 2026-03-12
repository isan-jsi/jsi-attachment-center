import { useVirtualizer } from "@tanstack/react-virtual";
import { useRef } from "react";
import type { Document } from "@/types/document";

interface Props {
  documents: Document[];
  onSelect: (doc: Document) => void;
}

export function DocumentListVirtual({ documents, onSelect }: Props) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: documents.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 64,
    overscan: 10,
  });

  return (
    <div ref={parentRef} className="h-[calc(100vh-200px)] overflow-auto">
      <div style={{ height: `${virtualizer.getTotalSize()}px`, position: "relative" }}>
        {virtualizer.getVirtualItems().map((virtualRow) => {
          const doc = documents[virtualRow.index];
          return (
            <div
              key={doc.id}
              className="absolute w-full cursor-pointer hover:bg-muted/50 border-b px-4 py-3 flex items-center gap-4"
              style={{
                height: `${virtualRow.size}px`,
                transform: `translateY(${virtualRow.start}px)`,
              }}
              onClick={() => onSelect(doc)}
            >
              <span className="truncate flex-1 font-medium">{doc.file_name}</span>
              <span className="text-sm text-muted-foreground hidden md:inline">{doc.mime_type}</span>
              <span className="text-sm text-muted-foreground hidden sm:inline">
                {(doc.file_size / 1024).toFixed(1)} KB
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
