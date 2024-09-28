import { BrowserRouter, Route, Routes } from "react-router-dom";
import Landing from "./landing/Landing";
import { LandingHeader, LoginHeader, RegisterHeader } from "./components/app/header/Header";
import Footer from "@/components/app/footer/Footer";
import RegisterForm from "./forms/RegisterForm";
import LoginForm from "./forms/LoginForm";

function Layout({ children }: { children: React.ReactNode }) {
    return (
        <div className="max-w-7xl m-auto backgr bg-white">
            {children}
        </div>
    );
}

export default function AppRoutes() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<LandingPage />} />
                <Route path="/register" element={<RegisterPage />} />
                <Route path="/login" element={<LoginPage />} />
            </Routes>
            <Layout>
                <Footer />
            </Layout>
        </BrowserRouter>
    );
}

function LandingPage() {

    return (<>
        <LandingHeader />
        <Landing />
    </>);
}

function RegisterPage() {
    return (<>
        <RegisterHeader />
        <Layout>
            <RegisterForm />
        </Layout>
    </>);
}

function LoginPage() {
    return (<>
        <LoginHeader />
        <Layout>
            <LoginForm />
        </Layout>
    </>);
}
