import { useState } from "react"
import { useMutation } from "@tanstack/react-query"
import { useNavigate } from "react-router-dom"
import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { api } from "@/lib/api"
import { Check, X } from "lucide-react"

export function Subscription() {
  const navigate = useNavigate()
  const [selectedPlan, setSelectedPlan] = useState<"Free" | "Pro">("Pro")

  const subscribeMutation = useMutation({
    mutationFn: () => api.subscribe("Pro"),
    onSuccess: () => {
      navigate("/settings?subscribed=true")
    },
  })

  const plans = [
    {
      name: "Free",
      price: "0",
      features: [
        "Просмотр счетов",
        "Просмотр транзакций",
        "Базовая аналитика",
      ],
      limitations: [
        "Без PDF отчётов",
        "Без прогнозов",
        "Без уведомлений",
      ],
    },
    {
      name: "Pro",
      price: "499",
      period: "месяц",
      features: [
        "Всё из Free",
        "PDF отчёты",
        "Прогнозы расходов",
        "Уведомления",
        "Приоритетная поддержка",
      ],
      popular: true,
    },
  ]

  return (
    <div className="md:ml-64 pt-16 md:pt-0 p-4 md:p-8">
      <div className="max-w-5xl mx-auto space-y-8">
        <div className="text-center">
          <h1 className="text-3xl font-bold">Выберите план</h1>
          <p className="text-muted-foreground mt-2">Выберите подходящий тариф для вас</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {plans.map((plan, index) => (
            <motion.div
              key={plan.name}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
            >
              <Card className={plan.popular ? "border-primary border-2 relative" : ""}>
                {plan.popular && (
                  <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                    <span className="bg-primary text-primary-foreground px-3 py-1 rounded-full text-xs font-semibold">
                      Популярный
                    </span>
                  </div>
                )}
                <CardHeader>
                  <CardTitle className="text-2xl">{plan.name}</CardTitle>
                  <CardDescription>
                    <div className="mt-4">
                      <span className="text-4xl font-bold">{plan.price}</span>
                      {plan.period && (
                        <span className="text-muted-foreground"> ₽/{plan.period}</span>
                      )}
                    </div>
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <ul className="space-y-3">
                    {plan.features.map((feature) => (
                      <li key={feature} className="flex items-center gap-2">
                        <Check className="h-4 w-4 text-green-600" />
                        <span>{feature}</span>
                      </li>
                    ))}
                    {plan.limitations?.map((limitation) => (
                      <li key={limitation} className="flex items-center gap-2 text-muted-foreground">
                        <X className="h-4 w-4" />
                        <span>{limitation}</span>
                      </li>
                    ))}
                  </ul>
                </CardContent>
                <CardFooter>
                  {plan.name === "Pro" ? (
                    <Button
                      className="w-full"
                      onClick={() => subscribeMutation.mutate()}
                      disabled={subscribeMutation.isPending}
                    >
                      {subscribeMutation.isPending ? "Оформление..." : "Оформить подписку"}
                    </Button>
                  ) : (
                    <Button variant="outline" className="w-full" disabled>
                      Текущий план
                    </Button>
                  )}
                </CardFooter>
              </Card>
            </motion.div>
          ))}
        </div>
      </div>
    </div>
  )
}

