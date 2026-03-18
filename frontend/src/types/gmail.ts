export interface GmailStatus {
  connected: boolean
  email?: string
  scopes?: string
  last_scan_at?: string
  connected_at?: string
}

export interface GmailSettings {
  enabled: boolean
  sender_keywords: string[]
  subject_keywords: string[]
  require_attachment: boolean
  last_scan_at?: string
}

export interface GmailSettingsInput {
  enabled?: boolean
  sender_keywords?: string[]
  subject_keywords?: string[]
  require_attachment?: boolean
}

export interface GmailScanResult {
  scanned: number
  downloaded: number
  auto_parsed: number
  imported: number
  failed: number
}

export interface GmailScanHistory {
  id: number
  user_id: number
  scan_at: string
  emails_found: number
  pdfs_downloaded: number
  status: string
  error_message?: string
}

export interface GmailCallbackRequest {
  code: string
  state: string
}
