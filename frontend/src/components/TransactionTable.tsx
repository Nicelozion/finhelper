import { format } from "date-fns"
import { ru } from "date-fns/locale/ru"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { Transaction } from "@/lib/api"
import { cn } from "@/lib/utils"

interface TransactionTableProps {
  transactions: Transaction[]
  showBank?: boolean
}

export function TransactionTable({ transactions, showBank = true }: TransactionTableProps) {
  if (transactions.length === 0) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-center text-muted-foreground">Нет транзакций</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Последние транзакции</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b">
                <th className="text-left py-2 px-4 text-sm font-medium text-muted-foreground">Дата</th>
                <th className="text-left py-2 px-4 text-sm font-medium text-muted-foreground">Описание</th>
                <th className="text-left py-2 px-4 text-sm font-medium text-muted-foreground">Категория</th>
                {showBank && (
                  <th className="text-left py-2 px-4 text-sm font-medium text-muted-foreground">Банк</th>
                )}
                <th className="text-right py-2 px-4 text-sm font-medium text-muted-foreground">Сумма</th>
              </tr>
            </thead>
            <tbody>
              {transactions.map((tx) => (
                <tr key={tx.id} className="border-b hover:bg-muted/50">
                  <td className="py-3 px-4 text-sm">
                    {format(new Date(tx.date), "dd.MM.yyyy", { locale: ru })}
                  </td>
                  <td className="py-3 px-4 text-sm">{tx.merchant || tx.description}</td>
                  <td className="py-3 px-4 text-sm">{tx.category}</td>
                  {showBank && (
                    <td className="py-3 px-4 text-sm">{tx.bank}</td>
                  )}
                  <td className={cn(
                    "py-3 px-4 text-sm text-right font-medium",
                    tx.amount < 0 ? "text-destructive" : "text-green-600"
                  )}>
                    {tx.amount > 0 ? "+" : ""}
                    {tx.amount.toLocaleString("ru-RU")} {tx.currency}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  )
}

