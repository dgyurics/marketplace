import type { App, DirectiveBinding } from 'vue'

import { useAuthStore } from '@/store/auth'
import type { Role } from '@/types/user'

interface AuthDirectiveBinding extends DirectiveBinding {
  value: Role
}

function updateElement(el: HTMLElement, binding: AuthDirectiveBinding) {
  const authStore = useAuthStore()
  const requiredRole = binding.value
  const hasPermission = authStore.hasMinimumRole(requiredRole)

  if (!hasPermission) {
    el.setAttribute('disabled', 'true')
  } else {
    el.removeAttribute('disabled')
  }
}

export default {
  install(app: App) {
    app.directive('auth', {
      mounted(el: HTMLElement, binding: AuthDirectiveBinding) {
        updateElement(el, binding)
      },

      updated(el: HTMLElement, binding: AuthDirectiveBinding) {
        updateElement(el, binding)
      },
    })
  },
}
