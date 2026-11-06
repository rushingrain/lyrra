import type { Key, ReactNode } from 'react'

export interface ResultColumn<T> {
  /** 列标识，同时作为表头/单元格的 key */
  key: string
  /** 表头内容 */
  title: ReactNode
  /** 列宽、对齐、截断等布局类，表头和正文共用以保证对齐 */
  className?: string
  /** 单元格渲染 */
  render: (item: T, index: number) => ReactNode
}

interface ResultTableProps<T> {
  columns: ResultColumn<T>[]
  rows: T[]
  rowKey: (item: T, index: number) => Key
  /** 传入后整行可点击，并支持 Enter / 空格触发 */
  onRowClick?: (item: T, index: number) => void
}

/**
 * 通用结果表格：列配置驱动，不含任何业务概念。
 * 行布局为 flex，表头与正文复用同一份列 className（列宽/对齐）保证对齐。
 */
export function ResultTable<T>({ columns, rows, rowKey, onRowClick }: ResultTableProps<T>) {
  return (
    <div className="min-w-0">
      <div className="flex items-center px-3 gap-4 border-b border-[var(--color-border)] pb-2 text-sm text-[var(--color-text-secondary)]">
        {columns.map((col) => (
          <div key={col.key} className={col.className}>
            {col.title}
          </div>
        ))}
      </div>
      <div>
        {rows.map((item, index) => (
          <div
            key={rowKey(item, index)}
            className="flex items-center cursor-pointer gap-4 rounded-md px-3 py-3 transition-colors hover:bg-[var(--color-surface)]"
            role={onRowClick ? 'link' : undefined}
            tabIndex={onRowClick ? 0 : undefined}
            onClick={onRowClick ? () => onRowClick(item, index) : undefined}
            onKeyDown={
              onRowClick
                ? (event) => {
                    // 避免行内其它可交互元素的按键冒泡触发整行跳转
                    if (event.target !== event.currentTarget) return
                    if (event.key === 'Enter' || event.key === ' ') {
                      event.preventDefault()
                      onRowClick(item, index)
                    }
                  }
                : undefined
            }
          >
            {columns.map((col) => (
              <div key={col.key} className={col.className}>
                {col.render(item, index)}
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}
