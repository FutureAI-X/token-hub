import { HashRouter, Routes, Route } from 'react-router-dom'
import { Home } from './pages/Home'
import { Login } from './pages/Login'
import { Pricing } from './pages/Pricing'

function App() {
  return (
    <HashRouter>
      <Routes>
        <Route path='/' element={<Home />} />
        <Route path='/login' element={<Login />} />
        <Route path='/pricing' element={<Pricing />} />
      </Routes>
    </HashRouter>
  )
}

export default App
