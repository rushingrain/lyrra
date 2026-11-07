import { HashRouter, Routes, Route } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'
import { MainLayout } from '@/layouts/MainLayout'
import Home from './pages/Home'
import Following from './pages/Following'
import FM from './pages/FM'
import Likes from './pages/Likes'
import History from './pages/History'
import Cloud from './pages/Cloud'
import Search from './pages/Search'
import SongList from './pages/SongList'
import Playlist from './pages/Playlist'
import Playlists from './pages/Playlists'
import Album from './pages/Album'

function App() {
  return (
    <>
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
            <Route path="playlists" element={<Playlists />} />
            <Route path="songs/:kind" element={<SongList />} />
            <Route path="songs/playlist/:globalId" element={<Playlist />} />
            <Route path="songs/album/:albumId" element={<Album />} />
          </Route>
        </Routes>
      </HashRouter>
      {/* 全局错误提示容器：主题色与 style.css 的 CSS 变量保持一致 */}
      <Toaster
        position="top-center"
        toastOptions={{
          // 停留时长（ms）
          duration: 2000,
          style: {
            background: 'var(--color-background)',
            color: 'var(--color-text-primary)',
            border: '1px solid var(--color-border)',
            fontSize: '14px',
            borderRadius: '10px',
            padding: '10px 16px',
          },
          error: {
            iconTheme: { primary: '#EF4444', secondary: '#FFFFFF' },
          },
        }}
      />
    </>
  )
}

export default App
