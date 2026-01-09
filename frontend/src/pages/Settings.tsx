import { useState, useEffect } from 'react'
import { useAuthStore } from '@/store/authStore'
import { Button } from '@/components/common/Button'
import { useNavigate } from 'react-router-dom'
import {
  getPDFPasswords,
  setMultiplePDFPasswords,
  PDFPassword,
  PDFPasswordInput,
} from '@/api/upload'

export const Settings = () => {
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()
  const [passwords, setPasswords] = useState<PDFPassword[]>([])
  const [passwordInputs, setPasswordInputs] = useState<string[]>(['', '', '', ''])
  const [passwordLabels, setPasswordLabels] = useState<string[]>(['', '', '', ''])
  const [saving, setSaving] = useState(false)
  const [saveMessage, setSaveMessage] = useState<string | null>(null)

  useEffect(() => {
    loadPasswords()
  }, [])

  const loadPasswords = async () => {
    try {
      const response = await getPDFPasswords()
      setPasswords(response.passwords)
      // Initialize labels from existing passwords
      const labels = ['', '', '', '']
      response.passwords.forEach(p => {
        if (p.priority >= 1 && p.priority <= 4) {
          labels[p.priority - 1] = p.label || ''
        }
      })
      setPasswordLabels(labels)
    } catch (error) {
      console.error('Failed to load passwords:', error)
    }
  }

  const handleSavePasswords = async () => {
    setSaving(true)
    setSaveMessage(null)
    try {
      const inputs: PDFPasswordInput[] = passwordInputs
        .map((password, index) => ({
          password,
          priority: index + 1,
          label: passwordLabels[index],
        }))
        .filter(input => input.password !== '')

      await setMultiplePDFPasswords(inputs)
      setSaveMessage('Passwords saved successfully')
      setPasswordInputs(['', '', '', ''])
      loadPasswords()
    } catch (error) {
      setSaveMessage('Failed to save passwords')
    } finally {
      setSaving(false)
    }
  }

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Settings</h1>

      <div className="max-w-2xl">
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Profile Information</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Email
              </label>
              <div className="text-gray-900">{user?.email || 'N/A'}</div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Name
              </label>
              <div className="text-gray-900">{user?.name || 'N/A'}</div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                User ID
              </label>
              <div className="text-gray-900">{user?.id || 'N/A'}</div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Member Since
              </label>
              <div className="text-gray-900">
                {user?.created_at
                  ? new Date(user.created_at).toLocaleDateString()
                  : 'N/A'}
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Application Settings</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Currency
              </label>
              <select
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                defaultValue="USD"
              >
                <option value="USD">USD ($)</option>
                <option value="EUR">EUR (€)</option>
                <option value="GBP">GBP (£)</option>
                <option value="TWD">TWD (NT$)</option>
              </select>
              <p className="text-sm text-gray-500 mt-1">
                This feature is coming soon
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Date Format
              </label>
              <select
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                defaultValue="MM/DD/YYYY"
              >
                <option value="MM/DD/YYYY">MM/DD/YYYY</option>
                <option value="DD/MM/YYYY">DD/MM/YYYY</option>
                <option value="YYYY-MM-DD">YYYY-MM-DD</option>
              </select>
              <p className="text-sm text-gray-500 mt-1">
                This feature is coming soon
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Theme
              </label>
              <select
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                defaultValue="light"
              >
                <option value="light">Light</option>
                <option value="dark">Dark</option>
                <option value="auto">Auto</option>
              </select>
              <p className="text-sm text-gray-500 mt-1">
                This feature is coming soon
              </p>
            </div>
          </div>
        </div>

        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">PDF Password Management</h2>
          <p className="text-sm text-gray-600 mb-4">
            Set up to 4 passwords for decrypting credit card statement PDFs.
            Passwords are tried in order (1 → 2 → 3 → 4).
          </p>

          {saveMessage && (
            <div
              className={`mb-4 p-3 rounded ${
                saveMessage.includes('success')
                  ? 'bg-green-100 text-green-700'
                  : 'bg-red-100 text-red-700'
              }`}
            >
              {saveMessage}
            </div>
          )}

          <div className="space-y-4">
            {[0, 1, 2, 3].map(index => {
              const existingPassword = passwords.find(
                p => p.priority === index + 1
              )
              return (
                <div key={index} className="flex gap-4 items-center">
                  <span className="text-gray-500 w-8">#{index + 1}</span>
                  <input
                    type="password"
                    placeholder={
                      existingPassword?.has_value
                        ? '••••••••'
                        : 'Enter password'
                    }
                    value={passwordInputs[index]}
                    onChange={e => {
                      const newInputs = [...passwordInputs]
                      newInputs[index] = e.target.value
                      setPasswordInputs(newInputs)
                    }}
                    className="flex-1 border border-gray-300 rounded-md px-3 py-2"
                  />
                  <input
                    type="text"
                    placeholder="Label (optional)"
                    value={passwordLabels[index]}
                    onChange={e => {
                      const newLabels = [...passwordLabels]
                      newLabels[index] = e.target.value
                      setPasswordLabels(newLabels)
                    }}
                    className="w-40 border border-gray-300 rounded-md px-3 py-2"
                  />
                  {existingPassword?.has_value && (
                    <span className="text-green-500 text-sm">Set</span>
                  )}
                </div>
              )
            })}
          </div>

          <Button
            onClick={handleSavePasswords}
            disabled={saving}
            className="mt-4"
          >
            {saving ? 'Saving...' : 'Save Passwords'}
          </Button>
        </div>

        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">About</h2>
          <div className="space-y-2">
            <div className="text-gray-700">
              <span className="font-medium">Version:</span> 1.0.0
            </div>
            <div className="text-gray-700">
              <span className="font-medium">Status:</span> Phase 1 - MVP
            </div>
            <div className="text-gray-700">
              <span className="font-medium">Description:</span> A simple billing
              and expense tracking application
            </div>
          </div>
        </div>

        <div className="bg-red-50 shadow rounded-lg p-6">
          <h2 className="text-xl font-semibold mb-4 text-red-900">
            Danger Zone
          </h2>
          <div className="space-y-4">
            <div>
              <p className="text-sm text-gray-700 mb-4">
                Once you logout, you will need to login again to access your
                account.
              </p>
              <Button variant="secondary" onClick={handleLogout}>
                Logout
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
