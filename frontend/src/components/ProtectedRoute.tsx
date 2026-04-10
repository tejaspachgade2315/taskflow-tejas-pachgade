import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../context/useAuth'

export function ProtectedRoute() {
    const { token, ready } = useAuth()
    const location = useLocation()

    if (!ready) {
        return <div className="page-state">Loading session...</div>
    }

    if (!token) {
        return <Navigate to="/login" state={{ from: location }} replace />
    }

    return <Outlet />
}
