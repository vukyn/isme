import "./App.css";
import { lazy, Suspense } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Flex, Spinner } from "@chakra-ui/react";

import { ProtectedRoute } from "./components/ProtectedRoute";
import { AuroraBackground } from "./components/ui/aurora-background";

// Route-level code splitting: every page is its own lazily-loaded chunk so the
// initial bundle doesn't ship the whole app (each page loads only when
// navigated to). The pages are named exports, so the dynamic import is mapped to
// a default for React.lazy. Structural wrappers (ProtectedRoute, AuroraBackground)
// stay eager.
const Login = lazy(() => import("./pages/Login").then((m) => ({ default: m.Login })));
const SSOLogin = lazy(() => import("./pages/SSOLogin").then((m) => ({ default: m.SSOLogin })));
const AcceptInvite = lazy(() => import("./pages/AcceptInvite").then((m) => ({ default: m.AcceptInvite })));
const Welcome = lazy(() => import("./pages/Welcome").then((m) => ({ default: m.Welcome })));
const Users = lazy(() => import("./pages/Users").then((m) => ({ default: m.Users })));
const InviteUser = lazy(() => import("./pages/InviteUser").then((m) => ({ default: m.InviteUser })));
const Roles = lazy(() => import("./pages/Roles").then((m) => ({ default: m.Roles })));
const AppServices = lazy(() => import("./pages/AppServices").then((m) => ({ default: m.AppServices })));
const EditAppService = lazy(() => import("./pages/EditAppService").then((m) => ({ default: m.EditAppService })));
const Settings = lazy(() => import("./pages/Settings").then((m) => ({ default: m.Settings })));
const Profile = lazy(() => import("./pages/Profile").then((m) => ({ default: m.Profile })));
const Sessions = lazy(() => import("./pages/Sessions").then((m) => ({ default: m.Sessions })));
const Activity = lazy(() => import("./pages/Activity").then((m) => ({ default: m.Activity })));
const NotFound = lazy(() => import("./pages/NotFound").then((m) => ({ default: m.NotFound })));

/** Full-screen fallback shown while a route chunk loads. */
const RouteFallback = () => (
	<Flex minH="100dvh" align="center" justify="center" bg="bg">
		<Spinner size="lg" color="accent" />
	</Flex>
);

function App() {
	return (
		<>
			<AuroraBackground />
			<BrowserRouter>
				<Suspense fallback={<RouteFallback />}>
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
				</Suspense>
			</BrowserRouter>
		</>
	);
}

export default App;
