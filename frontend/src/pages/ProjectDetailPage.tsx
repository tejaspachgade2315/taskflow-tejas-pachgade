import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ApiError, api } from '../api/client'
import type { ProjectDetail, Task, TaskStatus, User } from '../api/types'
import { TaskModal } from '../components/TaskModal'
import { useAuth } from '../context/useAuth'

type ModalMode = 'create' | 'edit'

function nextStatus(status: TaskStatus): TaskStatus {
    if (status === 'todo') return 'in_progress'
    if (status === 'in_progress') return 'done'
    return 'todo'
}

export function ProjectDetailPage() {
    const { id } = useParams<{ id: string }>()
    const { token } = useAuth()

    const [project, setProject] = useState<ProjectDetail | null>(null)
    const [users, setUsers] = useState<User[]>([])
    const [tasks, setTasks] = useState<Task[]>([])

    const [loadingProject, setLoadingProject] = useState(true)
    const [loadingTasks, setLoadingTasks] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const [statusFilter, setStatusFilter] = useState('')
    const [assigneeFilter, setAssigneeFilter] = useState('')

    const [modalOpen, setModalOpen] = useState(false)
    const [modalMode, setModalMode] = useState<ModalMode>('create')
    const [activeTask, setActiveTask] = useState<Task | undefined>(undefined)

    const projectId = id ?? ''

    useEffect(() => {
        if (!token || !projectId) return

        let cancelled = false
        const load = async () => {
            setLoadingProject(true)
            setError(null)
            try {
                const [projectResponse, usersResponse] = await Promise.all([
                    api.getProject(token, projectId),
                    api.getUsers(token),
                ])

                if (!cancelled) {
                    setProject(projectResponse)
                    setTasks(projectResponse.tasks)
                    setUsers(usersResponse.users)
                }
            } catch (err) {
                if (!cancelled) {
                    const message = err instanceof ApiError ? err.data?.error ?? 'Failed to load project' : 'Failed to load project'
                    setError(message)
                }
            } finally {
                if (!cancelled) {
                    setLoadingProject(false)
                }
            }
        }

        load()
        return () => {
            cancelled = true
        }
    }, [projectId, token])

    const reloadTasks = useCallback(async () => {
        if (!token || !projectId) return

        setLoadingTasks(true)
        try {
            const response = await api.getTasks(token, projectId, statusFilter || undefined, assigneeFilter || undefined)
            setTasks(response.tasks)
        } catch (err) {
            const message = err instanceof ApiError ? err.data?.error ?? 'Failed to load tasks' : 'Failed to load tasks'
            setError(message)
        } finally {
            setLoadingTasks(false)
        }
    }, [assigneeFilter, projectId, statusFilter, token])

    useEffect(() => {
        if (!project) return
        reloadTasks()
    }, [project, statusFilter, assigneeFilter, reloadTasks])

    const saveTask = async (payload: {
        title: string
        description?: string | null
        status?: TaskStatus
        priority?: 'low' | 'medium' | 'high'
        assignee_id?: string | null
        due_date?: string | null
    }) => {
        if (!token || !projectId) return

        if (modalMode === 'create') {
            await api.createTask(token, projectId, payload)
        } else if (activeTask) {
            await api.updateTask(token, activeTask.id, payload)
        }

        await reloadTasks()
    }

    const cycleTaskStatus = async (task: Task) => {
        if (!token) return

        const optimistic = nextStatus(task.status)
        const previousSnapshot = tasks
        setTasks((prev) => prev.map((item) => (item.id === task.id ? { ...item, status: optimistic } : item)))

        try {
            const updated = await api.updateTask(token, task.id, { status: optimistic })
            setTasks((prev) => prev.map((item) => (item.id === task.id ? updated : item)))
        } catch (err) {
            setTasks(previousSnapshot)
            const message = err instanceof ApiError ? err.data?.error ?? 'Status update failed' : 'Status update failed'
            setError(message)
        }
    }

    const deleteTask = async (taskId: string) => {
        if (!token) return
        setError(null)
        try {
            await api.deleteTask(token, taskId)
            setTasks((prev) => prev.filter((task) => task.id !== taskId))
        } catch (err) {
            const message = err instanceof ApiError ? err.data?.error ?? 'Delete failed' : 'Delete failed'
            setError(message)
        }
    }

    const openCreateModal = () => {
        setModalMode('create')
        setActiveTask(undefined)
        setModalOpen(true)
    }

    const openEditModal = (task: Task) => {
        setModalMode('edit')
        setActiveTask(task)
        setModalOpen(true)
    }

    const groupedSummary = useMemo(() => {
        const counts = { todo: 0, in_progress: 0, done: 0 }
        for (const task of tasks) {
            counts[task.status] += 1
        }
        return counts
    }, [tasks])

    if (loadingProject) {
        return <div className="page-state">Loading project...</div>
    }

    if (!project) {
        return <div className="page-state">Project not found.</div>
    }

    return (
        <main className="page-shell">
            <section className="panel">
                <div className="panel-top">
                    <div>
                        <Link className="back-link" to="/projects">
                            Back to Projects
                        </Link>
                        <h1>{project.name}</h1>
                        <p className="muted">{project.description ?? 'No description added yet.'}</p>
                    </div>
                    <button className="button" onClick={openCreateModal}>
                        New Task
                    </button>
                </div>

                <div className="summary-row">
                    <span>Todo: {groupedSummary.todo}</span>
                    <span>In Progress: {groupedSummary.in_progress}</span>
                    <span>Done: {groupedSummary.done}</span>
                </div>

                <div className="filters-row">
                    <label>
                        Status
                        <select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value)}>
                            <option value="">All</option>
                            <option value="todo">Todo</option>
                            <option value="in_progress">In Progress</option>
                            <option value="done">Done</option>
                        </select>
                    </label>

                    <label>
                        Assignee
                        <select value={assigneeFilter} onChange={(event) => setAssigneeFilter(event.target.value)}>
                            <option value="">Everyone</option>
                            {users.map((user) => (
                                <option key={user.id} value={user.id}>
                                    {user.name}
                                </option>
                            ))}
                        </select>
                    </label>
                </div>
            </section>

            {error ? <p className="error-banner">{error}</p> : null}

            {loadingTasks ? <div className="page-state">Refreshing tasks...</div> : null}

            {!loadingTasks && tasks.length === 0 ? (
                <section className="empty-state">
                    <h2>No matching tasks</h2>
                    <p>Create a task or broaden your filters.</p>
                </section>
            ) : null}

            {!loadingTasks && tasks.length > 0 ? (
                <section className="task-list">
                    {tasks.map((task) => (
                        <article key={task.id} className="task-card">
                            <header>
                                <h2>{task.title}</h2>
                                <button className="status-chip" onClick={() => cycleTaskStatus(task)}>
                                    {task.status.replace('_', ' ')}
                                </button>
                            </header>

                            <p>{task.description ?? 'No extra description.'}</p>

                            <div className="task-meta">
                                <span className={`priority ${task.priority}`}>{task.priority}</span>
                                <span>{task.assignee_name ?? 'Unassigned'}</span>
                                <span>{task.due_date ? new Date(task.due_date).toLocaleDateString() : 'No due date'}</span>
                            </div>

                            <div className="task-actions">
                                <button className="button button-ghost" onClick={() => openEditModal(task)}>
                                    Edit
                                </button>
                                <button className="button button-danger" onClick={() => deleteTask(task.id)}>
                                    Delete
                                </button>
                            </div>
                        </article>
                    ))}
                </section>
            ) : null}

            <TaskModal
                open={modalOpen}
                mode={modalMode}
                task={activeTask}
                users={users}
                onClose={() => setModalOpen(false)}
                onSubmit={saveTask}
            />
        </main>
    )
}
