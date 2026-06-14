import { CINEMA_HALLS, SHOWTIME_SLOTS } from '@/constants'

export type MovieFormValues = {
  title: string
  description: string
  genre: string
  duration: number
  rating: number
  language: string
  director: string
  cast: string
  posterUrl: string
  releaseDate: string
  isActive: boolean
  halls: string[]
  slots: string[]
}

export function createDefaultMovieForm(): MovieFormValues {
  return {
    title: '',
    description: '',
    genre: '',
    duration: 120,
    rating: 8.0,
    language: 'TH/EN',
    director: '',
    cast: '',
    posterUrl: '',
    releaseDate: new Date().toISOString().substring(0, 10),
    isActive: true,
    halls: CINEMA_HALLS.map((hall) => hall.value) as string[],
    slots: SHOWTIME_SLOTS.map((slot) => slot.value) as string[],
  }
}

export function parseCommaSeparatedList(value: string) {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}
