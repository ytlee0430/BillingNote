import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { TagInput } from './TagInput'

describe('TagInput', () => {
  it('should render existing tags', () => {
    render(<TagInput tags={['food', 'lunch']} onChange={() => {}} />)
    expect(screen.getByText('food')).toBeInTheDocument()
    expect(screen.getByText('lunch')).toBeInTheDocument()
  })

  it('should add tag on Enter', () => {
    const onChange = vi.fn()
    render(<TagInput tags={[]} onChange={onChange} />)

    const input = screen.getByPlaceholderText('Add tag...')
    fireEvent.change(input, { target: { value: 'new-tag' } })
    fireEvent.keyDown(input, { key: 'Enter' })

    expect(onChange).toHaveBeenCalledWith(['new-tag'])
  })

  it('should add tag on comma', () => {
    const onChange = vi.fn()
    render(<TagInput tags={[]} onChange={onChange} />)

    const input = screen.getByPlaceholderText('Add tag...')
    fireEvent.change(input, { target: { value: 'comma-tag' } })
    fireEvent.keyDown(input, { key: ',' })

    expect(onChange).toHaveBeenCalledWith(['comma-tag'])
  })

  it('should remove tag on x click', () => {
    const onChange = vi.fn()
    render(<TagInput tags={['food', 'lunch']} onChange={onChange} />)

    const removeButtons = screen.getAllByText('x')
    fireEvent.click(removeButtons[0])

    expect(onChange).toHaveBeenCalledWith(['lunch'])
  })

  it('should not add duplicate tags', () => {
    const onChange = vi.fn()
    render(<TagInput tags={['food']} onChange={onChange} />)

    const input = screen.getByRole('textbox')
    fireEvent.change(input, { target: { value: 'food' } })
    fireEvent.keyDown(input, { key: 'Enter' })

    expect(onChange).not.toHaveBeenCalled()
  })

  it('should remove last tag on Backspace when input is empty', () => {
    const onChange = vi.fn()
    render(<TagInput tags={['food', 'lunch']} onChange={onChange} />)

    const input = screen.getByRole('textbox')
    fireEvent.keyDown(input, { key: 'Backspace' })

    expect(onChange).toHaveBeenCalledWith(['food'])
  })

  it('should lowercase tags', () => {
    const onChange = vi.fn()
    render(<TagInput tags={[]} onChange={onChange} />)

    const input = screen.getByPlaceholderText('Add tag...')
    fireEvent.change(input, { target: { value: 'FOOD' } })
    fireEvent.keyDown(input, { key: 'Enter' })

    expect(onChange).toHaveBeenCalledWith(['food'])
  })
})
