import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  uploadPDFs,
  importTransactions,
  ParsedTransaction,
  UploadResult,
} from '../api/upload'
import { formatCurrency, formatDate } from '../utils/format'

export default function Upload() {
  const navigate = useNavigate()
  const [files, setFiles] = useState<File[]>([])
  const [uploading, setUploading] = useState(false)
  const [results, setResults] = useState<UploadResult[]>([])
  const [selectedTransactions, setSelectedTransactions] = useState<
    Map<string, ParsedTransaction[]>
  >(new Map())
  const [importing, setImporting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    const droppedFiles = Array.from(e.dataTransfer.files).filter(
      file => file.type === 'application/pdf'
    )
    setFiles(prev => [...prev, ...droppedFiles])
  }, [])

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const selectedFiles = Array.from(e.target.files).filter(
        file => file.type === 'application/pdf'
      )
      setFiles(prev => [...prev, ...selectedFiles])
    }
  }

  const removeFile = (index: number) => {
    setFiles(prev => prev.filter((_, i) => i !== index))
  }

  const handleUpload = async () => {
    if (files.length === 0) return

    setUploading(true)
    setError(null)

    try {
      const response = await uploadPDFs(files)
      setResults(response.results)

      // Initialize selected transactions (exclude duplicates)
      const initialSelected = new Map<string, ParsedTransaction[]>()
      response.results.forEach(result => {
        if (result.transactions) {
          const nonDuplicates = result.transactions.filter(t => !t.is_duplicate)
          initialSelected.set(result.filename, nonDuplicates)
        }
      })
      setSelectedTransactions(initialSelected)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Upload failed')
    } finally {
      setUploading(false)
    }
  }

  const toggleTransaction = (
    filename: string,
    transaction: ParsedTransaction
  ) => {
    setSelectedTransactions(prev => {
      const newMap = new Map(prev)
      const current = newMap.get(filename) || []
      const index = current.findIndex(
        t =>
          t.date === transaction.date &&
          t.amount === transaction.amount &&
          t.description === transaction.description
      )

      if (index >= 0) {
        newMap.set(
          filename,
          current.filter((_, i) => i !== index)
        )
      } else {
        newMap.set(filename, [...current, transaction])
      }

      return newMap
    })
  }

  const isSelected = (filename: string, transaction: ParsedTransaction) => {
    const selected = selectedTransactions.get(filename) || []
    return selected.some(
      t =>
        t.date === transaction.date &&
        t.amount === transaction.amount &&
        t.description === transaction.description
    )
  }

  const handleImport = async () => {
    const allTransactions: ParsedTransaction[] = []
    selectedTransactions.forEach(transactions => {
      allTransactions.push(...transactions)
    })

    if (allTransactions.length === 0) {
      setError('No transactions selected')
      return
    }

    setImporting(true)
    setError(null)

    try {
      await importTransactions(allTransactions)
      navigate('/transactions')
    } catch (err: any) {
      setError(err.response?.data?.error || 'Import failed')
    } finally {
      setImporting(false)
    }
  }

  const totalSelected = Array.from(selectedTransactions.values()).reduce(
    (sum, transactions) => sum + transactions.length,
    0
  )

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold mb-6">Upload PDF Statements</h1>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      {/* Upload Area */}
      {results.length === 0 && (
        <div className="mb-8">
          <div
            onDrop={handleDrop}
            onDragOver={e => e.preventDefault()}
            className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:border-blue-500 transition-colors"
          >
            <input
              type="file"
              accept=".pdf"
              multiple
              onChange={handleFileSelect}
              className="hidden"
              id="file-upload"
            />
            <label
              htmlFor="file-upload"
              className="cursor-pointer text-gray-600"
            >
              <div className="text-4xl mb-2">📄</div>
              <p className="text-lg">Drop PDF files here or click to select</p>
              <p className="text-sm text-gray-400 mt-2">
                Supports: Credit card statements from Cathay, Taishin, Fubon
              </p>
            </label>
          </div>

          {/* File List */}
          {files.length > 0 && (
            <div className="mt-4">
              <h3 className="font-semibold mb-2">
                Selected Files ({files.length})
              </h3>
              <ul className="space-y-2">
                {files.map((file, index) => (
                  <li
                    key={index}
                    className="flex items-center justify-between bg-gray-50 p-3 rounded"
                  >
                    <span>
                      {file.name} ({(file.size / 1024).toFixed(1)} KB)
                    </span>
                    <button
                      onClick={() => removeFile(index)}
                      className="text-red-500 hover:text-red-700"
                    >
                      Remove
                    </button>
                  </li>
                ))}
              </ul>

              <button
                onClick={handleUpload}
                disabled={uploading}
                className="mt-4 bg-blue-500 text-white px-6 py-2 rounded hover:bg-blue-600 disabled:opacity-50"
              >
                {uploading ? 'Parsing...' : 'Parse PDFs'}
              </button>
            </div>
          )}
        </div>
      )}

      {/* Results */}
      {results.length > 0 && (
        <div>
          {results.map((result, index) => (
            <div key={index} className="mb-8 bg-white rounded-lg shadow p-6">
              <div className="flex justify-between items-center mb-4">
                <div>
                  <h3 className="font-semibold text-lg">{result.filename}</h3>
                  {result.bank && (
                    <span className="text-sm text-gray-500">
                      Bank: {result.bank}
                    </span>
                  )}
                </div>
                {result.total_amount > 0 && (
                  <span className="text-lg font-semibold">
                    Total: {formatCurrency(result.total_amount)}
                  </span>
                )}
              </div>

              {result.error ? (
                <div className="text-red-500">{result.error}</div>
              ) : result.transactions?.length > 0 ? (
                <table className="w-full">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left py-2 w-8">
                        <input type="checkbox" disabled />
                      </th>
                      <th className="text-left py-2">Date</th>
                      <th className="text-left py-2">Description</th>
                      <th className="text-right py-2">Amount</th>
                      <th className="text-center py-2">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {result.transactions.map((t, tIndex) => (
                      <tr
                        key={tIndex}
                        className={`border-b ${
                          t.is_duplicate ? 'bg-yellow-50' : ''
                        }`}
                      >
                        <td className="py-2">
                          <input
                            type="checkbox"
                            checked={isSelected(result.filename, t)}
                            onChange={() =>
                              toggleTransaction(result.filename, t)
                            }
                            disabled={t.is_duplicate}
                          />
                        </td>
                        <td className="py-2">{formatDate(t.date)}</td>
                        <td className="py-2">{t.description}</td>
                        <td className="py-2 text-right">
                          {formatCurrency(t.amount)}
                        </td>
                        <td className="py-2 text-center">
                          {t.is_duplicate && (
                            <span className="text-yellow-600 text-sm">
                              Duplicate
                            </span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className="text-gray-500">No transactions found</div>
              )}
            </div>
          ))}

          {/* Import Button */}
          <div className="flex justify-between items-center bg-gray-100 p-4 rounded">
            <span>
              Selected: {totalSelected} transaction
              {totalSelected !== 1 ? 's' : ''}
            </span>
            <div className="space-x-4">
              <button
                onClick={() => {
                  setResults([])
                  setFiles([])
                  setSelectedTransactions(new Map())
                }}
                className="px-4 py-2 border rounded hover:bg-gray-200"
              >
                Start Over
              </button>
              <button
                onClick={handleImport}
                disabled={importing || totalSelected === 0}
                className="bg-green-500 text-white px-6 py-2 rounded hover:bg-green-600 disabled:opacity-50"
              >
                {importing ? 'Importing...' : `Import ${totalSelected} Transactions`}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
