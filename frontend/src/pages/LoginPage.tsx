import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { ApiError } from '../api/client'
import { useAuth } from '../context/useAuth'

export function LoginPage() {
    const { login } = useAuth()
    const navigate = useNavigate()
    const location = useLocation()
    const fromPath = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname ?? '/projects'

    const [email, setEmail] = useState('test@example.com')
    const [password, setPassword] = useState('password123')
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const submit = async (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault()
        setError(null)

        if (!email.trim() || !password.trim()) {
            setError('Email and password are required.')
            return
        }

        setBusy(true)
        try {
            await login(email.trim(), password)
            navigate(fromPath, { replace: true })
        } catch (err) {
            if (err instanceof ApiError) {
                setError(err.data?.error ?? 'Login failed')
            } else {
                setError('Unexpected error while logging in')
            }
        } finally {
            setBusy(false)
        }
    }

    return (
        <main className="auth-layout">
            <section className="auth-card">
                <h1>Welcome back</h1>
                <p className="auth-copy">Sign in to continue managing projects and tasks.</p>

                <form className="form-grid" onSubmit={submit}>
                    <label>
                        Email
                        <input
                            type="email"
                            value={email}
                            onChange={(event) => setEmail(event.target.value)}
                            placeholder="you@company.com"
                        />
                    </label>

                    <label>
                        Password
                        <input
                            type="password"
                            value={password}
                            onChange={(event) => setPassword(event.target.value)}
                            placeholder="********"
                        />
                    </label>

                    {error ? <p className="error-banner">{error}</p> : null}

                    <button type="submit" className="button" disabled={busy}>
                        {busy ? 'Signing in...' : 'Login'}
                    </button>
                </form>

                <p className="auth-footnote">
                    New here? <Link to="/register">Create an account</Link>
                </p>
            </section>
        </main>
    )
}
