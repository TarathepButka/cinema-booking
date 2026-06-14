import { onUnmounted, ref, watch, type Ref } from 'vue'

type UseAutocompleteSearchOptions = {
  query: Ref<string>
  fetchSuggestions: (query: string) => Promise<string[]>
  onSearch: (query: string) => void
  searchDelay?: number
  suggestionDelay?: number
  shouldSkip?: () => boolean
  onSuggestionError?: (error: unknown) => void
}

export function useAutocompleteSearch({
  query,
  fetchSuggestions,
  onSearch,
  searchDelay = 3000,
  suggestionDelay = 300,
  shouldSkip,
  onSuggestionError,
}: UseAutocompleteSearchOptions) {
  const suggestions = ref<string[]>([])
  const showSuggestions = ref(false)
  const skipNextSearch = ref(false)

  let suggestionTimeout: ReturnType<typeof setTimeout> | null = null
  let searchTimeout: ReturnType<typeof setTimeout> | null = null

  function clearTimers() {
    if (suggestionTimeout) {
      clearTimeout(suggestionTimeout)
      suggestionTimeout = null
    }

    if (searchTimeout) {
      clearTimeout(searchTimeout)
      searchTimeout = null
    }
  }

  function clearSuggestions() {
    suggestions.value = []
    showSuggestions.value = false
  }

  function triggerSearchInstant() {
    if (searchTimeout) {
      clearTimeout(searchTimeout)
      searchTimeout = null
    }

    onSearch(query.value)
    showSuggestions.value = false
  }

  function selectSuggestion(item: string) {
    skipNextSearch.value = true
    query.value = item
    showSuggestions.value = false
    onSearch(item)
  }

  watch(query, (newValue) => {
    if (shouldSkip?.()) {
      return
    }

    if (suggestionTimeout) {
      clearTimeout(suggestionTimeout)
      suggestionTimeout = null
    }

    if (!newValue) {
      suggestions.value = []
    } else {
      suggestionTimeout = setTimeout(async () => {
        try {
          suggestions.value = await fetchSuggestions(newValue)
        } catch (error) {
          onSuggestionError?.(error)
        }
      }, suggestionDelay)
    }

    if (searchTimeout) {
      clearTimeout(searchTimeout)
      searchTimeout = null
    }

    if (skipNextSearch.value) {
      skipNextSearch.value = false
      return
    }

    if (!newValue) {
      onSearch('')
      return
    }

    searchTimeout = setTimeout(() => {
      onSearch(newValue)
    }, searchDelay)
  })

  onUnmounted(() => {
    clearTimers()
  })

  return {
    suggestions,
    showSuggestions,
    selectSuggestion,
    triggerSearchInstant,
    clearSuggestions,
    clearTimers,
  }
}
