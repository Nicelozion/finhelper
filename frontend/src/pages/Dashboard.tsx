import { useQuery } from "@tanstack/react-query"
import { useNavigate } from "react-router-dom"
import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { BankCard } from "@/components/BankCard"
import { TransactionTable } from "@/components/TransactionTable"
import { MonthlyTrend } from "@/components/Charts/MonthlyTrend"
import { ExpensesDonut } from "@/components/Charts/ExpensesDonut"
import { CardSkeleton } from "@/components/LoadingSkeleton"
import { api } from "@/lib/api"
import { mockMonthlyData, mockCategories } from "@/lib/mockData"
import { TrendingUp, ArrowRight } from "lucide-react"

export function Dashboard() {
  const navigate = useNavigate()

  const { data: accounts, isLoading: accountsLoading } = useQuery({
    queryKey: ["accounts"],
    queryFn: api.getAccounts,
  })

  const { data: transactions, isLoading: transactionsLoading } = useQuery({
    queryKey: ["transactions"],
    queryFn: () => api.getTransactions(),
  })

  // Группируем счета по банкам
  const bankData = accounts?.reduce((acc, account) => {
    if (!acc[account.bank]) {
      acc[account.bank] = {
        bank: account.bank,
        balance: 0,
        currency: account.currency,
        accountsCount: 0,
      }
    }
    acc[account.bank].balance += account.balance
    acc[account.bank].accountsCount += 1
    return acc
  }, {} as Record<string, { bank: string; balance: number; currency: string; accountsCount: number }>)

  const bankCards = bankData ? Object.values(bankData) : []
  const totalBalance = bankCards.reduce((sum, bank) => sum + bank.balance, 0)
  const recentTransactions = transactions?.slice(0, 5) || []

  return (
    <div className="md:ml-64 pt-16 md:pt-0 p-4 md:p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Заголовок */}
        <div>
          <h1 className="text-3xl font-bold">Dashboard</h1>
          <p className="text-muted-foreground mt-2">Обзор ваших финансов</p>
        </div>

        {/* Суммарный баланс */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <Card className="bg-gradient-to-r from-primary/10 to-primary/5 border-primary/20">
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Суммарный баланс</p>
                  <h2 className="text-4xl font-bold mt-2">
                    {totalBalance.toLocaleString("ru-RU")} ₽
                  </h2>
                  <div className="flex items-center gap-2 mt-2 text-sm text-green-600">
                    <TrendingUp className="h-4 w-4" />
                    <span>+12% за месяц</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>

        {/* Карточки банков */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.1 }}
        >
          <h3 className="text-xl font-semibold mb-4">Банки</h3>
          {accountsLoading ? (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {[...Array(3)].map((_, i) => (
                <CardSkeleton key={i} />
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {bankCards.map((bank) => (
                <BankCard key={bank.bank} {...bank} />
              ))}
            </div>
          )}
        </motion.div>

        {/* Графики */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <MonthlyTrend data={mockMonthlyData} />
          </motion.div>
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            <ExpensesDonut data={mockCategories} />
          </motion.div>
        </div>

        {/* Последние транзакции */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.4 }}
        >
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-xl font-semibold">Последние транзакции</h3>
            <Button
              variant="outline"
              onClick={() => navigate("/transactions")}
              className="gap-2"
            >
              Показать все
              <ArrowRight className="h-4 w-4" />
            </Button>
          </div>
          {transactionsLoading ? (
            <CardSkeleton />
          ) : (
            <TransactionTable transactions={recentTransactions} />
          )}
        </motion.div>
      </div>
    </div>
  )
}

