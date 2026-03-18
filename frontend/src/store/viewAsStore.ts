import { create } from 'zustand'

interface ViewAsState {
  viewAsUserId: number | null
  viewAsEmail: string | null
  setViewAs: (userId: number | null, email?: string | null) => void
  clearViewAs: () => void
  isViewingOther: boolean
}

export const useViewAsStore = create<ViewAsState>((set) => ({
  viewAsUserId: null,
  viewAsEmail: null,
  isViewingOther: false,

  setViewAs: (userId, email = null) => {
    set({
      viewAsUserId: userId,
      viewAsEmail: email,
      isViewingOther: userId !== null,
    })
  },

  clearViewAs: () => {
    set({
      viewAsUserId: null,
      viewAsEmail: null,
      isViewingOther: false,
    })
  },
}))
