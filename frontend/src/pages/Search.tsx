import { useSearchParams } from 'react-router-dom'
import { SearchResult } from '@/components/SearchResult'

export default function Search() {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q') || ''

  if (!query) {
    return (
      <div className="flex items-center justify-center h-64 text-[var(--color-text-secondary)]">
        <p className="text-lg">请输入搜索内容</p>
      </div>
    )
  }

  return <SearchResult query={query} />
}
