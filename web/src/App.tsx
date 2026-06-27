import { Routes, Route, Navigate, BrowserRouter } from "react-router-dom";
import { AuthProvider, ProtectedRoute } from "./contexts/AuthContext";
import LoginPage from "./pages/Login";
import DashboardPage from "./pages/Dashboard";
import DevicesPage from "./pages/Devices";
import AddressBookPage from "./pages/AddressBook";
import LogsPage from "./pages/Logs";
import SettingsPage from "./pages/Settings";
import Sidebar from "./components/Sidebar";
import Header from "./components/Header";

function Layout() {
  return (
    <div className="flex h-screen overflow-hidden" style={{ fontFamily: "Inter, system-ui, sans-serif", backgroundColor: "#F7F9FB" }}>
      <Sidebar />
      <div className="flex-1 flex flex-col overflow-hidden">
        <Header />
        <main className="flex-1 overflow-auto">
          <Routes>
            <Route path="/dashboard" element={<DashboardPage />} />
            <Route path="/devices" element={<DevicesPage />} />
            <Route path="/address-book" element={<AddressBookPage />} />
            <Route path="/logs" element={<LogsPage />} />
            <Route path="/settings" element={<SettingsPage />} />
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </main>
      </div>
    </div>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/*"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
