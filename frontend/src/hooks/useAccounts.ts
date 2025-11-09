import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useAccounts() {
  return useQuery({
    queryKey: ["accounts"],
    queryFn: api.getAccounts,
  });
}

