import { browser } from '$app/environment'
import { api } from '$lib/api'

export const auth = $state({
    token: browser ? localStorage.getItem('token') : null,
    email: browser ? localStorage.getItem('email') : null,
})

export async function login(email: string, password: string) {
    console.log("🔐 Attempting login for:", email);
    try {
        const res = await api.post('/auth/login', { email, password });
        console.log("✅ Login response:", res.data);

        auth.token = res.data.token;
        auth.email = res.data.email;
        localStorage.setItem('token', res.data.token);
        localStorage.setItem('email', res.data.email);

        console.log("💾 Token stored in localStorage");
        console.log("🔐 Auth state updated:", { token: !!auth.token, email: auth.email });
        return { success: true };
    } catch (error) {
        console.error("❌ Login failed:", error);
        if (error.response) {
            console.error("Response status:", error.response.status);
            console.error("Response data:", error.response.data);
        }
        throw error;
    }
}

export async function register(email: string, password: string) {
    console.log("📝 Attempting registration for:", email);
    try {
        const res = await api.post('/auth/register', { email, password });
        console.log("✅ Registration response:", res.data);

        auth.token = res.data.token;
        auth.email = res.data.email;
        localStorage.setItem('token', res.data.token);
        localStorage.setItem('email', res.data.email);

        console.log("💾 Token stored in localStorage");
        return { success: true };
    } catch (error) {
        console.error("❌ Registration failed:", error);
        throw error;
    }
}

export function logout() {
    console.log("🚪 Logging out");
    auth.token = null;
    auth.email = null;
    localStorage.removeItem('token');
    localStorage.removeItem('email');
    console.log("✅ Cleared auth data");
}