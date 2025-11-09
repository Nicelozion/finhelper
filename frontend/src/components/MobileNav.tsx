import { useState } from "react"
import { Link, useLocation } from "react-router-dom"
import { LayoutDashboard, Receipt, BarChart3, Settings, Menu, X } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

const navItems = [
  { path: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { path: "/transactions", label: "Transactions", icon: Receipt },
  { path: "/analytics", label: "Analytics", icon: BarChart3 },
  { path: "/settings", label: "Settings", icon: Settings },
]

export function MobileNav() {
  const [isOpen, setIsOpen] = useState(false)
  const location = useLocation()

  return (
    <div className="md:hidden fixed top-0 left-0 right-0 z-50 bg-background border-b">
      <div className="flex items-center justify-between p-4">
        <span className="text-lg font-bold">FinHelper</span>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setIsOpen(!isOpen)}
        >
          {isOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </Button>
      </div>
      {isOpen && (
        <nav className="border-t bg-background">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = location.pathname === item.path
            return (
              <Link
                key={item.path}
                to={item.path}
                onClick={() => setIsOpen(false)}
                className={cn(
                  "flex items-center gap-3 px-4 py-3 border-b transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-accent"
                )}
              >
                <Icon className="h-5 w-5" />
                {item.label}
              </Link>
            )
          })}
        </nav>
      )}
    </div>
  )
}

