import { ReactNode } from 'react'
import { Home } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Link } from "react-router-dom";
import { Globe } from 'lucide-react'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'

interface HeaderProps {
    children?: ReactNode
}

function Header({ children }: HeaderProps) {
    return (
        <header className="bg-background border-b shadow-sm">
            <div className="m-auto max-w-7xl px-4 py-3 flex items-center justify-between">
                <div className="flex items-center space-x-4">
                    <Home className="h-8 w-8 text-primary" />
                    <Link to="/" className="text-xl font-bold">fixup.com</Link>
                </div>
                <div className="flex items-center space-x-4">
                    {children}
                </div>
            </div>
        </header>
    )
}

function SignIn() {
    return (
        <Link to="/login">
            <Button variant="outline">Sign In</Button>
        </Link>
    )
}

function SignUp() {
    return (
        <Link to="/register">
            <Button variant="outline">Sign Up</Button>
        </Link>
    )
}

function LanguageSwitch() {
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

export function LandingHeader() {
    return (
        <Header>
            <LanguageSwitch />
            <SignIn />
        </Header>
    )
}

export function RegisterHeader() {
    return (
        <Header>
            <LanguageSwitch />
            <SignIn />
        </Header>
    )
}

export function LoginHeader() {
    return (
        <Header>
            <LanguageSwitch />
            <SignUp />
        </Header>
    )
}
