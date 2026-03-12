import { useQuery } from "@tanstack/react-query";
import { ownersApi } from "@/api/owners";

export const ownerKeys = {
  all: ["owners"] as const,
};

export function useOwners() {
  return useQuery({
    queryKey: ownerKeys.all,
    queryFn: () => ownersApi.list().then((r) => r.data),
  });
}
