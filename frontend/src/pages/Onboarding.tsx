import { useState } from "react"
import { useMutation } from "@tanstack/react-query"
import { useNavigate } from "react-router-dom"
import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { api } from "@/lib/api"
import { Building2, Wallet } from "lucide-react"

const banks = [
  { code: "ABank", name: "ABank", color: "bg-blue-500" },
  { code: "VBank", name: "VBank", color: "bg-green-500" },
  { code: "SBank", name: "SBank", color: "bg-purple-500" },
]

export function Onboarding() {
  const navigate = useNavigate()
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [selectedBank, setSelectedBank] = useState<string | null>(null)

  const connectMutation = useMutation({
    mutationFn: (bank: string) => api.connectBank(bank),
    onSuccess: () => {
      setIsDialogOpen(false)
      navigate("/dashboard")
    },
  })

  const handleConnect = (bank: string) => {
    setSelectedBank(bank)
    connectMutation.mutate(bank)
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary/10 via-background to-background p-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="max-w-2xl w-full"
      >
        <Card className="text-center">
          <CardHeader className="space-y-4">
            <div className="mx-auto h-16 w-16 rounded-full bg-primary/10 flex items-center justify-center">
              <Wallet className="h-8 w-8 text-primary" />
            </div>
            <CardTitle className="text-3xl">Добро пожаловать в FinHelper</CardTitle>
            <CardDescription className="text-lg">
              Агрегируйте данные со всех ваших банков в одном месте
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <p className="text-muted-foreground">
              Начните с подключения вашего первого банка
            </p>
            <Button
              size="lg"
              onClick={() => setIsDialogOpen(true)}
              className="w-full md:w-auto"
            >
              Подключить банк
            </Button>
          </CardContent>
        </Card>
      </motion.div>

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Выберите банк</DialogTitle>
            <DialogDescription>
              Выберите банк для подключения через OAuth2
            </DialogDescription>
          </DialogHeader>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
            {banks.map((bank) => (
              <motion.button
                key={bank.code}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={() => handleConnect(bank.code)}
                disabled={connectMutation.isPending}
                className={`${bank.color} text-white p-6 rounded-lg flex flex-col items-center gap-2 hover:opacity-90 transition-opacity disabled:opacity-50`}
              >
                <Building2 className="h-8 w-8" />
                <span className="font-semibold">{bank.name}</span>
              </motion.button>
            ))}
          </div>
          {connectMutation.isError && (
            <div className="mt-4 p-3 bg-destructive/10 text-destructive rounded-md text-sm">
              Ошибка подключения. Попробуйте ещё раз.
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}

