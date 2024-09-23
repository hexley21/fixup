"use client"

import Landing from "./landing/Landing"
import { LandingHeader } from "./components/app/header/Header";
import Footer from "@/components/app/footer/Footer"
import RegisternForm from "./register/RegisterForm";


export default function App() {
  return (
    <div className="App">
      <LandingHeader />
      <div className="max-w-7xl m-auto backgr bg-white">
        <RegisternForm />
        <Footer />
      </div>
    </div>
  );
}
