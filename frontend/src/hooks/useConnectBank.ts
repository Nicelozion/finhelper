import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useConnectBank() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (bank: string) => api.connectBank(bank),
    onSuccess: () => {
      // Инвалидируем кеш счетов, чтобы обновить список после подключения
      qc.invalidateQueries({ queryKey: ["accounts"] });
    },
  });
}

