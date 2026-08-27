import { HashRouter, Routes, Route } from 'react-router-dom'
import { Home } from './pages/Home'
import { Login } from './pages/Login'
import { Pricing } from './pages/Pricing'
import { DashboardLayout } from './pages/dashboard/Layout'
import { Profile } from './pages/dashboard/Profile'
import { Wallet } from './pages/dashboard/Wallet'

function App() {
  return (
    <HashRouter>
      <Routes>
        <Route path='/' element={<Home />} />
        <Route path='/login' element={<Login />} />
        <Route path='/pricing' element={<Pricing />} />

        {/* 控制台 */}
        <Route path='/dashboard' element={<DashboardLayout />}>
          <Route path='profile' element={<Profile />} />
          <Route path='wallet' element={<Wallet />} />
        </Route>
      </Routes>
    </HashRouter>
  )
}

export default App
