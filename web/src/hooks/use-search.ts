import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { searchApi, type SearchParams } from "@/api/search";

export const searchKeys = {
  all: ["search"] as const,
  results: (params: SearchParams) => ["search", params] as const,
};

export function useSearch(params: SearchParams) {
  return useQuery({
    queryKey: searchKeys.results(params),
    queryFn: () => searchApi.search(params),
    placeholderData: keepPreviousData,
    enabled: params.q.length > 0,
  });
}
