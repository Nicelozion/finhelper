import { useState } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { api } from "@/lib/api"
import { User, CreditCard, Moon, Sun, RefreshCw } from "lucide-react"
import { useNavigate } from "react-router-dom"

export function Settings() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [theme, setTheme] = useState<"light" | "dark">("light")

  const { data: profile, isLoading } = useQuery({
    queryKey: ["profile"],
    queryFn: api.getUserProfile,
  })

  const toggleTheme = () => {
    const newTheme = theme === "light" ? "dark" : "light"
    setTheme(newTheme)
    document.documentElement.classList.toggle("dark", newTheme === "dark")
  }

  const refreshData = () => {
    queryClient.invalidateQueries()
  }

  return (
    <div className="md:ml-64 pt-16 md:pt-0 p-4 md:p-8">
      <div className="max-w-4xl mx-auto space-y-8">
        <div>
          <h1 className="text-3xl font-bold">Настройки</h1>
          <p className="text-muted-foreground mt-2">Управление аккаунтом и подпиской</p>
        </div>

        {/* Профиль */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <User className="h-5 w-5" />
                Профиль
              </CardTitle>
              <CardDescription>Информация о вашем аккаунте</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {isLoading ? (
                <div className="space-y-2">
                  <div className="h-4 bg-muted rounded w-1/4 animate-pulse" />
                  <div className="h-4 bg-muted rounded w-1/3 animate-pulse" />
                </div>
              ) : (
                <>
                  <div>
                    <label className="text-sm font-medium">Имя</label>
                    <p className="text-lg">{profile?.name || "Не указано"}</p>
                  </div>
                  <div>
                    <label className="text-sm font-medium">Email</label>
                    <p className="text-lg">{profile?.email || "Не указано"}</p>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </motion.div>

        {/* Подключенные банки */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
        >
          <Card>
            <CardHeader>
              <CardTitle>Подключённые банки</CardTitle>
              <CardDescription>Управление подключениями к банкам</CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="space-y-2">
                  <div className="h-10 bg-muted rounded animate-pulse" />
                  <div className="h-10 bg-muted rounded animate-pulse" />
                </div>
              ) : (
                <div className="space-y-2">
                  {profile?.connectedBanks.map((bank) => (
                    <div key={bank} className="flex items-center justify-between p-3 border rounded-lg">
                      <span className="font-medium">{bank}</span>
                      <Button variant="outline" size="sm">Отключить</Button>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </motion.div>

        {/* Настройки */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
        >
          <Card>
            <CardHeader>
              <CardTitle>Настройки приложения</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Тема</p>
                  <p className="text-sm text-muted-foreground">Переключение между светлой и тёмной темой</p>
                </div>
                <Button variant="outline" onClick={toggleTheme} className="gap-2">
                  {theme === "light" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
                  {theme === "light" ? "Светлая" : "Тёмная"}
                </Button>
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Обновить данные</p>
                  <p className="text-sm text-muted-foreground">Синхронизировать данные со всеми банками</p>
                </div>
                <Button variant="outline" onClick={refreshData} className="gap-2">
                  <RefreshCw className="h-4 w-4" />
                  Обновить
                </Button>
              </div>
            </CardContent>
          </Card>
        </motion.div>

        {/* Подписка */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
        >
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <CreditCard className="h-5 w-5" />
                Подписка
              </CardTitle>
              <CardDescription>Управление подпиской</CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="h-20 bg-muted rounded animate-pulse" />
              ) : (
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-lg font-semibold">
                        Текущий план: {profile?.subscription || "Free"}
                      </p>
                      {profile?.subscription === "Free" && (
                        <p className="text-sm text-muted-foreground mt-1">
                          Базовые функции: просмотр счетов и транзакций
                        </p>
                      )}
                      {profile?.subscription === "Pro" && (
                        <p className="text-sm text-muted-foreground mt-1">
                          Премиум функции: PDF отчёты, прогнозы, уведомления
                        </p>
                      )}
                    </div>
                  </div>
                  {profile?.subscription === "Free" && (
                    <Button onClick={() => navigate("/subscription")} className="w-full">
                      Перейти на Pro
                    </Button>
                  )}
                  {profile?.subscription === "Pro" && (
                    <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
                      <p className="text-sm text-green-800 dark:text-green-200">
                        ✅ У вас активна Pro подписка
                      </p>
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </motion.div>
      </div>
    </div>
  )
}

