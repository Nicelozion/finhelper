import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Building2 } from "lucide-react"
import { cn } from "@/lib/utils"

interface BankCardProps {
  bank: string
  balance: number
  currency: string
  accountsCount: number
  className?: string
}

const bankColors: Record<string, string> = {
  ABank: "bg-blue-500",
  VBank: "bg-green-500",
  SBank: "bg-purple-500",
}

export function BankCard({ bank, balance, currency, accountsCount, className }: BankCardProps) {
  return (
    <Card className={cn("hover:shadow-lg transition-shadow", className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{bank}</CardTitle>
        <div className={cn("h-8 w-8 rounded-full flex items-center justify-center", bankColors[bank] || "bg-gray-500")}>
          <Building2 className="h-4 w-4 text-white" />
        </div>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">
          {balance.toLocaleString("ru-RU")} {currency}
        </div>
        <p className="text-xs text-muted-foreground mt-1">
          {accountsCount} {accountsCount === 1 ? "счёт" : accountsCount < 5 ? "счёта" : "счетов"}
        </p>
      </CardContent>
    </Card>
  )
}

