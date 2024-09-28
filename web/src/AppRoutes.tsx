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
                <Route path="/" element={
                    <div>
                        <LandingHeader />
                        <Layout>
                            <Landing />
                        </Layout>
                    </div>}>
                </Route>
                <Route path="/register" element={
                    <div>
                        <RegisterHeader />
                        <Layout>
                            <RegisterForm />
                        </Layout>
                    </div>}>
                </Route>
                <Route path="/login" element={
                    <div>
                        <LoginHeader />
                        <Layout>
                            <LoginForm />
                        </Layout>
                    </div>}>
                </Route>
            </Routes>
            <Layout>
                <Footer />
            </Layout>
        </BrowserRouter>
    );
};
