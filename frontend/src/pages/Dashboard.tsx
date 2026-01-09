import React, { useState } from 'react'
import { useMonthlyStats } from '@/hooks/useTransactions'
import { formatCurrency } from '@/utils/format'

export const Dashboard: React.FC = () => {
  const now = new Date()
  const [year] = useState(now.getFullYear())
  const [month] = useState(now.getMonth() + 1)

  const { data: stats, isLoading } = useMonthlyStats(year, month)

  if (isLoading) {
    return <div className="flex justify-center items-center h-64">載入中...</div>
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">儀表板</h1>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Income Card */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">本月收入</p>
              <p className="text-2xl font-bold text-green-600">
                {formatCurrency(stats?.income || 0)}
              </p>
            </div>
            <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center">
              <span className="text-2xl">💰</span>
            </div>
          </div>
        </div>

        {/* Expense Card */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">本月支出</p>
              <p className="text-2xl font-bold text-red-600">
                {formatCurrency(stats?.expense || 0)}
              </p>
            </div>
            <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
              <span className="text-2xl">💸</span>
            </div>
          </div>
        </div>

        {/* Balance Card */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">結餘</p>
              <p
                className={`text-2xl font-bold ${
                  (stats?.balance || 0) >= 0 ? 'text-blue-600' : 'text-red-600'
                }`}
              >
                {formatCurrency(stats?.balance || 0)}
              </p>
            </div>
            <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
              <span className="text-2xl">📊</span>
            </div>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold mb-4">
          {year} 年 {month} 月統計
        </h2>
        <p className="text-gray-600">
          查看更多統計資料，請前往「圖表」頁面。
        </p>
      </div>
    </div>
  )
}
