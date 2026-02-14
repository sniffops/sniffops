import { useState, useEffect } from 'react'
import { Search } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Separator } from '@/components/ui/separator'
import { SidebarTrigger } from '@/components/ui/sidebar'
import { Input } from '@/components/ui/input'
import { ThemeToggle } from './ThemeToggle'

export function Header() {
  const [offset, setOffset] = useState(0)

  useEffect(() => {
    const onScroll = () => {
      setOffset(document.body.scrollTop || document.documentElement.scrollTop)
    }

    document.addEventListener('scroll', onScroll, { passive: true })
    return () => document.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <header
      className={cn(
        'sticky top-0 z-50 h-16 transition-shadow',
        offset > 10 ? 'shadow-sm' : 'shadow-none'
      )}
    >
      <div
        className={cn(
          'relative flex h-full items-center gap-3 border-b bg-background/95 p-4 backdrop-blur supports-[backdrop-filter]:bg-background/60 sm:gap-4'
        )}
      >
        <SidebarTrigger variant="outline" />
        <Separator orientation="vertical" className="h-6" />
        
        <div className="flex flex-1 items-center gap-4">
          <form className="flex-1 max-w-md">
            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                type="search"
                placeholder="Search traces... (⌘K)"
                className="pl-8 sm:w-[300px] md:w-[400px] lg:w-[500px]"
              />
            </div>
          </form>

          <div className="ml-auto flex items-center gap-2">
            <ThemeToggle />
          </div>
        </div>
      </div>
    </header>
  )
}
