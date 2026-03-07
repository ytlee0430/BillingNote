import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { gmailApi } from '@/api/gmail'
import { GmailSettingsInput } from '@/types/gmail'
import { Button } from '@/components/common/Button'
import { formatDateTime } from '@/utils/format'

export const GmailConnect = () => {
  const queryClient = useQueryClient()
  const [scanMessage, setScanMessage] = useState<string | null>(null)
  const [senderInput, setSenderInput] = useState('')
  const [subjectInput, setSubjectInput] = useState('')
  const [settingsMessage, setSettingsMessage] = useState<string | null>(null)

  const { data: status, isLoading: statusLoading } = useQuery({
    queryKey: ['gmail-status'],
    queryFn: () => gmailApi.getStatus(),
  })

  const { data: settings } = useQuery({
    queryKey: ['gmail-settings'],
    queryFn: () => gmailApi.getSettings(),
    enabled: !!status?.connected,
  })

  useEffect(() => {
    if (settings) {
      setSenderInput(settings.sender_keywords?.join(', ') || '')
      setSubjectInput(settings.subject_keywords?.join(', ') || '')
    }
  }, [settings])

  const connectMutation = useMutation({
    mutationFn: () => gmailApi.getAuthURL(),
    onSuccess: (data) => {
      window.location.href = data.url
    },
  })

  const disconnectMutation = useMutation({
    mutationFn: () => gmailApi.disconnect(),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['gmail-status'] })
      await queryClient.invalidateQueries({ queryKey: ['gmail-settings'] })
    },
  })

  const scanMutation = useMutation({
    mutationFn: () => gmailApi.triggerScan(),
    onSuccess: async (result) => {
      setScanMessage(
        `Scan complete: ${result.scanned} emails scanned, ${result.downloaded} PDFs downloaded`
      )
      await queryClient.invalidateQueries({ queryKey: ['gmail-status'] })
    },
    onError: () => {
      setScanMessage('Scan failed. Please try again.')
    },
  })

  const updateSettingsMutation = useMutation({
    mutationFn: (data: GmailSettingsInput) => gmailApi.updateSettings(data),
    onSuccess: async () => {
      setSettingsMessage('Settings saved successfully')
      await queryClient.invalidateQueries({ queryKey: ['gmail-settings'] })
    },
    onError: () => {
      setSettingsMessage('Failed to save settings')
    },
  })

  const handleDisconnect = () => {
    if (window.confirm('Are you sure you want to disconnect Gmail?')) {
      disconnectMutation.mutate()
    }
  }

  const handleSaveSettings = () => {
    setSettingsMessage(null)
    const senderKeywords = senderInput
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    const subjectKeywords = subjectInput
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)

    updateSettingsMutation.mutate({
      enabled: settings?.enabled ?? true,
      sender_keywords: senderKeywords,
      subject_keywords: subjectKeywords,
      require_attachment: settings?.require_attachment ?? true,
    })
  }

  const handleToggleEnabled = (enabled: boolean) => {
    updateSettingsMutation.mutate({ enabled })
  }

  if (statusLoading) {
    return (
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Gmail Integration</h2>
        <div className="text-gray-500">Loading...</div>
      </div>
    )
  }

  return (
    <div className="bg-white shadow rounded-lg p-6 mb-6">
      <h2 className="text-xl font-semibold mb-4">Gmail Integration</h2>

      {!status?.connected ? (
        <div>
          <p className="text-sm text-gray-600 mb-4">
            Connect your Gmail to automatically scan for credit card statement PDFs.
          </p>
          <Button
            onClick={() => connectMutation.mutate()}
            loading={connectMutation.isPending}
          >
            Connect Gmail
          </Button>
        </div>
      ) : (
        <div className="space-y-6">
          {/* Connection Status */}
          <div className="flex items-center justify-between bg-green-50 rounded-lg p-4">
            <div>
              <div className="text-sm font-medium text-green-800">Connected</div>
              <div className="text-sm text-green-700">{status.email}</div>
              {status.last_scan_at && (
                <div className="text-xs text-green-600 mt-1">
                  Last scan: {formatDateTime(status.last_scan_at)}
                </div>
              )}
            </div>
            <Button
              variant="danger"
              size="sm"
              onClick={handleDisconnect}
              loading={disconnectMutation.isPending}
            >
              Disconnect
            </Button>
          </div>

          {/* Enable Toggle */}
          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="gmail-enabled"
              checked={settings?.enabled ?? false}
              onChange={(e) => handleToggleEnabled(e.target.checked)}
              className="h-4 w-4 text-blue-600 rounded border-gray-300"
            />
            <label htmlFor="gmail-enabled" className="text-sm font-medium text-gray-700">
              Enable Gmail auto-scan
            </label>
          </div>

          {/* Scan Rules */}
          <div className="space-y-4">
            <h3 className="text-sm font-semibold text-gray-700">Scan Rules</h3>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Sender keywords (comma separated)
              </label>
              <input
                type="text"
                value={senderInput}
                onChange={(e) => setSenderInput(e.target.value)}
                placeholder="credit, statement, 帳單"
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Subject keywords (comma separated)
              </label>
              <input
                type="text"
                value={subjectInput}
                onChange={(e) => setSubjectInput(e.target.value)}
                placeholder="帳單, 電子帳單, statement"
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
              />
            </div>

            <div className="flex items-center gap-3">
              <input
                type="checkbox"
                id="require-attachment"
                checked={settings?.require_attachment ?? true}
                onChange={(e) =>
                  updateSettingsMutation.mutate({ require_attachment: e.target.checked })
                }
                className="h-4 w-4 text-blue-600 rounded border-gray-300"
              />
              <label htmlFor="require-attachment" className="text-sm text-gray-700">
                Only emails with attachments
              </label>
            </div>

            {settingsMessage && (
              <div
                className={`p-3 rounded text-sm ${
                  settingsMessage.includes('success')
                    ? 'bg-green-100 text-green-700'
                    : 'bg-red-100 text-red-700'
                }`}
              >
                {settingsMessage}
              </div>
            )}

            <Button
              variant="secondary"
              onClick={handleSaveSettings}
              loading={updateSettingsMutation.isPending}
            >
              Save Settings
            </Button>
          </div>

          {/* Scan Actions */}
          <div className="border-t pt-4">
            <div className="flex items-center gap-4">
              <Button
                onClick={() => {
                  setScanMessage(null)
                  scanMutation.mutate()
                }}
                loading={scanMutation.isPending}
              >
                {scanMutation.isPending ? 'Scanning...' : 'Scan Now'}
              </Button>
            </div>
            {scanMessage && (
              <div
                className={`mt-3 p-3 rounded text-sm ${
                  scanMessage.includes('failed')
                    ? 'bg-red-100 text-red-700'
                    : 'bg-green-100 text-green-700'
                }`}
              >
                {scanMessage}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
