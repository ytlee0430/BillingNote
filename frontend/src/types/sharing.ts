export interface PairingCodeResponse {
  code: string
}

export interface SharedAccess {
  id: number
  owner_id: number
  viewer_id: number
  created_at: string
  owner?: { id: number; email: string; name: string }
  viewer?: { id: number; email: string; name: string }
}

export interface ConnectionsResponse {
  viewers: SharedAccess[]
  owners: SharedAccess[]
}
