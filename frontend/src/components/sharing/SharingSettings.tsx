import { useState, useEffect } from 'react'
import { Button } from '@/components/common/Button'
import { sharingApi } from '@/api/sharing'
import { SharedAccess } from '@/types/sharing'

export const SharingSettings = () => {
  const [pairingCode, setPairingCode] = useState('')
  const [pairInput, setPairInput] = useState('')
  const [viewers, setViewers] = useState<SharedAccess[]>([])
  const [owners, setOwners] = useState<SharedAccess[]>([])
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState<string | null>(null)
  const [messageType, setMessageType] = useState<'success' | 'error'>('success')

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [codeRes, connRes] = await Promise.all([
        sharingApi.getMyCode(),
        sharingApi.getConnections(),
      ])
      setPairingCode(codeRes.code)
      setViewers(connRes.viewers || [])
      setOwners(connRes.owners || [])
    } catch {
      showMessage('Failed to load sharing data', 'error')
    } finally {
      setLoading(false)
    }
  }

  const showMessage = (msg: string, type: 'success' | 'error') => {
    setMessage(msg)
    setMessageType(type)
    setTimeout(() => setMessage(null), 3000)
  }

  const handleRegenerate = async () => {
    try {
      const res = await sharingApi.regenerateCode()
      setPairingCode(res.code)
      showMessage('Pairing code regenerated', 'success')
    } catch {
      showMessage('Failed to regenerate code', 'error')
    }
  }

  const handlePair = async () => {
    if (!pairInput.trim()) return
    try {
      await sharingApi.pair(pairInput.trim())
      setPairInput('')
      showMessage('Paired successfully!', 'success')
      loadData()
    } catch (err: any) {
      const msg = err?.response?.data?.error || 'Failed to pair'
      showMessage(msg, 'error')
    }
  }

  const handleRevoke = async (uid: number) => {
    try {
      await sharingApi.revokeAccess(uid)
      showMessage('Access revoked', 'success')
      loadData()
    } catch {
      showMessage('Failed to revoke access', 'error')
    }
  }

  if (loading) {
    return (
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Sharing</h2>
        <p className="text-gray-500">Loading...</p>
      </div>
    )
  }

  return (
    <div className="bg-white shadow rounded-lg p-6 mb-6">
      <h2 className="text-xl font-semibold mb-4">Sharing</h2>

      {message && (
        <div
          className={`mb-4 p-3 rounded ${
            messageType === 'success'
              ? 'bg-green-100 text-green-700'
              : 'bg-red-100 text-red-700'
          }`}
        >
          {message}
        </div>
      )}

      {/* My Pairing Code */}
      <div className="mb-6">
        <h3 className="text-sm font-medium text-gray-700 mb-2">My Pairing Code</h3>
        <p className="text-xs text-gray-500 mb-2">
          Share this code with family members so they can view your financial data.
        </p>
        <div className="flex items-center gap-3">
          <code className="text-2xl font-mono tracking-widest bg-gray-100 px-4 py-2 rounded">
            {pairingCode}
          </code>
          <Button variant="secondary" size="sm" onClick={handleRegenerate}>
            Regenerate
          </Button>
        </div>
      </div>

      {/* Pair with someone */}
      <div className="mb-6">
        <h3 className="text-sm font-medium text-gray-700 mb-2">Pair with Someone</h3>
        <div className="flex gap-3">
          <input
            type="text"
            value={pairInput}
            onChange={(e) => setPairInput(e.target.value.toUpperCase())}
            placeholder="AB12-CD34"
            maxLength={9}
            className="border border-gray-300 rounded-md px-3 py-2 font-mono uppercase w-40"
          />
          <Button size="sm" onClick={handlePair} disabled={!pairInput.trim()}>
            Pair
          </Button>
        </div>
      </div>

      {/* Viewers (who can see my data) */}
      {viewers.length > 0 && (
        <div className="mb-6">
          <h3 className="text-sm font-medium text-gray-700 mb-2">
            Who can see my data
          </h3>
          <ul className="divide-y divide-gray-200">
            {viewers.map((v) => (
              <li key={v.id} className="py-2 flex justify-between items-center">
                <span className="text-sm text-gray-900">
                  {v.viewer?.email || `User #${v.viewer_id}`}
                </span>
                <Button
                  variant="danger"
                  size="sm"
                  onClick={() => handleRevoke(v.viewer_id)}
                >
                  Revoke
                </Button>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Owners (whose data I can see) */}
      {owners.length > 0 && (
        <div>
          <h3 className="text-sm font-medium text-gray-700 mb-2">
            I can view data from
          </h3>
          <ul className="divide-y divide-gray-200">
            {owners.map((o) => (
              <li key={o.id} className="py-2 text-sm text-gray-900">
                {o.owner?.email || `User #${o.owner_id}`}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
