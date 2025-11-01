import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route, Link, useLocation } from "react-router-dom";
import { Connect } from "./pages/Connect";
import { Accounts } from "./pages/Accounts";
import { Transactions } from "./pages/Transactions";
import "./App.css";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function Navigation() {
  const location = useLocation();
  
  return (
    <nav className="navigation">
      <div className="nav-container">
        <Link to="/" className="nav-logo">
          FinHelper
        </Link>
        <div className="nav-links">
          <Link
            to="/connect"
            className={location.pathname === "/connect" ? "active" : ""}
          >
            Подключение
          </Link>
          <Link
            to="/accounts"
            className={location.pathname === "/accounts" ? "active" : ""}
          >
            Счета
          </Link>
          <Link
            to="/transactions"
            className={location.pathname === "/transactions" ? "active" : ""}
          >
            Операции
          </Link>
        </div>
      </div>
    </nav>
  );
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Accounts />} />
      <Route path="/connect" element={<Connect />} />
      <Route path="/accounts" element={<Accounts />} />
      <Route path="/transactions" element={<Transactions />} />
    </Routes>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="app">
          <Navigation />
          <main className="app-main">
            <AppRoutes />
          </main>
        </div>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
