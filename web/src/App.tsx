import { HashRouter, Routes, Route, useLocation } from 'react-router-dom'
import { useEffect } from 'react'

function ScrollToTop() {
  const { pathname } = useLocation()
  useEffect(() => {
    window.scrollTo(0, 0)
  }, [pathname])
  return null
}
import { Home } from './pages/Home'
import { Login } from './pages/Login'
import { Pricing } from './pages/Pricing'
import { DashboardLayout } from './pages/dashboard/Layout'
import { Profile } from './pages/dashboard/Profile'
import { Wallet } from './pages/dashboard/Wallet'
import { AdminUsers } from './pages/admin/Users'

function App() {
  return (
    <HashRouter>
      <ScrollToTop />
      <Routes>
        <Route path='/' element={<Home />} />
        <Route path='/login' element={<Login />} />
        <Route path='/pricing' element={<Pricing />} />

        {/* 控制台 */}
        <Route path='/dashboard' element={<DashboardLayout />}>
          <Route path='profile' element={<Profile />} />
          <Route path='wallet' element={<Wallet />} />
          <Route path='admin/users' element={<AdminUsers />} />
        </Route>
      </Routes>
    </HashRouter>
  )
}

export default App
