import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ApiError } from '../api/client'
import { useAuth } from '../context/useAuth'

export function RegisterPage() {
    const { register } = useAuth()
    const navigate = useNavigate()

    const [name, setName] = useState('')
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const submit = async (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault()
        setError(null)

        if (!name.trim()) {
            setError('Name is required.')
            return
        }
        if (!email.trim() || !email.includes('@')) {
            setError('A valid email is required.')
            return
        }
        if (password.length < 8) {
            setError('Password must be at least 8 characters.')
            return
        }

        setBusy(true)
        try {
            await register(name.trim(), email.trim(), password)
            navigate('/projects', { replace: true })
        } catch (err) {
            if (err instanceof ApiError) {
                const firstFieldError = err.data?.fields ? Object.values(err.data.fields)[0] : null
                setError(firstFieldError ?? err.data?.error ?? 'Registration failed')
            } else {
                setError('Unexpected error while registering')
            }
        } finally {
            setBusy(false)
        }
    }

    return (
        <main className="auth-layout">
            <section className="auth-card">
                <h1>Create account</h1>
                <p className="auth-copy">Get your workspace ready in less than a minute.</p>

                <form className="form-grid" onSubmit={submit}>
                    <label>
                        Name
                        <input
                            value={name}
                            onChange={(event) => setName(event.target.value)}
                            placeholder="Jane Doe"
                        />
                    </label>

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
                            placeholder="At least 8 characters"
                        />
                    </label>

                    {error ? <p className="error-banner">{error}</p> : null}

                    <button type="submit" className="button" disabled={busy}>
                        {busy ? 'Creating account...' : 'Register'}
                    </button>
                </form>

                <p className="auth-footnote">
                    Already registered? <Link to="/login">Login</Link>
                </p>
            </section>
        </main>
    )
}
