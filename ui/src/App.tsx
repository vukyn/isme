import "./App.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Login } from "./pages/Login";
import { SSOLogin } from "./pages/SSOLogin";
import { Signup } from "./pages/Signup";
import { Welcome } from "./pages/Welcome";
import { ForgotPassword } from "./pages/ForgotPassword";
import { Sessions } from "./pages/Sessions";
import { Team } from "./pages/Team";
import { Settings } from "./pages/Settings";
import { NotFound } from "./pages/NotFound";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { AuroraBackground } from "./components/ui/aurora-background";

function App() {
	return (
		<>
			<AuroraBackground />
			<BrowserRouter>
				<Routes>
					<Route path="/login" element={<Login />} />
					<Route path="/sso/login" element={<SSOLogin />} />
					<Route path="/signup" element={<Signup />} />
					<Route path="/forgot-password" element={<ForgotPassword />} />
					<Route
						path="/welcome"
						element={
							<ProtectedRoute>
								<Welcome />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/sessions"
						element={
							<ProtectedRoute>
								<Sessions />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/team"
						element={
							<ProtectedRoute>
								<Team />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/settings"
						element={
							<ProtectedRoute>
								<Settings />
							</ProtectedRoute>
						}
					/>
					<Route path="/" element={<Navigate to="/welcome" replace />} />
					<Route path="/404" element={<NotFound />} />
					<Route path="*" element={<Navigate to="/404" replace />} />
				</Routes>
			</BrowserRouter>
		</>
	);
}

export default App;
