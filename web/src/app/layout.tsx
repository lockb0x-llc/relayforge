import type { Metadata } from "next";
import "./globals.css";
import { AuthProvider } from "@/contexts/AuthContext";
import Navigation from "@/components/Navigation";

export const metadata: Metadata = {
  title: "RelayForge - Infrastructure Orchestration Platform",
  description: "Decentralized infrastructure orchestration with workflow automation",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="font-sans">
        <AuthProvider>
          <div className="min-h-screen bg-gray-50">
            <Navigation />
            <main className="container mx-auto px-4 py-8">
              {children}
            </main>
          </div>
        </AuthProvider>
      </body>
    </html>
  );
}
