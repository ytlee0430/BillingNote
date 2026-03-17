import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { categoryKeywordsApi, CategoryKeyword } from '@/api/categoryKeywords'
import { categoriesApi } from '@/api/categories'
import { Button } from '@/components/common/Button'

export const CategoryKeywords = () => {
  const queryClient = useQueryClient()
  const [newKeyword, setNewKeyword] = useState('')
  const [selectedCategoryId, setSelectedCategoryId] = useState<number | null>(null)
  const [message, setMessage] = useState<string | null>(null)

  const { data: keywords = [], isLoading } = useQuery({
    queryKey: ['category-keywords'],
    queryFn: () => categoryKeywordsApi.list(),
  })

  const { data: categories = [] } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.getAll(),
  })

  const expenseCategories = categories.filter((c) => c.type === 'expense')

  // Group keywords by category
  const grouped = useMemo(() => {
    const map = new Map<number, { category: CategoryKeyword['category']; keywords: CategoryKeyword[] }>()
    for (const kw of keywords) {
      if (!kw.category) continue
      const existing = map.get(kw.category_id)
      if (existing) {
        existing.keywords.push(kw)
      } else {
        map.set(kw.category_id, { category: kw.category, keywords: [kw] })
      }
    }
    return map
  }, [keywords])

  const addMutation = useMutation({
    mutationFn: ({ categoryId, keyword }: { categoryId: number; keyword: string }) =>
      categoryKeywordsApi.add(categoryId, keyword),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['category-keywords'] })
      setNewKeyword('')
      setMessage(null)
    },
    onError: () => setMessage('新增失敗'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => categoryKeywordsApi.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['category-keywords'] })
    },
  })

  const initMutation = useMutation({
    mutationFn: () => categoryKeywordsApi.initDefaults(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['category-keywords'] })
      setMessage('已載入預設關鍵字')
    },
    onError: () => setMessage('載入失敗'),
  })

  const reclassifyMutation = useMutation({
    mutationFn: () => categoryKeywordsApi.reclassify(),
    onSuccess: (data) => {
      setMessage(`已重新分類 ${data.updated} 筆交易`)
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
    },
    onError: () => setMessage('重新分類失敗'),
  })

  const handleAdd = () => {
    if (!selectedCategoryId || !newKeyword.trim()) return
    addMutation.mutate({ categoryId: selectedCategoryId, keyword: newKeyword.trim() })
  }

  if (isLoading) {
    return (
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">分類關鍵字規則</h2>
        <div className="text-gray-500">載入中...</div>
      </div>
    )
  }

  return (
    <div className="bg-white shadow rounded-lg p-6 mb-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold">分類關鍵字規則</h2>
        {keywords.length === 0 && (
          <Button
            size="sm"
            onClick={() => initMutation.mutate()}
            loading={initMutation.isPending}
          >
            載入預設關鍵字
          </Button>
        )}
      </div>

      <p className="text-sm text-gray-600 mb-4">
        設定關鍵字自動分類交易。匯入交易時，描述中包含關鍵字的交易將自動歸類到對應分類。
      </p>

      {message && (
        <div className="mb-4 p-3 rounded bg-blue-100 text-blue-700 text-sm">{message}</div>
      )}

      {/* Add new keyword */}
      <div className="flex gap-2 mb-6">
        <select
          value={selectedCategoryId ?? ''}
          onChange={(e) => setSelectedCategoryId(Number(e.target.value) || null)}
          className="border border-gray-300 rounded-md px-3 py-2 text-sm"
        >
          <option value="">選擇分類</option>
          {expenseCategories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.icon} {c.name}
            </option>
          ))}
        </select>
        <input
          type="text"
          value={newKeyword}
          onChange={(e) => setNewKeyword(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
          placeholder="輸入關鍵字（不分大小寫）"
          className="flex-1 border border-gray-300 rounded-md px-3 py-2 text-sm"
        />
        <Button
          size="sm"
          onClick={handleAdd}
          disabled={!selectedCategoryId || !newKeyword.trim()}
          loading={addMutation.isPending}
        >
          新增
        </Button>
      </div>

      {/* Keyword list grouped by category */}
      <div className="space-y-4">
        {expenseCategories
          .filter((c) => grouped.has(c.id))
          .map((cat) => {
            const group = grouped.get(cat.id)!
            return (
              <div key={cat.id} className="border rounded-lg p-3">
                <div className="flex items-center gap-2 mb-2">
                  <span>{cat.icon}</span>
                  <span className="font-medium text-sm">{cat.name}</span>
                  <span className="text-xs text-gray-400">({group.keywords.length})</span>
                </div>
                <div className="flex flex-wrap gap-2">
                  {group.keywords.map((kw) => (
                    <span
                      key={kw.id}
                      className="inline-flex items-center gap-1 px-2 py-1 bg-gray-100 rounded-full text-sm"
                    >
                      {kw.keyword}
                      <button
                        onClick={() => deleteMutation.mutate(kw.id)}
                        className="text-gray-400 hover:text-red-500 ml-0.5"
                        title="刪除"
                      >
                        &times;
                      </button>
                    </span>
                  ))}
                </div>
              </div>
            )
          })}
      </div>

      {keywords.length > 0 && (
        <div className="mt-4 pt-4 border-t flex items-center gap-3 flex-wrap">
          <Button
            onClick={() => reclassifyMutation.mutate()}
            loading={reclassifyMutation.isPending}
            size="sm"
          >
            套用規則到未分類交易
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => initMutation.mutate()}
            loading={initMutation.isPending}
          >
            重新載入預設關鍵字
          </Button>
          <span className="text-xs text-gray-500">（不會覆蓋已有的規則）</span>
        </div>
      )}
    </div>
  )
}
