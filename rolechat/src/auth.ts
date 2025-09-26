import { reactive, computed } from 'vue'

interface AuthUser { username: string; email?: string; nickname?: string }

const authState = reactive<{ token: string; user: AuthUser | null}>({
  token: localStorage.getItem('accessToken') || '',
  user: (() => { try { return JSON.parse(localStorage.getItem('user')||'null') } catch { return null } })()
})

function setAuth(token: string, user?: AuthUser | null) {
  authState.token = token
  localStorage.setItem('accessToken', token)
  if (user) {
    authState.user = user
    localStorage.setItem('user', JSON.stringify(user))
  }
}

function clearAuth() {
  authState.token = ''
  authState.user = null
  localStorage.removeItem('accessToken')
  localStorage.removeItem('user')
}

const isLoggedIn = computed(() => !!authState.token)

export { authState, setAuth, clearAuth, isLoggedIn }
