import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ExpensesDonut } from "@/components/Charts/ExpensesDonut"
import { MonthlyTrend } from "@/components/Charts/MonthlyTrend"
import { mockCategories, mockMonthlyData } from "@/lib/mockData"
import { TrendingUp, TrendingDown, Lightbulb } from "lucide-react"

export function Analytics() {
  const totalExpenses = mockCategories.reduce((sum, cat) => sum + cat.amount, 0)
  const projectedRemaining = 70000 - totalExpenses * 1.2 // Примерный прогноз

  return (
    <div className="md:ml-64 pt-16 md:pt-0 p-4 md:p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        <div>
          <h1 className="text-3xl font-bold">Аналитика</h1>
          <p className="text-muted-foreground mt-2">Анализ ваших финансов</p>
        </div>

        {/* Прогнозы */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
          >
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <TrendingUp className="h-5 w-5 text-green-600" />
                  Прогноз на конец месяца
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold text-green-600">
                  {projectedRemaining.toLocaleString("ru-RU")} ₽
                </div>
                <p className="text-sm text-muted-foreground mt-2">
                  Если траты сохранятся, к концу месяца останется
                </p>
              </CardContent>
            </Card>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <TrendingDown className="h-5 w-5 text-red-600" />
                  Средние траты в день
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold">
                  {(totalExpenses / 30).toLocaleString("ru-RU")} ₽
                </div>
                <p className="text-sm text-muted-foreground mt-2">
                  За последние 30 дней
                </p>
              </CardContent>
            </Card>
          </motion.div>
        </div>

        {/* Графики */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.2 }}
          >
            <MonthlyTrend data={mockMonthlyData} />
          </motion.div>
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.3 }}
          >
            <ExpensesDonut data={mockCategories} />
          </motion.div>
        </div>

        {/* Советы */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
        >
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Lightbulb className="h-5 w-5 text-yellow-500" />
                Рекомендации
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="p-4 border rounded-lg">
                  <h4 className="font-semibold mb-2">Попробуйте ограничить траты на еду</h4>
                  <p className="text-sm text-muted-foreground">
                    Вы тратите {mockCategories.find(c => c.name === "Еда")?.amount.toLocaleString("ru-RU")} ₽ на еду в месяц. 
                    Попробуйте готовить дома чаще.
                  </p>
                </div>
                <div className="p-4 border rounded-lg">
                  <h4 className="font-semibold mb-2">Экономьте на транспорте</h4>
                  <p className="text-sm text-muted-foreground">
                    Рассмотрите возможность использования общественного транспорта вместо такси.
                  </p>
                </div>
                <div className="p-4 border rounded-lg">
                  <h4 className="font-semibold mb-2">Создайте резервный фонд</h4>
                  <p className="text-sm text-muted-foreground">
                    Откладывайте 10% от дохода каждый месяц для создания финансовой подушки.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      </div>
    </div>
  )
}

