import { useFolders } from "@/hooks/use-folders";
import { useOwners } from "@/hooks/use-owners";

interface SearchFiltersProps {
  folderId?: number;
  onFolderChange: (id?: number) => void;
  ownerClass?: string;
  onOwnerChange: (owner?: string) => void;
  documentType?: string;
  onDocumentTypeChange: (type?: string) => void;
}

const selectClass =
  "flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 touch-manipulation sm:w-auto";

export function SearchFilters({
  folderId,
  onFolderChange,
  ownerClass,
  onOwnerChange,
  documentType,
  onDocumentTypeChange,
}: SearchFiltersProps) {
  const { data: folders } = useFolders();
  const { data: owners } = useOwners();

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
      {/* Folder filter */}
      <select
        value={folderId ?? ""}
        onChange={(e) =>
          onFolderChange(e.target.value ? Number(e.target.value) : undefined)
        }
        className={selectClass}
      >
        <option value="">All Folders</option>
        {folders?.map((f) => (
          <option key={f.id} value={f.id}>
            {f.name}
          </option>
        ))}
      </select>

      {/* Owner filter */}
      <select
        value={ownerClass ?? ""}
        onChange={(e) => onOwnerChange(e.target.value || undefined)}
        className={selectClass}
      >
        <option value="">All Owners</option>
        {owners?.map((o) => (
          <option
            key={`${o.owner_class_library}.${o.owner_class_name}`}
            value={o.owner_class_name}
          >
            {o.owner_class_library}.{o.owner_class_name} ({o.document_count})
          </option>
        ))}
      </select>

      {/* Document type filter */}
      <select
        value={documentType ?? ""}
        onChange={(e) => onDocumentTypeChange(e.target.value || undefined)}
        className={selectClass}
      >
        <option value="">All Types</option>
        <option value="application/pdf">PDF</option>
        <option value="image/jpeg">JPEG</option>
        <option value="image/png">PNG</option>
        <option value="application/msword">Word</option>
        <option value="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet">
          Excel
        </option>
        <option value="text/plain">Text</option>
      </select>
    </div>
  );
}
