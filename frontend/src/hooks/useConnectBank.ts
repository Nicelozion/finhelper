import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";

export function useConnectBank() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (bank: "vbank" | "abank" | "sbank") =>
      api<{ ok: boolean; bank: string; consent_id: string }>(`/api/banks/${bank}/connect`, {
        method: "POST",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["accounts"] });
    },
  });
}

