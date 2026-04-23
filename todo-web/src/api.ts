import { getApiUrl } from './env';
import { clearAuth, getToken } from './auth';

export interface AuthResponse {
  token: string;
  userId: string;
}

export interface Category {
  id: string;
  name: string;
}

export interface Task {
  id: string;
  title: string;
  dueDate: string | null;
  categoryId: string | null;
  completed: boolean;
  createdAt: string;
}

export interface TaskInput {
  title: string;
  dueDate?: string | null;
  categoryId?: string | null;
  completed?: boolean;
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(
  path: string,
  options: RequestInit & { auth?: boolean } = {}
): Promise<T> {
  const { auth = true, headers, ...rest } = options;
  const base = getApiUrl();
  const url = `${base}${path}`;

  const finalHeaders: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(headers as Record<string, string> | undefined),
  };
  if (auth) {
    const token = getToken();
    if (token) finalHeaders['Authorization'] = `Bearer ${token}`;
  }

  let res: Response;
  try {
    res = await fetch(url, { ...rest, headers: finalHeaders });
  } catch (err) {
    throw new ApiError(0, `Network error: ${(err as Error).message}`);
  }

  if (res.status === 401 && auth) {
    clearAuth();
    // Let the router-level guard redirect; also hard-redirect for safety.
    if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
      window.location.href = '/login';
    }
    throw new ApiError(401, 'Unauthorized');
  }

  if (res.status === 204) {
    return undefined as unknown as T;
  }

  const text = await res.text();
  const data = text ? safeParseJson(text) : null;

  if (!res.ok) {
    const msg =
      (data && typeof data === 'object' && 'error' in (data as object)
        ? String((data as Record<string, unknown>).error)
        : text) || res.statusText;
    throw new ApiError(res.status, msg || `HTTP ${res.status}`);
  }

  return data as T;
}

function safeParseJson(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

// --- Auth ---
export function signup(email: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>('/auth/signup', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
    auth: false,
  });
}

export function login(email: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
    auth: false,
  });
}

// --- Categories ---
export function listCategories(): Promise<Category[]> {
  return request<Category[]>('/categories', { method: 'GET' }).then(
    (d) => d ?? []
  );
}

export function createCategory(name: string): Promise<Category> {
  return request<Category>('/categories', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export function deleteCategory(id: string): Promise<void> {
  return request<void>(`/categories/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  });
}

// --- Tasks ---
export function listTasks(categoryId?: string): Promise<Task[]> {
  const q = categoryId ? `?categoryId=${encodeURIComponent(categoryId)}` : '';
  return request<Task[]>(`/tasks${q}`, { method: 'GET' }).then((d) => d ?? []);
}

export function createTask(input: TaskInput): Promise<Task> {
  return request<Task>('/tasks', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export function updateTask(id: string, input: TaskInput): Promise<Task> {
  return request<Task>(`/tasks/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}

export function deleteTask(id: string): Promise<void> {
  return request<void>(`/tasks/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  });
}

export function toggleTaskComplete(id: string, completed: boolean): Promise<Task> {
  return request<Task>(
    `/tasks/${encodeURIComponent(id)}/complete?completed=${completed}`,
    { method: 'POST' }
  );
}
