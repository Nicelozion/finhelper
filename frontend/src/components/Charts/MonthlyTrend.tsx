import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

interface MonthlyData {
  month: string
  income: number
  expenses: number
}

interface MonthlyTrendProps {
  data: MonthlyData[]
}

export function MonthlyTrend({ data }: MonthlyTrendProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Динамика доходов и расходов</CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="month" />
            <YAxis />
            <Tooltip formatter={(value: number) => `${value.toLocaleString("ru-RU")} ₽`} />
            <Legend />
            <Line type="monotone" dataKey="income" stroke="#10b981" strokeWidth={2} name="Доходы" />
            <Line type="monotone" dataKey="expenses" stroke="#ef4444" strokeWidth={2} name="Расходы" />
          </LineChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  )
}

