import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ApiError, api } from '../api/client'
import type { Project } from '../api/types'
import { useAuth } from '../context/useAuth'

export function ProjectsPage() {
    const { token } = useAuth()
    const [projects, setProjects] = useState<Project[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    const [name, setName] = useState('')
    const [description, setDescription] = useState('')
    const [creating, setCreating] = useState(false)

    useEffect(() => {
        if (!token) return

        let cancelled = false
        const load = async () => {
            setLoading(true)
            setError(null)
            try {
                const response = await api.getProjects(token)
                if (!cancelled) {
                    setProjects(response.projects)
                }
            } catch (err) {
                if (!cancelled) {
                    const message = err instanceof ApiError ? err.data?.error ?? 'Failed to load projects' : 'Failed to load projects'
                    setError(message)
                }
            } finally {
                if (!cancelled) {
                    setLoading(false)
                }
            }
        }

        load()
        return () => {
            cancelled = true
        }
    }, [token])

    const createProject = async (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault()
        if (!token) return

        if (!name.trim()) {
            setError('Project name is required.')
            return
        }

        setCreating(true)
        setError(null)
        try {
            const created = await api.createProject(token, {
                name: name.trim(),
                description: description.trim() || undefined,
            })
            setProjects((prev) => [created, ...prev])
            setName('')
            setDescription('')
        } catch (err) {
            const message = err instanceof ApiError ? err.data?.error ?? 'Failed to create project' : 'Failed to create project'
            setError(message)
        } finally {
            setCreating(false)
        }
    }

    return (
        <main className="page-shell">
            <section className="panel">
                <h1>Projects</h1>
                <p className="muted">Everything you own or collaborate on shows up here.</p>

                <form className="form-grid" onSubmit={createProject}>
                    <label>
                        Project name
                        <input
                            value={name}
                            onChange={(event) => setName(event.target.value)}
                            placeholder="Q2 Product Launch"
                        />
                    </label>
                    <label>
                        Description
                        <input
                            value={description}
                            onChange={(event) => setDescription(event.target.value)}
                            placeholder="Optional"
                        />
                    </label>
                    <button type="submit" className="button" disabled={creating}>
                        {creating ? 'Creating...' : 'Create Project'}
                    </button>
                </form>
            </section>

            {error ? <p className="error-banner">{error}</p> : null}

            {loading ? <div className="page-state">Loading projects...</div> : null}

            {!loading && projects.length === 0 ? (
                <section className="empty-state">
                    <h2>No projects yet</h2>
                    <p>Create your first project to start tracking tasks.</p>
                </section>
            ) : null}

            {!loading && projects.length > 0 ? (
                <section className="card-grid">
                    {projects.map((project) => (
                        <article key={project.id} className="project-card">
                            <h2>{project.name}</h2>
                            <p>{project.description ?? 'No description yet.'}</p>
                            <div className="card-footer">
                                <small>{new Date(project.created_at).toLocaleDateString()}</small>
                                <Link className="button button-ghost" to={`/projects/${project.id}`}>
                                    Open
                                </Link>
                            </div>
                        </article>
                    ))}
                </section>
            ) : null}
        </main>
    )
}
