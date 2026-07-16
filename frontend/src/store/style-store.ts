import { create } from "zustand"

const STORAGE_KEY = "resume-app-style"

type StyleValue = "default" | "glass"

interface StyleState {
  style: StyleValue
}

interface StyleActions {
  setStyle: (style: StyleValue) => void
}

type StyleStore = StyleState & StyleActions

function readInitialStyle(): StyleValue {
  if (typeof window === "undefined") return "default"
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === "default" || stored === "glass") return stored
  return "default"
}

function applyStyle(style: StyleValue): void {
  if (typeof document === "undefined") return
  document.documentElement.dataset.style = style
}

function persistStyle(style: StyleValue): void {
  if (typeof window === "undefined") return
  localStorage.setItem(STORAGE_KEY, style)
}

const useStyleStore = create<StyleStore>()((set) => ({
  style: readInitialStyle(),
  setStyle: (style: StyleValue) => {
    persistStyle(style)
    applyStyle(style)
    set({ style })
  },
}))

applyStyle(useStyleStore.getState().style)

export { type StyleState, type StyleActions, type StyleStore, useStyleStore }
