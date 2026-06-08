import "./App.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Login } from "./pages/Login";
import { SSOLogin } from "./pages/SSOLogin";
import { AcceptInvite } from "./pages/AcceptInvite";
import { Welcome } from "./pages/Welcome";
import { Sessions } from "./pages/Sessions";
import { Team } from "./pages/Team";
import { Users } from "./pages/Users";
import { Roles } from "./pages/Roles";
import { AppServices } from "./pages/AppServices";
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
					{/* public — invite links are the only way to create an account */}
					<Route path="/accept-invite" element={<AcceptInvite />} />
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
						path="/users"
						element={
							<ProtectedRoute>
								<Users />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/roles"
						element={
							<ProtectedRoute>
								<Roles />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/app-services"
						element={
							<ProtectedRoute>
								<AppServices />
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
