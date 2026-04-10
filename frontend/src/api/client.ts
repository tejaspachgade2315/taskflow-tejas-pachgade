import type { Project, ProjectDetail, Task, TaskPriority, TaskStatus, User, ValidationErrorPayload } from './types'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export class ApiError extends Error {
    status: number
    data?: ValidationErrorPayload

    constructor(status: number, message: string, data?: ValidationErrorPayload) {
        super(message)
        this.status = status
        this.data = data
    }
}

interface RequestOptions {
    method?: 'GET' | 'POST' | 'PATCH' | 'DELETE'
    token?: string | null
    body?: unknown
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const headers: Record<string, string> = {
        Accept: 'application/json',
    }

    if (options.body !== undefined) {
        headers['Content-Type'] = 'application/json'
    }

    if (options.token) {
        headers.Authorization = `Bearer ${options.token}`
    }

    const response = await fetch(`${API_URL}${path}`, {
        method: options.method ?? 'GET',
        headers,
        body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
    })

    if (response.status === 204) {
        return undefined as T
    }

    const data = (await response.json()) as T | ValidationErrorPayload
    if (!response.ok) {
        const payload = data as ValidationErrorPayload
        throw new ApiError(response.status, payload.error ?? 'Request failed', payload)
    }

    return data as T
}

interface AuthPayload {
    token: string
    user: User
}

interface AuthRequest {
    name?: string
    email: string
    password: string
}

interface CreateProjectRequest {
    name: string
    description?: string
}

interface CreateTaskRequest {
    title: string
    description?: string | null
    status?: TaskStatus
    priority?: TaskPriority
    assignee_id?: string | null
    due_date?: string | null
}

interface UpdateTaskRequest {
    title?: string
    description?: string | null
    status?: TaskStatus
    priority?: TaskPriority
    assignee_id?: string | null
    due_date?: string | null
}

export const api = {
    register(payload: AuthRequest) {
        return request<AuthPayload>('/auth/register', { method: 'POST', body: payload })
    },

    login(payload: AuthRequest) {
        return request<AuthPayload>('/auth/login', { method: 'POST', body: payload })
    },

    getUsers(token: string) {
        return request<{ users: User[] }>('/users', { token })
    },

    getProjects(token: string, page = 1, limit = 20) {
        return request<{ projects: Project[] }>(`/projects?page=${page}&limit=${limit}`, { token })
    },

    createProject(token: string, payload: CreateProjectRequest) {
        return request<Project>('/projects', { method: 'POST', token, body: payload })
    },

    getProject(token: string, projectId: string) {
        return request<ProjectDetail>(`/projects/${projectId}`, { token })
    },

    getTasks(token: string, projectId: string, status?: string, assignee?: string) {
        const query = new URLSearchParams()
        if (status) query.set('status', status)
        if (assignee) query.set('assignee', assignee)
        const suffix = query.toString() ? `?${query.toString()}` : ''
        return request<{ tasks: Task[] }>(`/projects/${projectId}/tasks${suffix}`, { token })
    },

    createTask(token: string, projectId: string, payload: CreateTaskRequest) {
        return request<Task>(`/projects/${projectId}/tasks`, {
            method: 'POST',
            token,
            body: payload,
        })
    },

    updateTask(token: string, taskId: string, payload: UpdateTaskRequest) {
        return request<Task>(`/tasks/${taskId}`, {
            method: 'PATCH',
            token,
            body: payload,
        })
    },

    deleteTask(token: string, taskId: string) {
        return request<void>(`/tasks/${taskId}`, {
            method: 'DELETE',
            token,
        })
    },
}
