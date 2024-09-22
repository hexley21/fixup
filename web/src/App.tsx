"use client"

import Landing from "./landing/Landing"
import { LandingHeader } from "./components/app/header/Header";
import Footer from "@/components/app/footer/Footer"


export default function App() {
  return (
    <div className="App">
      <LandingHeader />
      <div className="max-w-7xl m-auto backgr bg-white">
        <Landing />
        <Footer />
      </div>
    </div>
  );
}
