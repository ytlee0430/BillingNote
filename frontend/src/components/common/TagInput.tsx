import { useState, KeyboardEvent } from 'react'

interface TagInputProps {
  tags: string[]
  onChange: (tags: string[]) => void
  placeholder?: string
}

export const TagInput = ({ tags, onChange, placeholder = 'Add tag...' }: TagInputProps) => {
  const [input, setInput] = useState('')

  const addTag = (tag: string) => {
    const trimmed = tag.trim().toLowerCase()
    if (trimmed && !tags.includes(trimmed)) {
      onChange([...tags, trimmed])
    }
    setInput('')
  }

  const removeTag = (index: number) => {
    onChange(tags.filter((_, i) => i !== index))
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      addTag(input)
    } else if (e.key === 'Backspace' && input === '' && tags.length > 0) {
      removeTag(tags.length - 1)
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-1 border border-gray-300 rounded-md px-2 py-1 min-h-[38px] focus-within:ring-2 focus-within:ring-primary-500 focus-within:border-primary-500">
      {tags.map((tag, i) => (
        <span
          key={i}
          className="inline-flex items-center gap-1 bg-primary-100 text-primary-800 text-sm px-2 py-0.5 rounded-full"
        >
          {tag}
          <button
            type="button"
            onClick={() => removeTag(i)}
            className="text-primary-600 hover:text-primary-800 font-bold"
          >
            x
          </button>
        </span>
      ))}
      <input
        type="text"
        className="flex-1 min-w-[80px] outline-none border-none bg-transparent text-sm py-1"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={() => { if (input) addTag(input) }}
        placeholder={tags.length === 0 ? placeholder : ''}
      />
    </div>
  )
}
