import React, { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { useAuthStore } from '@/store/authStore'
import { useViewAsStore } from '@/store/viewAsStore'
import { sharingApi } from '@/api/sharing'
import { SharedAccess } from '@/types/sharing'

interface LayoutProps {
  children: React.ReactNode
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  const location = useLocation()
  const { logout } = useAuth()
  const { user } = useAuthStore()
  const { viewAsUserId, viewAsEmail, isViewingOther, setViewAs, clearViewAs } = useViewAsStore()
  const [owners, setOwners] = useState<SharedAccess[]>([])

  useEffect(() => {
    sharingApi.getConnections()
      .then((res) => setOwners(res.owners || []))
      .catch(() => {})
  }, [])

  const navItems = [
    { path: '/dashboard', label: '儀表板', icon: '📊' },
    { path: '/transactions', label: '交易記錄', icon: '💳' },
    { path: '/upload', label: '上傳帳單', icon: '📄' },
    { path: '/invoices', label: '雲端發票', icon: '🧾' },
    { path: '/budget', label: '預算', icon: '💰' },
    { path: '/charts', label: '圖表', icon: '📈' },
    { path: '/settings', label: '設定', icon: '⚙️' },
  ]

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <div className="flex-shrink-0 flex items-center">
                <h1 className="text-xl font-bold text-primary-600">Billing Note</h1>
              </div>
              <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                {navItems.map((item) => (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={`inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium ${
                      location.pathname === item.path
                        ? 'border-primary-500 text-gray-900'
                        : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
                    }`}
                  >
                    <span className="mr-2">{item.icon}</span>
                    {item.label}
                  </Link>
                ))}
              </div>
            </div>
            <div className="flex items-center gap-3">
              {owners.length > 0 && (
                <select
                  className="text-sm border border-gray-300 rounded-md px-2 py-1"
                  value={viewAsUserId ?? ''}
                  onChange={(e) => {
                    const val = e.target.value
                    if (!val) {
                      clearViewAs()
                    } else {
                      const owner = owners.find(o => o.owner_id === Number(val))
                      setViewAs(Number(val), owner?.owner?.email || null)
                    }
                  }}
                >
                  <option value="">My Data</option>
                  {owners.map((o) => (
                    <option key={o.owner_id} value={o.owner_id}>
                      {o.owner?.email || `User #${o.owner_id}`}
                    </option>
                  ))}
                </select>
              )}
              <span className="text-sm text-gray-700">{user?.email}</span>
              <button
                onClick={logout}
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700"
              >
                登出
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Read-only banner */}
      {isViewingOther && (
        <div className="bg-yellow-50 border-b border-yellow-200 px-4 py-2 text-center">
          <span className="text-sm text-yellow-800">
            You are viewing {viewAsEmail || 'another user'}'s data (read-only).{' '}
            <button
              onClick={clearViewAs}
              className="underline font-medium hover:text-yellow-900"
            >
              Switch back
            </button>
          </span>
        </div>
      )}

      {/* Main Content */}
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">{children}</div>
      </main>
    </div>
  )
}
