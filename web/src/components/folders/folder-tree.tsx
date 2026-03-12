import { useFolderTree } from "@/hooks/use-folders";
import { FolderNode } from "./folder-node";
import type { Folder } from "@/types/folder";

function buildTree(folders: Folder[]): Folder[] {
  const map = new Map<number, Folder>();
  const roots: Folder[] = [];

  for (const f of folders) {
    map.set(f.id, { ...f, children: [] });
  }

  for (const f of map.values()) {
    if (f.parent_id && map.has(f.parent_id)) {
      map.get(f.parent_id)!.children!.push(f);
    } else {
      roots.push(f);
    }
  }

  return roots;
}

export function FolderTreeComponent() {
  const { data: folders, isLoading, error } = useFolderTree();

  if (isLoading) {
    return <p className="px-3 text-xs text-muted-foreground">Loading...</p>;
  }

  if (error || !folders) {
    return (
      <p className="px-3 text-xs text-destructive">Failed to load folders</p>
    );
  }

  const tree = buildTree(folders);

  if (tree.length === 0) {
    return (
      <p className="px-3 text-xs text-muted-foreground">No folders yet</p>
    );
  }

  return (
    <div className="space-y-0.5">
      {tree.map((folder) => (
        <FolderNode key={folder.id} folder={folder} level={0} />
      ))}
    </div>
  );
}
