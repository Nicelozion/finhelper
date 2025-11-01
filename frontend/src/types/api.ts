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

