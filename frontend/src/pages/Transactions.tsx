import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { TransactionTable } from "@/components/TransactionTable"
import { TableSkeleton } from "@/components/LoadingSkeleton"
import { api } from "@/lib/api"
import { Search, Filter } from "lucide-react"

export function Transactions() {
  const [search, setSearch] = useState("")
  const [bankFilter, setBankFilter] = useState<string>("all")
  const [categoryFilter, setCategoryFilter] = useState<string>("all")

  const { data: transactions, isLoading } = useQuery({
    queryKey: ["transactions", bankFilter],
    queryFn: () => api.getTransactions({ bank: bankFilter !== "all" ? bankFilter : undefined }),
  })

  const filteredTransactions = transactions?.filter((tx) => {
    const matchesSearch = search === "" || 
      tx.merchant.toLowerCase().includes(search.toLowerCase()) ||
      tx.description.toLowerCase().includes(search.toLowerCase())
    const matchesCategory = categoryFilter === "all" || tx.category === categoryFilter
    return matchesSearch && matchesCategory
  }) || []

  const categories = Array.from(new Set(transactions?.map(tx => tx.category) || []))

  return (
    <div className="md:ml-64 pt-16 md:pt-0 p-4 md:p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        <div>
          <h1 className="text-3xl font-bold">Транзакции</h1>
          <p className="text-muted-foreground mt-2">Все ваши операции</p>
        </div>

        {/* Фильтры */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Filter className="h-5 w-5" />
                Фильтры
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label className="text-sm font-medium mb-2 block">Поиск</label>
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <input
                      type="text"
                      placeholder="Поиск по описанию..."
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                      className="w-full pl-10 pr-4 py-2 border rounded-md"
                    />
                  </div>
                </div>
                <div>
                  <label className="text-sm font-medium mb-2 block">Банк</label>
                  <select
                    value={bankFilter}
                    onChange={(e) => setBankFilter(e.target.value)}
                    className="w-full px-4 py-2 border rounded-md"
                  >
                    <option value="all">Все банки</option>
                    <option value="ABank">ABank</option>
                    <option value="VBank">VBank</option>
                    <option value="SBank">SBank</option>
                  </select>
                </div>
                <div>
                  <label className="text-sm font-medium mb-2 block">Категория</label>
                  <select
                    value={categoryFilter}
                    onChange={(e) => setCategoryFilter(e.target.value)}
                    className="w-full px-4 py-2 border rounded-md"
                  >
                    <option value="all">Все категории</option>
                    {categories.map((cat) => (
                      <option key={cat} value={cat}>{cat}</option>
                    ))}
                  </select>
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>

        {/* Таблица транзакций */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
        >
          {isLoading ? (
            <TableSkeleton />
          ) : (
            <TransactionTable transactions={filteredTransactions} />
          )}
        </motion.div>
      </div>
    </div>
  )
}
