import "./App.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Login } from "./pages/Login";
import { Welcome } from "./pages/Welcome";
import { ProtectedRoute } from "./components/ProtectedRoute";

function App() {
	return (
		<BrowserRouter>
			<Routes>
				<Route path="/login" element={<Login />} />
				<Route
					path="/welcome"
					element={
						<ProtectedRoute>
							<Welcome />
						</ProtectedRoute>
					}
				/>
				<Route path="/" element={<Navigate to="/welcome" replace />} />
			</Routes>
		</BrowserRouter>
	);
}

export default App;
