import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

interface TransactionParams {
  from?: string;
  to?: string;
  bank?: string;
}

export function useTransactions(params: TransactionParams) {
  return useQuery({
    queryKey: ["transactions", params],
    queryFn: () => api.getTransactions(params),
  });
}

