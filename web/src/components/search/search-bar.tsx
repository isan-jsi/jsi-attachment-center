import { useState, useCallback } from "react";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

interface SearchBarProps {
  value: string;
  onSearch: (query: string) => void;
}

export function SearchBar({ value, onSearch }: SearchBarProps) {
  const [draft, setDraft] = useState(value);

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      onSearch(draft.trim());
    },
    [draft, onSearch]
  );

  return (
    <form onSubmit={handleSubmit} className="flex items-center gap-2">
      <div className="relative flex-1">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder="Search documents by name, title, or content..."
          className="pl-10"
        />
      </div>
      <Button type="submit" size="sm">
        Search
      </Button>
    </form>
  );
}
