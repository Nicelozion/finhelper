import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"
import { motion, AnimatePresence } from "framer-motion"
import { Sidebar } from "@/components/Sidebar"
import { MobileNav } from "@/components/MobileNav"
import { Dashboard } from "@/pages/Dashboard"
import { Transactions } from "@/pages/Transactions"
import { Analytics } from "@/pages/Analytics"
import { Settings } from "@/pages/Settings"
import { Subscription } from "@/pages/Subscription"
import { Onboarding } from "@/pages/Onboarding"
import { Connect } from "@/pages/Connect"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      <Sidebar />
      <MobileNav />
      <AnimatePresence mode="wait">
        <motion.main
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
        >
          {children}
        </motion.main>
      </AnimatePresence>
    </div>
  )
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/onboarding" element={<Onboarding />} />
          <Route
            path="/connect"
            element={
              <AppLayout>
                <Connect />
              </AppLayout>
            }
          />
          <Route
            path="/dashboard"
            element={
              <AppLayout>
                <Dashboard />
              </AppLayout>
            }
          />
          <Route
            path="/transactions"
            element={
              <AppLayout>
                <Transactions />
              </AppLayout>
            }
          />
          <Route
            path="/analytics"
            element={
              <AppLayout>
                <Analytics />
              </AppLayout>
            }
          />
          <Route
            path="/settings"
            element={
              <AppLayout>
                <Settings />
              </AppLayout>
            }
          />
          <Route
            path="/subscription"
            element={
              <AppLayout>
                <Subscription />
              </AppLayout>
            }
          />
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

export default App
