import { HashRouter, Routes, Route } from 'react-router-dom'
import { MainLayout } from './components/layout/MainLayout'
import Home from './pages/Home'
import Following from './pages/Following'
import FM from './pages/FM'
import Likes from './pages/Likes'
import History from './pages/History'
import Cloud from './pages/Cloud'
import Search from './pages/Search'

function App() {
  return (
    <HashRouter>
      <Routes>
        <Route element={<MainLayout />}>
          <Route index element={<Home />} />
          <Route path="following" element={<Following />} />
          <Route path="fm" element={<FM />} />
          <Route path="likes" element={<Likes />} />
          <Route path="history" element={<History />} />
          <Route path="cloud" element={<Cloud />} />
          <Route path="search" element={<Search />} />
        </Route>
      </Routes>
    </HashRouter>
  )
}

export default App
