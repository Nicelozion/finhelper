import { useAccounts } from "../hooks/useAccounts";
import "./Accounts.css";

export function Accounts() {
  const { data: accounts, isLoading, isError, error, refetch } = useAccounts();

  if (isLoading) {
    return (
      <div className="accounts-page">
        <div className="skeleton-container">
          <div className="skeleton skeleton-header"></div>
          <div className="skeleton skeleton-row"></div>
          <div className="skeleton skeleton-row"></div>
          <div className="skeleton skeleton-row"></div>
        </div>
      </div>
    );
  }

  if (isError) {
    const requestId = (error as any)?.requestId;
    return (
      <div className="accounts-page">
        <div className="error-card">
          <h2>❌ Ошибка загрузки счетов</h2>
          <p>{error instanceof Error ? error.message : "Неизвестная ошибка"}</p>
          {requestId && <p className="request-id">Request ID: {requestId}</p>}
        </div>
      </div>
    );
  }

  const totalBalance =
    accounts?.reduce((sum, acc) => sum + acc.balance, 0) || 0;

  return (
    <div className="accounts-page">
      <div className="accounts-header">
        <h1>Счета</h1>
        <button onClick={() => refetch()} className="refresh-button">
          Обновить
        </button>
      </div>

      <div className="total-balance">
        <label>Суммарный баланс:</label>
        <span className="balance-value">{totalBalance.toFixed(2)}</span>
      </div>

      {!accounts || accounts.length === 0 ? (
        <div className="empty-state">Счетов пока нет</div>
      ) : (
        <div className="table-container">
          <table className="accounts-table">
            <thead>
              <tr>
                <th>Банк</th>
                <th>Номер</th>
                <th>Тип</th>
                <th>Валюта</th>
                <th>Баланс</th>
                <th>Владелец</th>
              </tr>
            </thead>
            <tbody>
              {accounts.map((account) => (
                <tr key={account.id}>
                  <td>{account.bank.toUpperCase()}</td>
                  <td>{account.ext_id}</td>
                  <td>{account.type}</td>
                  <td>{account.currency}</td>
                  <td className="balance">{account.balance.toFixed(2)}</td>
                  <td>{account.owner}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

