import { describe, expect, it } from 'vitest'
import { createDefaultMovieForm, parseCommaSeparatedList } from '@/utils/movieForm'

describe('movieForm utilities', () => {
  describe('createDefaultMovieForm', () => {
    it('returns default movie form values', () => {
      const defaultForm = createDefaultMovieForm()
      expect(defaultForm).toMatchObject({
        title: '',
        description: '',
        genre: '',
        duration: 120,
        rating: 8.0,
        language: 'TH/EN',
        director: '',
        cast: '',
        posterUrl: '',
        isActive: true,
      })
      expect(defaultForm.releaseDate).toMatch(/^\d{4}-\d{2}-\d{2}$/)
      expect(defaultForm.halls).toContain('Hall A')
      expect(defaultForm.halls).toContain('Hall B')
      expect(defaultForm.slots).toContain('Morning')
      expect(defaultForm.slots).toContain('Night')
    })
  })

  describe('parseCommaSeparatedList', () => {
    it('parses standard list', () => {
      expect(parseCommaSeparatedList('Action, Drama, Sci-Fi')).toEqual([
        'Action',
        'Drama',
        'Sci-Fi',
      ])
    })

    it('trims leading/trailing spaces from items', () => {
      expect(parseCommaSeparatedList('  Action  ,   Drama   ')).toEqual([
        'Action',
        'Drama',
      ])
    })

    it('removes empty items', () => {
      expect(parseCommaSeparatedList('Action, , Drama, ,')).toEqual([
        'Action',
        'Drama',
      ])
    })

    it('returns empty array for empty string', () => {
      expect(parseCommaSeparatedList('')).toEqual([])
    })

    it('returns single item for input without commas', () => {
      expect(parseCommaSeparatedList('Action')).toEqual(['Action'])
    })
  })
})
