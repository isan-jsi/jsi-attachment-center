import { useAuth } from "@/context/auth-context";
import LoginPage from "@/pages/login";

export default function App() {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <LoginPage />;
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <p className="text-muted-foreground">Authenticated. Documents page coming soon.</p>
    </div>
  );
}
