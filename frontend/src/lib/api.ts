import axios from "axios";
import type { AxiosInstance, AxiosError } from "axios";
import { mockAccounts, mockTransactions, mockCategories, mockMonthlyData, mockUserProfile } from "./mockData";
import type { MockAccount, MockTransaction, MockCategory } from "./mockData";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

// Создаём экземпляр axios
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

// Интерцептор для добавления X-Request-Id
apiClient.interceptors.request.use((config) => {
  if (!config.headers["X-Request-Id"]) {
    config.headers["X-Request-Id"] = crypto.randomUUID();
  }
  return config;
});

// Интерцептор для обработки ошибок
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.code === "ECONNABORTED" || error.message.includes("timeout")) {
      console.warn("API timeout, using mock data");
    } else if (error.response?.status === 500 || !error.response) {
      console.warn("API error, using mock data:", error.message);
    }
    throw error;
  }
);

// Типы для API ответов
export interface Account {
  id: string;
  ext_id: string;
  bank: string;
  type: string;
  currency: string;
  balance: number;
  owner: string;
}

export interface Transaction {
  id: string;
  date: string;
  amount: number;
  currency: string;
  merchant: string;
  category: string;
  description: string;
  bank: string;
}

export interface UserProfile {
  name: string;
  email: string;
  subscription: "Free" | "Pro";
  connectedBanks: string[];
}

// Функция для безопасного вызова API с fallback на моки
async function withFallback<T>(
  apiCall: () => Promise<T>,
  mockData: T
): Promise<T> {
  try {
    return await apiCall();
  } catch (error) {
    console.warn("Using mock data due to API error");
    return mockData;
  }
}

// API функции
export const api = {
  // Подключение банка
  connectBank: async (bank: string): Promise<{ ok: boolean; bank: string; consent_id: string }> => {
    return withFallback(
      async () => {
        const response = await apiClient.post(`/api/banks/${bank}/connect`);
        return response.data;
      },
      { ok: true, bank, consent_id: `mock-consent-${bank}` }
    );
  },

  // Получение счетов
  getAccounts: async (): Promise<Account[]> => {
    return withFallback(
      async () => {
        const response = await apiClient.get("/api/accounts");
        return response.data;
      },
      mockAccounts.map((acc, idx) => ({
        id: `acc-${idx}`,
        ext_id: `****${Math.floor(Math.random() * 10000)}`,
        bank: acc.bank,
        type: "current",
        currency: acc.currency,
        balance: acc.balance,
        owner: "Иван Иванов",
      }))
    );
  },

  // Получение транзакций
  getTransactions: async (params?: {
    bank?: string;
    from?: string;
    to?: string;
  }): Promise<Transaction[]> => {
    return withFallback(
      async () => {
        const queryParams = new URLSearchParams();
        if (params?.bank) queryParams.append("bank", params.bank);
        if (params?.from) queryParams.append("from", params.from);
        if (params?.to) queryParams.append("to", params.to);

        const response = await apiClient.get(`/api/transactions?${queryParams.toString()}`);
        return response.data;
      },
      mockTransactions.map((tx) => ({
        id: tx.id,
        date: tx.date,
        amount: tx.amount,
        currency: "RUB",
        merchant: tx.merchant,
        category: tx.category,
        description: tx.description || "",
        bank: tx.bank,
      }))
    );
  },

  // Получение профиля пользователя
  getUserProfile: async (): Promise<UserProfile> => {
    return withFallback(
      async () => {
        const response = await apiClient.get("/api/user/profile");
        return response.data;
      },
      mockUserProfile
    );
  },

  // Оформление подписки
  subscribe: async (plan: "Pro"): Promise<{ success: boolean; message: string }> => {
    return withFallback(
      async () => {
        const response = await apiClient.post("/api/user/subscribe", { plan });
        return response.data;
      },
      { success: true, message: "Подписка оформлена успешно" }
    );
  },
};

export default api;
