export type TaskStatus = 'todo' | 'in_progress' | 'done'
export type TaskPriority = 'low' | 'medium' | 'high'

export interface User {
    id: string
    name: string
    email: string
    created_at: string
}

export interface Project {
    id: string
    name: string
    description?: string
    owner_id: string
    created_at: string
}

export interface Task {
    id: string
    title: string
    description?: string
    status: TaskStatus
    priority: TaskPriority
    project_id: string
    assignee_id?: string
    assignee_name?: string
    creator_id: string
    due_date?: string
    created_at: string
    updated_at: string
}

export interface ProjectDetail {
    id: string
    name: string
    description?: string
    owner_id: string
    created_at: string
    tasks: Task[]
}

export interface ValidationErrorPayload {
    error: string
    fields?: Record<string, string>
}
