import { cn } from "@/lib/utils"

export function LoadingSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("animate-pulse rounded-md bg-muted", className)} />
  )
}

export function CardSkeleton() {
  return (
    <div className="rounded-lg border bg-card p-6">
      <div className="space-y-3">
        <LoadingSkeleton className="h-4 w-1/4" />
        <LoadingSkeleton className="h-8 w-1/2" />
        <LoadingSkeleton className="h-4 w-1/3" />
      </div>
    </div>
  )
}

export function TableSkeleton() {
  return (
    <div className="space-y-3">
      {[...Array(5)].map((_, i) => (
        <div key={i} className="flex items-center space-x-4">
          <LoadingSkeleton className="h-12 flex-1" />
          <LoadingSkeleton className="h-12 w-24" />
          <LoadingSkeleton className="h-12 w-32" />
          <LoadingSkeleton className="h-12 w-20" />
        </div>
      ))}
    </div>
  )
}

