import { Link, useLocation } from "react-router-dom"
import { LayoutDashboard, Receipt, BarChart3, Settings, Wallet } from "lucide-react"
import { cn } from "@/lib/utils"

const navItems = [
  { path: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { path: "/transactions", label: "Transactions", icon: Receipt },
  { path: "/analytics", label: "Analytics", icon: BarChart3 },
  { path: "/settings", label: "Settings", icon: Settings },
]

export function Sidebar() {
  const location = useLocation()

  return (
    <aside className="fixed left-0 top-0 z-40 h-screen w-64 border-r bg-background hidden md:block">
      <div className="flex h-full flex-col">
        <div className="flex h-16 items-center border-b px-6">
          <Wallet className="mr-2 h-6 w-6 text-primary" />
          <span className="text-xl font-bold">FinHelper</span>
        </div>
        <nav className="flex-1 space-y-1 p-4">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = location.pathname === item.path
            return (
              <Link
                key={item.path}
                to={item.path}
                className={cn(
                  "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                )}
              >
                <Icon className="h-5 w-5" />
                {item.label}
              </Link>
            )
          })}
        </nav>
      </div>
    </aside>
  )
}

