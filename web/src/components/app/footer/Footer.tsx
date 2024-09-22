import { Button } from "@/components/ui/button"

export default function Footer() {
    return (
        <footer className="bg-gray-100 py-8 px-4 md:px-6 lg:px-8">
            <div className="max-w-6xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-8">
                <div>
                    <h3 className="font-semibold mb-4">About Us</h3>
                    <p className="text-sm text-muted-foreground">fixup connects you with trusted professionals for all your home service needs.</p>
                </div>
                <div>
                    <h3 className="font-semibold mb-4">Contact Us</h3>
                    <p className="text-sm text-muted-foreground">Email: support@fixup.com</p>
                    <p className="text-sm text-muted-foreground">Phone: (123) 456-7890</p>
                </div>
                <div>
                    <h3 className="font-semibold mb-4">For Service Providers</h3>
                    <Button variant="link" className="p-0 h-auto text-sm text-muted-foreground hover:text-primary">
                        Register as a Provider
                    </Button>
                </div>
            </div>
        </footer>
    )
}