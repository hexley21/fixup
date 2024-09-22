import { Globe } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export default function LanguageSwitch() {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm">
          <Globe className="h-4 w-4 mr-2" />
          EN
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem>English</DropdownMenuItem>
        <DropdownMenuItem>Español</DropdownMenuItem>
        <DropdownMenuItem>Français</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}