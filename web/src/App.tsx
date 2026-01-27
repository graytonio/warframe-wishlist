import { Routes, Route } from 'react-router-dom'
import { Header } from './components/layout/Header'
import { ProtectedRoute } from './components/auth/ProtectedRoute'
import Search from './pages/Search'
import Login from './pages/Login'
import Wishlist from './pages/Wishlist'

function App() {
  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main>
        <Routes>
          <Route path="/" element={<Search />} />
          <Route path="/login" element={<Login />} />
          <Route
            path="/wishlist"
            element={
              <ProtectedRoute>
                <Wishlist />
              </ProtectedRoute>
            }
          />
        </Routes>
      </main>
    </div>
  )
}

export default App
