import { useState } from "react";
import { useTransactions } from "../hooks/useTransactions";
import "./Transactions.css";

export function Transactions() {
  const [bank, setBank] = useState<string>("");
  const [from, setFrom] = useState<string>("");
  const [to, setTo] = useState<string>("");

  const params: { from?: string; to?: string; bank?: string } = {};
  if (bank && bank !== "all") params.bank = bank;
  if (from) {
    const fromDate = new Date(from);
    fromDate.setHours(0, 0, 0, 0);
    params.from = fromDate.toISOString();
  }
  if (to) {
    const toDate = new Date(to);
    toDate.setHours(23, 59, 59, 999);
    params.to = toDate.toISOString();
  }

  const { data: transactions, isLoading, isError, error } = useTransactions(params);

  const totalExpenses =
    transactions?.filter((t) => t.amount < 0).reduce((sum, t) => sum + Math.abs(t.amount), 0) || 0;
  const totalIncome =
    transactions?.filter((t) => t.amount > 0).reduce((sum, t) => sum + t.amount, 0) || 0;

  if (isLoading) {
    return (
      <div className="transactions-page">
        <div className="skeleton-container">
          <div className="skeleton skeleton-header"></div>
          <div className="skeleton skeleton-row"></div>
          <div className="skeleton skeleton-row"></div>
        </div>
      </div>
    );
  }

  if (isError) {
    const requestId = (error as any)?.requestId;
    return (
      <div className="transactions-page">
        <div className="error-card">
          <h2>❌ Ошибка загрузки операций</h2>
          <p>{error instanceof Error ? error.message : "Неизвестная ошибка"}</p>
          {requestId && <p className="request-id">Request ID: {requestId}</p>}
        </div>
      </div>
    );
  }

  return (
    <div className="transactions-page">
      <h1>Операции</h1>

      <div className="filters">
        <div className="filter-group">
          <label htmlFor="bank">Банк:</label>
          <select
            id="bank"
            value={bank}
            onChange={(e) => setBank(e.target.value)}
          >
            <option value="all">Все</option>
            <option value="vbank">VBank</option>
            <option value="abank">ABank</option>
            <option value="sbank">SBank</option>
          </select>
        </div>

        <div className="filter-group">
          <label htmlFor="from">От:</label>
          <input
            id="from"
            type="date"
            value={from}
            onChange={(e) => setFrom(e.target.value)}
          />
        </div>

        <div className="filter-group">
          <label htmlFor="to">До:</label>
          <input
            id="to"
            type="date"
            value={to}
            onChange={(e) => setTo(e.target.value)}
          />
        </div>
      </div>

      {!transactions || transactions.length === 0 ? (
        <div className="empty-state">
          Данных за выбранный период нет
        </div>
      ) : (
        <>
          <div className="table-container">
            <table className="transactions-table">
              <thead>
                <tr>
                  <th>Дата</th>
                  <th>Сумма</th>
                  <th>Валюта</th>
                  <th>Мерчант</th>
                  <th>Категория</th>
                  <th>Описание</th>
                  <th>Банк</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((tx) => (
                  <tr key={tx.id}>
                    <td>{new Date(tx.date).toLocaleDateString("ru-RU")}</td>
                    <td className={tx.amount < 0 ? "expense" : "income"}>
                      {tx.amount > 0 ? "+" : ""}
                      {tx.amount.toFixed(2)}
                    </td>
                    <td>{tx.currency}</td>
                    <td>{tx.merchant || "-"}</td>
                    <td>{tx.category || "-"}</td>
                    <td>{tx.description || "-"}</td>
                    <td>{tx.bank.toUpperCase()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="summary">
            <div className="summary-item">
              <span className="summary-label">Всего операций:</span>
              <span className="summary-value">{transactions.length}</span>
            </div>
            <div className="summary-item">
              <span className="summary-label">Сумма расходов:</span>
              <span className="summary-value expense">{totalExpenses.toFixed(2)}</span>
            </div>
            <div className="summary-item">
              <span className="summary-label">Сумма доходов:</span>
              <span className="summary-value income">+{totalIncome.toFixed(2)}</span>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

