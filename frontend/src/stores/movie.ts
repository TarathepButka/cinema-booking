import { defineStore } from 'pinia'
import { ref } from 'vue'
import { movieApi } from '@/api'

export interface Movie {
  id: string
  title: string
  description: string
  genre: string[]
  duration: number
  rating: number
  language: string
  director: string
  cast: string[]
  posterUrl: string
  releaseDate: string
  isActive: boolean
}

export const useMovieStore = defineStore('movie', () => {
  const movies = ref<Movie[]>([])
  const currentMovie = ref<Movie | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchAll() {
    loading.value = true
    error.value = null
    try {
      const res = await movieApi.getAll()
      movies.value = res.data.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to load movies'
    } finally {
      loading.value = false
    }
  }

  async function fetchById(id: string) {
    loading.value = true
    error.value = null
    try {
      const res = await movieApi.getById(id)
      currentMovie.value = res.data.data
      return res.data.data as Movie
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Movie not found'
      return null
    } finally {
      loading.value = false
    }
  }

  return { movies, currentMovie, loading, error, fetchAll, fetchById }
})
