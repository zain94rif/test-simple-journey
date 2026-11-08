/// <reference types="vue-router" />
import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    guestOnly?: boolean
    requiresAuth?: boolean
    roles?: string[]
    title?: string
  }
}

export {}
