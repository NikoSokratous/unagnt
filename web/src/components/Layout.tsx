import React from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Activity, Settings, GitBranch, Store, BarChart3, Shield, CheckSquare } from 'lucide-react'
import './Layout.css'

interface LayoutProps {
  children: React.ReactNode
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const location = useLocation()

  const navItems = [
    { to: '/', icon: Activity, label: 'Dashboard' },
    { to: '/workflows/designer', icon: GitBranch, label: 'Workflow Designer' },
    { to: '/workflows/marketplace', icon: Store, label: 'Marketplace' },
    { to: '/analytics', icon: BarChart3, label: 'Analytics' },
    { to: '/policy-playground', icon: Shield, label: 'Policy Playground' },
    { to: '/approvals', icon: CheckSquare, label: 'Approvals' },
  ]

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="logo">
            <Activity size={24} className="logo-icon" />
            <span className="logo-text">AgentRuntime</span>
          </div>
        </div>

        <nav className="nav">
          {navItems.map(({ to, icon: Icon, label }) => (
            <Link
              key={to}
              to={to}
              className={`nav-link ${location.pathname === to || (to !== '/' && location.pathname.startsWith(to)) ? 'active' : ''}`}
            >
              <Icon size={20} />
              <span>{label}</span>
            </Link>
          ))}
        </nav>

        <div className="sidebar-footer">
          <Link to="/settings" className="nav-link">
            <Settings size={20} />
            <span>Settings</span>
          </Link>
          <div className="version">v2.0.0</div>
        </div>
      </aside>

      <main className="main-content">
        {children}
      </main>
    </div>
  )
}

export default Layout
