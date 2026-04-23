const TOKEN_KEY = 'todo.jwt';
const USER_ID_KEY = 'todo.userId';

export function getToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY);
  } catch {
    return null;
  }
}

export function setAuth(token: string, userId: string): void {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_ID_KEY, userId);
}

export function clearAuth(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_ID_KEY);
}

export function isAuthenticated(): boolean {
  const token = getToken();
  return !!token && token.length > 0;
}
