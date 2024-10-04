"use client"
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { Landing } from "./components/app/landing/Landing";
import { Footer } from "@/components/app/common/Footer";
import { Register } from "./components/app/forms/Register";
import { Login } from "./components/app/forms/Login";
import { ContentLayout } from "./components/app/common/ContentLayout";

export default function App() {
  return (
    <div className="App">
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Landing />} />
          <Route path="/register" element={<Register />} />
          <Route path="/login" element={<Login />} />
          {/* <Route path="/profile/:id" element={<Profile />} /> */}
        </Routes>
        <ContentLayout>
          <Footer />
        </ContentLayout>
      </BrowserRouter>
    </div>
  );
}
