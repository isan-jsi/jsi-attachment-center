import { api } from "./client";
import type { ApiEnvelope } from "@/types/api";
import type { Owner } from "@/types/owner";

const BASE = "/api/v1/owners";

export const ownersApi = {
  list(): Promise<ApiEnvelope<Owner[]>> {
    return api.get<Owner[]>(BASE);
  },
};
