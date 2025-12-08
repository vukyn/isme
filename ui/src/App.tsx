import "./App.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Login } from "./pages/Login";
import { SSOLogin } from "./pages/SSOLogin";
import { Signup } from "./pages/Signup";
import { Welcome } from "./pages/Welcome";
import { NotFound } from "./pages/NotFound";
import { ProtectedRoute } from "./components/ProtectedRoute";

function App() {
	return (
		<BrowserRouter>
			<Routes>
				<Route path="/login" element={<Login />} />
				<Route path="/sso/login" element={<SSOLogin />} />
				<Route path="/signup" element={<Signup />} />
				<Route
					path="/welcome"
					element={
						<ProtectedRoute>
							<Welcome />
						</ProtectedRoute>
					}
				/>
				<Route path="/" element={<Navigate to="/welcome" replace />} />
				<Route path="/404" element={<NotFound />} />
			</Routes>
		</BrowserRouter>
	);
}

export default App;
