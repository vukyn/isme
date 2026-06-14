import "./App.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Login } from "./pages/Login";
import { SSOLogin } from "./pages/SSOLogin";
import { AcceptInvite } from "./pages/AcceptInvite";
import { Welcome } from "./pages/Welcome";
import { Users } from "./pages/Users";
import { InviteUser } from "./pages/InviteUser";
import { Roles } from "./pages/Roles";
import { AppServices } from "./pages/AppServices";
import { EditAppService } from "./pages/EditAppService";
import { Settings } from "./pages/Settings";
import { Profile } from "./pages/Profile";
import { Sessions } from "./pages/Sessions";
import { Activity } from "./pages/Activity";
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
						path="/users"
						element={
							<ProtectedRoute requiredPermission="user:read">
								<Users />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/users/invite"
						element={
							<ProtectedRoute requiredPermission="user:read">
								<InviteUser />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/roles"
						element={
							<ProtectedRoute requiredPermission="role:read">
								<Roles />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/app-services"
						element={
							<ProtectedRoute requiredPermission="app_service:read">
								<AppServices />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/app-services/:id/edit"
						element={
							<ProtectedRoute requiredPermission="app_service:update">
								<EditAppService />
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
					<Route
						path="/profile"
						element={
							<ProtectedRoute>
								<Profile />
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
						path="/activity"
						element={
							<ProtectedRoute>
								<Activity />
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
