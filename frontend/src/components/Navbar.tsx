import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/useAuth'

export function Navbar() {
    const { user, logout } = useAuth()
    const navigate = useNavigate()

    return (
        <header className="navbar">
            <div className="navbar-left">
                <Link className="brand" to="/projects">
                    TaskFlow
                </Link>
                <span className="badge">Full Stack Demo</span>
            </div>

            <div className="navbar-right">
                <span className="user-name">{user?.name}</span>
                <button
                    className="button button-ghost"
                    onClick={() => {
                        logout()
                        navigate('/login')
                    }}
                >
                    Logout
                </button>
            </div>
        </header>
    )
}
