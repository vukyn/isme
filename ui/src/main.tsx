import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "@/index.css";
import App from "@/App.tsx";
import { Provider } from "@/components/ui/provider";
import { Toaster } from "@/components/ui/toaster";
import { UserProvider } from "@/contexts/UserProvider";

createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<Provider>
			<UserProvider>
				<App />
				<Toaster />
			</UserProvider>
		</Provider>
	</StrictMode>
);
