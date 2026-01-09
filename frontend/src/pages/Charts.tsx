import { useState } from 'react'
import { PieChart } from '@/components/charts/PieChart'
import { BarChart } from '@/components/charts/BarChart'
import { useMonthlyStats, useCategoryStats } from '@/hooks/useTransactions'

export const Charts = () => {
  const currentDate = new Date()
  const [selectedYear, setSelectedYear] = useState(currentDate.getFullYear())
  const [selectedMonth, setSelectedMonth] = useState(currentDate.getMonth() + 1)
  const [statsType, setStatsType] = useState<'income' | 'expense'>('expense')

  // Get monthly stats
  const { data: monthlyStats, isLoading: isMonthlyLoading } = useMonthlyStats(
    selectedYear,
    selectedMonth
  )

  // Get category stats for the selected month
  const startDate = new Date(selectedYear, selectedMonth - 1, 1)
    .toISOString()
    .split('T')[0]
  const endDate = new Date(selectedYear, selectedMonth, 0)
    .toISOString()
    .split('T')[0]

  const { data: categoryStats, isLoading: isCategoryLoading } =
    useCategoryStats(startDate, endDate, statsType)

  const years = Array.from({ length: 5 }, (_, i) => currentDate.getFullYear() - i)
  const months = [
    { value: 1, label: 'January' },
    { value: 2, label: 'February' },
    { value: 3, label: 'March' },
    { value: 4, label: 'April' },
    { value: 5, label: 'May' },
    { value: 6, label: 'June' },
    { value: 7, label: 'July' },
    { value: 8, label: 'August' },
    { value: 9, label: 'September' },
    { value: 10, label: 'October' },
    { value: 11, label: 'November' },
    { value: 12, label: 'December' },
  ]

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Statistics</h1>

      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-lg font-semibold mb-4">Filter Period</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Year
            </label>
            <select
              className="w-full border border-gray-300 rounded-md px-3 py-2"
              value={selectedYear}
              onChange={(e) => setSelectedYear(parseInt(e.target.value))}
            >
              {years.map((year) => (
                <option key={year} value={year}>
                  {year}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Month
            </label>
            <select
              className="w-full border border-gray-300 rounded-md px-3 py-2"
              value={selectedMonth}
              onChange={(e) => setSelectedMonth(parseInt(e.target.value))}
            >
              {months.map((month) => (
                <option key={month.value} value={month.value}>
                  {month.label}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {isMonthlyLoading ? (
        <div className="bg-white shadow rounded-lg p-8 text-center mb-6">
          <div className="text-gray-500">Loading monthly statistics...</div>
        </div>
      ) : monthlyStats ? (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
          <div className="bg-green-50 shadow rounded-lg p-6">
            <div className="text-sm font-medium text-green-600 mb-2">
              Total Income
            </div>
            <div className="text-3xl font-bold text-green-700">
              ${monthlyStats.income?.toFixed(2) || '0.00'}
            </div>
          </div>

          <div className="bg-red-50 shadow rounded-lg p-6">
            <div className="text-sm font-medium text-red-600 mb-2">
              Total Expense
            </div>
            <div className="text-3xl font-bold text-red-700">
              ${monthlyStats.expense?.toFixed(2) || '0.00'}
            </div>
          </div>

          <div className="bg-blue-50 shadow rounded-lg p-6">
            <div className="text-sm font-medium text-blue-600 mb-2">
              Balance
            </div>
            <div
              className={`text-3xl font-bold ${
                (monthlyStats.balance || 0) >= 0
                  ? 'text-blue-700'
                  : 'text-red-700'
              }`}
            >
              ${monthlyStats.balance?.toFixed(2) || '0.00'}
            </div>
          </div>
        </div>
      ) : null}

      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold">Category Breakdown</h2>
          <div className="flex space-x-2">
            <button
              className={`px-4 py-2 rounded-md ${
                statsType === 'expense'
                  ? 'bg-red-100 text-red-700 font-semibold'
                  : 'bg-gray-100 text-gray-700'
              }`}
              onClick={() => setStatsType('expense')}
            >
              Expenses
            </button>
            <button
              className={`px-4 py-2 rounded-md ${
                statsType === 'income'
                  ? 'bg-green-100 text-green-700 font-semibold'
                  : 'bg-gray-100 text-gray-700'
              }`}
              onClick={() => setStatsType('income')}
            >
              Income
            </button>
          </div>
        </div>

        {isCategoryLoading ? (
          <div className="text-center py-8 text-gray-500">
            Loading category statistics...
          </div>
        ) : categoryStats && categoryStats.length > 0 ? (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div>
              <h3 className="text-md font-medium mb-4 text-center">
                Distribution
              </h3>
              <PieChart data={categoryStats} type={statsType} />
            </div>
            <div>
              <h3 className="text-md font-medium mb-4 text-center">
                Comparison
              </h3>
              <BarChart data={categoryStats} type={statsType} />
            </div>
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            No {statsType} data available for this period
          </div>
        )}
      </div>
    </div>
  )
}
