import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import { ChevronDown, Home, Paintbrush, Wrench, Car, Leaf, Laptop, Shirt, Dog, Book } from "lucide-react"
import { LandingHeader } from "../common/Header"
import { DownloadMobile } from "./DownloadMobile"
import { ContentLayout } from "../common/ContentLayout"

export function Landing() {
    return (<>
        <LandingHeader />
        <ContentLayout>
            <Component />
        </ContentLayout>
    </>);
}

const categories = [
    { icon: <Home className="h-8 w-8" />, name: "Renovation", subcategories: ["• Kitchen", "• Bathroom", "• Living Room", "• Bedroom", "• Outdoor", "• Basement", "• Attic"] },
    { icon: <Paintbrush className="h-8 w-8" />, name: "Cleaning", subcategories: ["• Deep Clean", "• Regular Clean", "• Window Cleaning", "• Carpet Cleaning", "• Post-Construction", "• Office Cleaning", "• Move-in/Move-out"] },
    { icon: <Wrench className="h-8 w-8" />, name: "Repairs", subcategories: ["• Plumbing", "• Electrical", "• Appliance Repair", "• Furniture Assembly", "• Painting", "• Drywall Repair", "• Flooring"] },
    { icon: <Car className="h-8 w-8" />, name: "Automotive", subcategories: ["• Oil Change", "• Tire Service", "• Brake Repair", "• Detailing", "• Battery Replacement", "• Engine Tune-up"] },
    { icon: <Leaf className="h-8 w-8" />, name: "Landscaping", subcategories: ["• Lawn Mowing", "• Tree Trimming", "• Garden Design", "• Irrigation", "• Weed Control", "• Hardscaping"] },
    { icon: <Laptop className="h-8 w-8" />, name: "Tech Support", subcategories: ["• Computer Repair", "• Network Setup", "• Virus Removal", "• Data Recovery", "• Smart Home Installation"] },
    { icon: <Shirt className="h-8 w-8" />, name: "Laundry", subcategories: ["• Wash & Fold", "• Dry Cleaning", "• Alterations", "• Shoe Repair", "• Leather Care"] },
    { icon: <Dog className="h-8 w-8" />, name: "Pet Care", subcategories: ["• Dog Walking", "• Pet Sitting", "• Grooming", "• Training", "• Veterinary Services"] },
    { icon: <Book className="h-8 w-8" />, name: "Tutoring", subcategories: ["• Math", "• Science", "• Language", "• Test Prep", "• Music Lessons", "• Art Classes"] },
]

const userReviews = [
    { name: "Alice", avatar: "A", rating: 5, comment: "Excellent service! The renovation team was professional and efficient.", time: "2023-07-15T14:30:00", service: "Renovation", provider: "Home Makeover Inc." },
    { name: "Bob", avatar: "B", rating: 4, comment: "Great cleaning service. My house has never looked better.", time: "2023-07-14T10:15:00", service: "Cleaning", provider: "Sparkle Clean Co." },
    { name: "Charlie", avatar: "C", rating: 5, comment: "The tech support was incredibly helpful. Fixed my computer issues in no time.", time: "2023-07-13T16:45:00", service: "Tech Support", provider: "Geek Squad" },
    { name: "Diana", avatar: "D", rating: 4, comment: "Reliable pet care services. My dog loves the walker!", time: "2023-07-12T09:00:00", service: "Pet Care", provider: "Happy Paws" },
    { name: "Ethan", avatar: "E", rating: 5, comment: "The tutoring service helped my daughter improve her grades significantly.", time: "2023-07-11T17:30:00", service: "Tutoring", provider: "Bright Minds Tutoring" },
]

function Component() {
    return (
        <div className="flex flex-col min-h-screen bg-white ">
            <main>
                <section className="py-12 px-4">
                    <div className="max-w-4xl mx-auto">
                        <h1 className="font-semibold text-3xl xs:text-4xl sm:text-5xl text-muted-foreground mb-8">Your solution for all services.</h1>
                        <div className="flex w-full h-full h-12 items-center space-x-2">
                            <Input type="text" placeholder="Search for a service..." className="p-4 text-lg w-full h-full" />
                            <Button type="submit" className="p-4 text-lg h-full">
                                Search
                            </Button>
                        </div>
                    </div>
                </section>

                <Separator />

                <section className="py-12 px-4">
                    <div className="max-w-6xl mx-auto">
                        <h2 className="text-4xl font-semibold mb-8 text-center">Our Services</h2>
                        <div className="grid grid-cols-1 xs:grid-cols-2 md:grid-cols-3 justify-items-center">
                            {categories.map((category, index) => (
                                <div key={index} className="flex flex-col max-w-sm h-auto p-6 rounded-lg">
                                    <div className="flex items-start space-x-4">
                                        <div>{category.icon}</div>
                                        <div className="flex flex-col text-left">
                                            <h3 className="text-2xl font-semibold mb-2">{category.name}</h3>
                                            <ul className="space-y-2">
                                                {category.subcategories.slice(0, 5).map((sub, subIndex) => (
                                                    <li key={subIndex} className="text-muted-foreground text-base text-black-400">{sub}</li>
                                                ))}
                                            </ul>
                                            <Button variant="link" className="justify-start text-gray-400 hover:text-gray-700 hover:underline mt-4 p-0">
                                                See more <ChevronDown className="p-1" />
                                            </Button>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                    <div className="flex justify-center pt-4">
                        <Button variant="ghost" className="underline text-muted-foreground">
                            All categories <ChevronDown className="p-1" />
                        </Button>
                    </div>
                </section>

                <Separator />

                <section className="py-12 px-4 md:px-6 lg:px-8">
                    <div className="max-w-4xl mx-auto">
                        <h2 className="text-3xl font-semibold mb-8 text-center">What Our Users Say</h2>
                        <ScrollArea className="h-[400px] w-full rounded-md border px-4">
                            {userReviews.map((review, index) => (
                                <div key={index} className="my-4">
                                    <div className="flex items-center space-x-4 mb-2">
                                        <Avatar>
                                            <AvatarImage src={`https://api.dicebear.com/6.x/initials/svg?seed=${review.name}`} />
                                            <AvatarFallback>{review.avatar}</AvatarFallback>
                                        </Avatar>
                                        <div>
                                            <p className="font-semibold">{review.name}</p>
                                            <div className="flex items-center">
                                                {[...Array(5)].map((_, i) => (
                                                    <svg key={i} className={`w-4 h-4 ${i < review.rating ? 'text-yellow-300' : 'text-gray-300'}`} aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 22 20">
                                                        <path d="M20.924 7.625a1.523 1.523 0 0 0-1.238-1.044l-5.051-.734-2.259-4.577a1.534 1.534 0 0 0-2.752 0L7.365 5.847l-5.051.734A1.535 1.535 0 0 0 1.463 9.2l3.656 3.563-.863 5.031a1.532 1.532 0 0 0 2.226 1.616L11 17.033l4.518 2.375a1.534 1.534 0 0 0 2.226-1.617l-.863-5.03L20.537 9.2a1.523 1.523 0 0 0 .387-1.575Z" />
                                                    </svg>
                                                ))}
                                            </div>
                                        </div>
                                    </div>
                                    <div className="text-sm text-muted-foreground">
                                        <p>Service: {review.service} - {review.provider}</p>
                                        <p>Date: {new Date(review.time).toLocaleString()}</p>
                                    </div>
                                    <p className="text-black-400 my-2">{review.comment}</p>
                                </div>
                            ))}
                        </ScrollArea>
                    </div>
                </section>

                <Separator />

                <section className="flex justify-center m-8">
                    <DownloadMobile />
                </section>
            </main>
        </div>
    )
}
