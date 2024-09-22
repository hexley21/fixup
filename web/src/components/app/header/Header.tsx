import { ReactNode } from 'react'
import { Home } from 'lucide-react'
import LanguageSwitch from './Language'
import SignIn from './SignIn'

interface HeaderProps {
    children?: ReactNode
}

function Header({ children }: HeaderProps) {
    return (
        <header className="bg-background border-b shadow-sm">
            <div className="m-auto max-w-7xl px-4 py-3 flex items-center justify-between">
                <div className="flex items-center space-x-4">
                    <Home className="h-8 w-8 text-primary" />
                    <a href="/" className="text-xl font-bold">fixup.com</a>
                </div>
                <div className="flex items-center space-x-4">
                    {children}
                </div>
            </div>
        </header>
    )
}

export function LandingHeader() {
    return (
        <Header>
            <LanguageSwitch/>
            <SignIn />
        </Header>
    )
}