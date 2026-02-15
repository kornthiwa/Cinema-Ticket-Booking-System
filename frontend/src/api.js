const base = import.meta.env.VITE_API_URL || ''

function headers() {
  const token = localStorage.getItem('token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export async function login(body) {
  const r = await fetch(`${base}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: body.email, password: body.password }),
  })
  if (!r.ok) {
    const e = await r.json().catch(() => ({}))
    throw new Error(e.error || r.statusText)
  }
  return r.json()
}

export async function register(body) {
  const r = await fetch(`${base}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email: body.email,
      name: body.name || body.email,
      password: body.password,
    }),
  })
  if (!r.ok) {
    const e = await r.json().catch(() => ({}))
    throw new Error(e.error || r.statusText)
  }
  return r.json()
}

export async function adminLogin(body) {
  const r = await fetch(`${base}/admin/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: body.email, password: body.password }),
  })
  if (!r.ok) {
    const e = await r.json().catch(() => ({}))
    throw new Error(e.error || r.statusText)
  }
  return r.json()
}

export async function getScreenings() {
  const r = await fetch(`${base}/api/screenings`, { headers: headers() })
  if (!r.ok) throw new Error('Failed to load screenings')
  return r.json()
}

export async function getScreening(id) {
  const r = await fetch(`${base}/api/screenings/${id}`, { headers: headers() })
  if (!r.ok) throw new Error('Screening not found')
  return r.json()
}

export async function getSeatMap(id) {
  const r = await fetch(`${base}/api/screenings/${id}/seats`, { headers: headers() })
  const data = await r.json().catch(() => ({}))
  if (!r.ok) {
    if (r.status === 401) throw new Error('Please log in again')
    throw new Error(data.error || 'Failed to load seats')
  }
  return data
}

/** รายการที่นั่งที่ถูกล็อก/จอง พร้อมเวลา (สำหรับหน้า ScreeningList) */
export async function getSeatDetails(id) {
  const r = await fetch(`${base}/api/screenings/${id}/seat-details`, { headers: headers() })
  const data = await r.json().catch(() => ({}))
  if (!r.ok) {
    if (r.status === 401) throw new Error('Please log in again')
    throw new Error(data.error || 'Failed to load seat details')
  }
  return data
}

export async function lockSeat(screeningId, row, col) {
  const r = await fetch(`${base}/api/screenings/${screeningId}/lock`, {
    method: 'POST',
    headers: headers(),
    body: JSON.stringify({ row, col }),
  })
  const data = await r.json().catch(() => ({}))
  if (!r.ok) throw new Error(data.error || 'Lock failed')
  return data
}

export async function confirmPayment(bookingId) {
  const r = await fetch(`${base}/api/bookings/confirm`, {
    method: 'POST',
    headers: headers(),
    body: JSON.stringify({ booking_id: bookingId }),
  })
  const data = await r.json().catch(() => ({}))
  if (!r.ok) throw new Error(data.error || 'Confirm failed')
  return data
}

export function wsUrl(screeningId) {
  const token = localStorage.getItem('token')
  const host = (import.meta.env.VITE_WS_URL || base || window.location.origin).replace(/^http/, 'ws')
  const path = `${host.replace(/\/$/, '')}/api/screenings/${screeningId}/ws`
  return token ? `${path}?token=${encodeURIComponent(token)}` : path
}

/** WebSocket URL for Admin (real-time refresh of bookings & audit logs) */
export function wsAdminUrl() {
  const token = localStorage.getItem('token')
  const host = (import.meta.env.VITE_WS_URL || base || window.location.origin).replace(/^http/, 'ws')
  const path = `${host.replace(/\/$/, '')}/admin/ws`
  return token ? `${path}?token=${encodeURIComponent(token)}` : path
}

// Admin
export async function adminBookings(params = {}) {
  const q = new URLSearchParams(params).toString()
  const r = await fetch(`${base}/admin/bookings?${q}`, { headers: headers() })
  if (!r.ok) throw new Error('Failed to load bookings')
  return r.json()
}

export async function adminAuditLogs() {
  const r = await fetch(`${base}/admin/audit-logs`, { headers: headers() })
  if (!r.ok) throw new Error('Failed to load audit logs')
  return r.json()
}

export async function createScreening(body) {
  const r = await fetch(`${base}/admin/screenings`, {
    method: 'POST',
    headers: headers(),
    body: JSON.stringify(body),
  })
  const data = await r.json().catch(() => ({}))
  if (!r.ok) throw new Error(data.error || 'Create failed')
  return data
}
