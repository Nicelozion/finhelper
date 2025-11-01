import { useQuery } from "@tanstack/react-query";
import { api } from "../lib/api";
import type { Transaction } from "../types/api";

interface TransactionParams {
  from?: string;
  to?: string;
  bank?: string;
}

export function useTransactions(params: TransactionParams) {
  const search = new URLSearchParams(
    Object.fromEntries(
      Object.entries(params).filter(([_, v]) => v !== undefined && v !== "")
    ) as Record<string, string>
  );
  return useQuery({
    queryKey: ["tx", params],
    queryFn: async () =>
      (await api<Transaction[]>(`/api/transactions?${search.toString()}`)).data,
  });
}

