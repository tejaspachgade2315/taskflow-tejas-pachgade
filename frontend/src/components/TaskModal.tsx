import { useEffect, useMemo, useState } from 'react'
import type { Task, TaskPriority, TaskStatus, User } from '../api/types'

interface TaskModalPayload {
    title: string
    description?: string | null
    status?: TaskStatus
    priority?: TaskPriority
    assignee_id?: string | null
    due_date?: string | null
}

interface TaskModalProps {
    open: boolean
    mode: 'create' | 'edit'
    task?: Task
    users: User[]
    onClose: () => void
    onSubmit: (payload: TaskModalPayload) => Promise<void>
}

interface FormState {
    title: string
    description: string
    status: TaskStatus
    priority: TaskPriority
    assignee_id: string
    due_date: string
}

const defaultState: FormState = {
    title: '',
    description: '',
    status: 'todo',
    priority: 'medium',
    assignee_id: '',
    due_date: '',
}

function taskToForm(task?: Task): FormState {
    if (!task) return defaultState
    return {
        title: task.title,
        description: task.description ?? '',
        status: task.status,
        priority: task.priority,
        assignee_id: task.assignee_id ?? '',
        due_date: task.due_date ? task.due_date.slice(0, 10) : '',
    }
}

export function TaskModal({ open, mode, task, users, onClose, onSubmit }: TaskModalProps) {
    const [form, setForm] = useState<FormState>(defaultState)
    const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
    const [busy, setBusy] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)

    useEffect(() => {
        if (!open) return
        setForm(taskToForm(task))
        setFieldErrors({})
        setSubmitError(null)
    }, [open, task])

    const title = useMemo(() => (mode === 'create' ? 'Create Task' : 'Edit Task'), [mode])

    if (!open) return null

    const validate = () => {
        const errors: Record<string, string> = {}
        if (!form.title.trim()) {
            errors.title = 'Title is required'
        }
        setFieldErrors(errors)
        return Object.keys(errors).length === 0
    }

    const submit = async (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault()
        if (!validate()) return

        setBusy(true)
        setSubmitError(null)

        try {
            await onSubmit({
                title: form.title.trim(),
                description: form.description.trim() === '' ? null : form.description.trim(),
                status: form.status,
                priority: form.priority,
                assignee_id: form.assignee_id || null,
                due_date: form.due_date || null,
            })
            onClose()
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Failed to save task'
            setSubmitError(message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <div className="modal-backdrop" role="presentation" onClick={onClose}>
            <section className="modal" role="dialog" aria-modal="true" onClick={(event) => event.stopPropagation()}>
                <h2>{title}</h2>

                <form className="form-grid" onSubmit={submit}>
                    <label>
                        Title
                        <input
                            value={form.title}
                            onChange={(event) => setForm((prev) => ({ ...prev, title: event.target.value }))}
                            placeholder="Write API docs"
                        />
                        {fieldErrors.title ? <small className="error-text">{fieldErrors.title}</small> : null}
                    </label>

                    <label>
                        Description
                        <textarea
                            rows={3}
                            value={form.description}
                            onChange={(event) => setForm((prev) => ({ ...prev, description: event.target.value }))}
                            placeholder="Optional details"
                        />
                    </label>

                    <div className="form-row">
                        <label>
                            Status
                            <select
                                value={form.status}
                                onChange={(event) => setForm((prev) => ({ ...prev, status: event.target.value as TaskStatus }))}
                            >
                                <option value="todo">Todo</option>
                                <option value="in_progress">In Progress</option>
                                <option value="done">Done</option>
                            </select>
                        </label>

                        <label>
                            Priority
                            <select
                                value={form.priority}
                                onChange={(event) => setForm((prev) => ({ ...prev, priority: event.target.value as TaskPriority }))}
                            >
                                <option value="low">Low</option>
                                <option value="medium">Medium</option>
                                <option value="high">High</option>
                            </select>
                        </label>
                    </div>

                    <div className="form-row">
                        <label>
                            Assignee
                            <select
                                value={form.assignee_id}
                                onChange={(event) => setForm((prev) => ({ ...prev, assignee_id: event.target.value }))}
                            >
                                <option value="">Unassigned</option>
                                {users.map((user) => (
                                    <option key={user.id} value={user.id}>
                                        {user.name}
                                    </option>
                                ))}
                            </select>
                        </label>

                        <label>
                            Due Date
                            <input
                                type="date"
                                value={form.due_date}
                                onChange={(event) => setForm((prev) => ({ ...prev, due_date: event.target.value }))}
                            />
                        </label>
                    </div>

                    {submitError ? <p className="error-banner">{submitError}</p> : null}

                    <div className="modal-actions">
                        <button type="button" className="button button-ghost" onClick={onClose} disabled={busy}>
                            Cancel
                        </button>
                        <button type="submit" className="button" disabled={busy}>
                            {busy ? 'Saving...' : mode === 'create' ? 'Create' : 'Save'}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    )
}
