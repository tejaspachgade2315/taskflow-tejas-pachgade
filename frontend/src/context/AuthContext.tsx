import { createContext, useEffect, useMemo, useState } from 'react'
import { api } from '../api/client'
import type { User } from '../api/types'

interface AuthContextValue {
    token: string | null
    user: User | null
    ready: boolean
    login: (email: string, password: string) => Promise<void>
    register: (name: string, email: string, password: string) => Promise<void>
    logout: () => void
}

const STORAGE_KEY = 'taskflow.auth'

// eslint-disable-next-line react-refresh/only-export-components
export const AuthContext = createContext<AuthContextValue | undefined>(undefined)

function saveAuth(token: string, user: User) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ token, user }))
}

function clearAuth() {
    localStorage.removeItem(STORAGE_KEY)
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [token, setToken] = useState<string | null>(null)
    const [user, setUser] = useState<User | null>(null)
    const [ready, setReady] = useState(false)

    useEffect(() => {
        const raw = localStorage.getItem(STORAGE_KEY)
        if (!raw) {
            setReady(true)
            return
        }

        try {
            const parsed = JSON.parse(raw) as { token: string; user: User }
            setToken(parsed.token)
            setUser(parsed.user)
        } catch {
            clearAuth()
        } finally {
            setReady(true)
        }
    }, [])

    const value = useMemo<AuthContextValue>(
        () => ({
            token,
            user,
            ready,
            async login(email: string, password: string) {
                const response = await api.login({ email, password })
                setToken(response.token)
                setUser(response.user)
                saveAuth(response.token, response.user)
            },
            async register(name: string, email: string, password: string) {
                const response = await api.register({ name, email, password })
                setToken(response.token)
                setUser(response.user)
                saveAuth(response.token, response.user)
            },
            logout() {
                setToken(null)
                setUser(null)
                clearAuth()
            },
        }),
        [ready, token, user],
    )

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
