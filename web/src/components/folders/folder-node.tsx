import { useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useDrop } from "@/hooks/use-dnd";
import { ChevronRight, ChevronDown, Folder as FolderIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Folder } from "@/types/folder";

interface FolderNodeProps {
  folder: Folder;
  level: number;
  onDocumentDrop?: (folderId: number, documentId: string) => void;
}

export function FolderNode({ folder, level, onDocumentDrop }: FolderNodeProps) {
  const [expanded, setExpanded] = useState(false);
  const navigate = useNavigate();
  const hasChildren = folder.children && folder.children.length > 0;

  const { isOver, dropRef } = useDrop({
    accept: "document",
    onDrop: (documentId: string) => onDocumentDrop?.(folder.id, documentId),
  });

  return (
    <div>
      <div
        ref={dropRef as React.RefCallback<HTMLDivElement>}
        className={cn(
          "flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm hover:bg-accent/50",
          isOver && "bg-accent",
        )}
        style={{ paddingLeft: `${level * 12 + 8}px` }}
        onClick={() =>
          navigate({
            to: "/documents/$folderId",
            params: { folderId: String(folder.id) },
          })
        }
      >
        <button
          onClick={(e) => {
            e.stopPropagation();
            setExpanded(!expanded);
          }}
          className="flex h-4 w-4 items-center justify-center"
        >
          {hasChildren ? (
            expanded ? (
              <ChevronDown size={14} />
            ) : (
              <ChevronRight size={14} />
            )
          ) : (
            <span className="w-4" />
          )}
        </button>
        <FolderIcon size={14} className="text-muted-foreground" />
        <span className="flex-1 truncate">{folder.name}</span>
        {folder.document_count !== undefined && folder.document_count > 0 && (
          <span className="rounded-full bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
            {folder.document_count}
          </span>
        )}
      </div>
      {expanded && hasChildren && (
        <div>
          {folder.children!.map((child) => (
            <FolderNode
              key={child.id}
              folder={child}
              level={level + 1}
              onDocumentDrop={onDocumentDrop}
            />
          ))}
        </div>
      )}
    </div>
  );
}
