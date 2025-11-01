import { useState } from "react";
import { Link } from "react-router-dom";
import { useConnectBank } from "../hooks/useConnectBank";
import "./Connect.css";

export function Connect() {
  const [connectedBank, setConnectedBank] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [requestId, setRequestId] = useState<string | null>(null);
  const connectBank = useConnectBank();

  const handleConnect = async (bank: "vbank" | "abank" | "sbank") => {
    setError(null);
    setRequestId(null);
    setConnectedBank(null);

    connectBank.mutate(bank, {
      onSuccess: (response) => {
        if (response.data.ok) {
          setConnectedBank(bank);
        } else {
          setError("Подключение не удалось");
        }
        if (response.requestId) {
          setRequestId(response.requestId);
        }
      },
      onError: (error: any) => {
        setError(error.message || "Ошибка подключения");
        setRequestId(error.requestId || null);
      },
    });
  };

  return (
    <div className="connect-page">
      <h1>Подключение банка</h1>
      <p>Выберите банк для подключения</p>

      <div className="banks-grid">
        <button
          className="bank-button"
          onClick={() => handleConnect("vbank")}
          disabled={connectBank.isPending}
        >
          VBank
        </button>
        <button
          className="bank-button"
          onClick={() => handleConnect("abank")}
          disabled={connectBank.isPending}
        >
          ABank
        </button>
        <button
          className="bank-button"
          onClick={() => handleConnect("sbank")}
          disabled={connectBank.isPending}
        >
          SBank
        </button>
      </div>

      {connectBank.isPending && (
        <div className="status status-loading">Подключение...</div>
      )}

      {connectedBank && (
        <div className="status status-success">
          <p>✅ Банк {connectedBank.toUpperCase()} успешно подключен!</p>
          <Link to="/accounts" className="link-button">
            Перейти к счетам
          </Link>
        </div>
      )}

      {error && (
        <div className="status status-error">
          <p>❌ {error}</p>
          {requestId && (
            <p className="request-id">Request ID: {requestId}</p>
          )}
        </div>
      )}
    </div>
  );
}

