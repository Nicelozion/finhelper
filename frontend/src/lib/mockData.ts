export interface MockAccount {
  bank: string;
  balance: number;
  currency: string;
  accountsCount: number;
}

export interface MockTransaction {
  id: string;
  date: string;
  merchant: string;
  category: string;
  amount: number;
  bank: string;
  description?: string;
}

export interface MockCategory {
  name: string;
  amount: number;
  color: string;
}

export const mockAccounts: MockAccount[] = [
  { bank: "ABank", balance: 55000, currency: "RUB", accountsCount: 2 },
  { bank: "VBank", balance: 82000, currency: "RUB", accountsCount: 1 },
  { bank: "SBank", balance: 120000, currency: "RUB", accountsCount: 3 },
];

export const mockTransactions: MockTransaction[] = [
  {
    id: "1",
    date: "2025-10-20",
    merchant: "Пятёрочка",
    category: "Еда",
    amount: -1250,
    bank: "SBank",
    description: "Покупка продуктов",
  },
  {
    id: "2",
    date: "2025-10-21",
    merchant: "Зарплата",
    category: "Доход",
    amount: 70000,
    bank: "ABank",
    description: "Зарплата за октябрь",
  },
  {
    id: "3",
    date: "2025-10-22",
    merchant: "Яндекс.Такси",
    category: "Транспорт",
    amount: -560,
    bank: "VBank",
    description: "Поездка на работу",
  },
  {
    id: "4",
    date: "2025-10-23",
    merchant: "МТС",
    category: "Связь",
    amount: -500,
    bank: "ABank",
    description: "Пополнение счёта",
  },
  {
    id: "5",
    date: "2025-10-24",
    merchant: "Кафе",
    category: "Еда",
    amount: -850,
    bank: "SBank",
    description: "Обед",
  },
];

export const mockCategories: MockCategory[] = [
  { name: "Еда", amount: 15000, color: "#ef4444" },
  { name: "Транспорт", amount: 5000, color: "#3b82f6" },
  { name: "Развлечения", amount: 8000, color: "#10b981" },
  { name: "Связь", amount: 2000, color: "#f59e0b" },
  { name: "Прочее", amount: 3000, color: "#8b5cf6" },
];

export const mockMonthlyData = [
  { month: "Янв", income: 70000, expenses: 45000 },
  { month: "Фев", income: 70000, expenses: 48000 },
  { month: "Мар", income: 70000, expenses: 42000 },
  { month: "Апр", income: 70000, expenses: 50000 },
  { month: "Май", income: 70000, expenses: 47000 },
  { month: "Июн", income: 70000, expenses: 49000 },
];

export const mockUserProfile = {
  name: "Иван Иванов",
  email: "ivan@example.com",
  subscription: "Free" as "Free" | "Pro",
  connectedBanks: ["ABank", "VBank", "SBank"],
};

