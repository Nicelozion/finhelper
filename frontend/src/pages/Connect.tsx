import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { api, type Account } from "@/lib/api"
import { Building2, Plus, CheckCircle2 } from "lucide-react"
import { cn } from "@/lib/utils"

export function Connect() {
  const [showConnectForm, setShowConnectForm] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { data: accounts, isLoading: accountsLoading, refetch: refetchAccounts } = useQuery({
    queryKey: ["accounts"],
    queryFn: api.getAccounts,
  })

  // Определяем подключенные банки по наличию счетов
  const connectedBanks = new Set<string>()
  if (accounts) {
    accounts.forEach((acc) => connectedBanks.add(acc.bank))
  }
  const connectedBanksList = Array.from(connectedBanks)

  // Группируем счета по банкам
  const accountsByBank = new Map<string, Account[]>()
  if (accounts) {
    accounts.forEach((acc) => {
      const bankAccounts = accountsByBank.get(acc.bank) || []
      accountsByBank.set(acc.bank, [...bankAccounts, acc])
    })
  }

  const handleConnect = async (bank: "ABank" | "VBank" | "SBank") => {
    setError(null)
    setShowConnectForm(false)

    try {
      const result = await api.connectBank(bank.toLowerCase())
      if (result.ok) {
        refetchAccounts()
      } else {
        setError("Подключение не удалось")
      }
    } catch (err: any) {
      let errorMessage = err.message || "Ошибка подключения"
      if (errorMessage.includes("connection refused") || errorMessage.includes("dial tcp")) {
        errorMessage = "Банковский API недоступен. Убедитесь, что банковский сервис запущен."
      } else if (errorMessage.includes("Failed to connect")) {
        errorMessage = "Не удалось подключиться к банку. Проверьте настройки backend."
      }
      setError(errorMessage)
    }
  }

  const getBankDisplayName = (bank: string) => {
    const names: Record<string, string> = {
      vbank: "VBank",
      abank: "ABank",
      sbank: "SBank",
    }
    return names[bank.toLowerCase()] || bank.toUpperCase()
  }

  const totalBalance = accounts?.reduce((sum, acc) => sum + acc.balance, 0) || 0

  return (
    <div className="md:ml-64 pt-16 md:pt-0 p-4 md:p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        <div>
          <h1 className="text-3xl font-bold">Управление банками</h1>
          <p className="text-muted-foreground mt-2">Подключите свои банки для агрегации данных</p>
        </div>

        {/* Секция 1: Подключение банка */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle>Подключить банк</CardTitle>
                <CardDescription>Добавьте новый банк для отслеживания</CardDescription>
              </div>
              {!showConnectForm && (
                <Button onClick={() => setShowConnectForm(true)} className="gap-2">
                  <Plus className="h-4 w-4" />
                  Подключить банк
                </Button>
              )}
            </CardHeader>
            {showConnectForm && (
              <CardContent>
                <div className="space-y-4">
                  <p className="text-sm text-muted-foreground">
                    Выберите банк для подключения
                  </p>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    {(["ABank", "VBank", "SBank"] as const).map((bank) => {
                      const isConnected = connectedBanks.has(bank)
                      return (
                        <motion.button
                          key={bank}
                          whileHover={{ scale: 1.02 }}
                          whileTap={{ scale: 0.98 }}
                          onClick={() => handleConnect(bank)}
                          disabled={isConnected}
                          className={cn(
                            "relative p-6 border-2 rounded-lg transition-all",
                            isConnected
                              ? "border-muted bg-muted/50 cursor-not-allowed opacity-60"
                              : "border-primary hover:border-primary/80 hover:bg-primary/5 cursor-pointer"
                          )}
                        >
                          <Building2 className="h-8 w-8 mx-auto mb-2" />
                          <div className="font-semibold">{bank}</div>
                          {isConnected && (
                            <div className="absolute top-2 right-2">
                              <CheckCircle2 className="h-5 w-5 text-green-600" />
                            </div>
                          )}
                        </motion.button>
                      )
                    })}
                  </div>
                  {error && (
                    <div className="p-3 bg-destructive/10 text-destructive rounded-md text-sm">
                      {error}
                    </div>
                  )}
                  <Button
                    variant="outline"
                    onClick={() => {
                      setShowConnectForm(false)
                      setError(null)
                    }}
                  >
                    Отмена
                  </Button>
                </div>
              </CardContent>
            )}
          </Card>
        </motion.div>

        {/* Секция 2: Подключенные банки */}
        {connectedBanksList.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <span>Подключенные банки</span>
                  <span className="text-sm font-normal text-muted-foreground">
                    {connectedBanksList.length} {connectedBanksList.length === 1 ? "банк" : "банков"}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  {connectedBanksList.map((bank) => {
                    const bankAccounts = accountsByBank.get(bank) || []
                    const bankBalance = bankAccounts.reduce((sum, acc) => sum + acc.balance, 0)
                    return (
                      <Card key={bank} className="border-2">
                        <CardHeader className="pb-3">
                          <div className="flex items-center justify-between">
                            <CardTitle className="text-lg">{getBankDisplayName(bank)}</CardTitle>
                            <CheckCircle2 className="h-5 w-5 text-green-600" />
                          </div>
                        </CardHeader>
                        <CardContent>
                          <div className="space-y-2">
                            <div>
                              <p className="text-sm text-muted-foreground">Счетов</p>
                              <p className="text-lg font-semibold">{bankAccounts.length}</p>
                            </div>
                            <div>
                              <p className="text-sm text-muted-foreground">Баланс</p>
                              <p className="text-lg font-semibold">
                                {bankBalance.toLocaleString("ru-RU")} {bankAccounts[0]?.currency || "RUB"}
                              </p>
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    )
                  })}
                </div>
              </CardContent>
            </Card>
          </motion.div>
        )}

        {/* Секция 3: Счета */}
        {connectedBanksList.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
          >
            <Card>
              <CardHeader>
                <CardTitle>Счета подключенных банков</CardTitle>
                <CardDescription>
                  Общий баланс: {totalBalance.toLocaleString("ru-RU")} ₽
                </CardDescription>
              </CardHeader>
              <CardContent>
                {accountsLoading ? (
                  <div className="space-y-4">
                    {[...Array(3)].map((_, i) => (
                      <div key={i} className="h-16 bg-muted rounded animate-pulse" />
                    ))}
                  </div>
                ) : accounts && accounts.length > 0 ? (
                  <div className="space-y-6">
                    {connectedBanksList.map((bank) => {
                      const bankAccounts = accountsByBank.get(bank) || []
                      return (
                        <div key={bank}>
                          <h3 className="text-lg font-semibold mb-3">{getBankDisplayName(bank)}</h3>
                          <div className="overflow-x-auto">
                            <table className="w-full">
                              <thead>
                                <tr className="border-b">
                                  <th className="text-left py-2 px-4 text-sm font-medium text-muted-foreground">
                                    Номер
                                  </th>
                                  <th className="text-left py-2 px-4 text-sm font-medium text-muted-foreground">
                                    Тип
                                  </th>
                                  <th className="text-left py-2 px-4 text-sm font-medium text-muted-foreground">
                                    Валюта
                                  </th>
                                  <th className="text-right py-2 px-4 text-sm font-medium text-muted-foreground">
                                    Баланс
                                  </th>
                                  <th className="text-left py-2 px-4 text-sm font-medium text-muted-foreground">
                                    Владелец
                                  </th>
                                </tr>
                              </thead>
                              <tbody>
                                {bankAccounts.map((account) => (
                                  <tr key={account.id} className="border-b hover:bg-muted/50">
                                    <td className="py-3 px-4 text-sm">{account.ext_id || account.id}</td>
                                    <td className="py-3 px-4 text-sm">{account.type}</td>
                                    <td className="py-3 px-4 text-sm">{account.currency}</td>
                                    <td className="py-3 px-4 text-sm text-right font-semibold">
                                      {account.balance.toLocaleString("ru-RU")}
                                    </td>
                                    <td className="py-3 px-4 text-sm">{account.owner || "-"}</td>
                                  </tr>
                                ))}
                              </tbody>
                            </table>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                ) : (
                  <p className="text-center text-muted-foreground py-8">Счетов пока нет</p>
                )}
              </CardContent>
            </Card>
          </motion.div>
        )}

        {/* Пустое состояние */}
        {connectedBanksList.length === 0 && !accountsLoading && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-12"
          >
            <Building2 className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
            <h3 className="text-xl font-semibold mb-2">Нет подключенных банков</h3>
            <p className="text-muted-foreground mb-4">
              Нажмите "Подключить банк" выше, чтобы начать
            </p>
          </motion.div>
        )}
      </div>
    </div>
  )
}
