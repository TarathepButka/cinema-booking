<script setup lang="ts">
import { useRouter } from 'vue-router'

const props = defineProps<{
  currentStep: 1 | 2 | 3
  backUrl?: string
}>()

const router = useRouter()

function goBack() {
  if (props.backUrl) {
    router.push(props.backUrl)
  } else {
    router.back()
  }
}
</script>

<template>
  <div class="booking-steps-bar">
    <div class="steps-container">
      <button class="btn-step-back" @click="goBack" aria-label="Go back">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="19" y1="12" x2="5" y2="12"></line><polyline points="12 19 5 12 12 5"></polyline></svg>
        <span>Back</span>
      </button>
      
      <div class="steps-list">
        <div class="step-item" :class="{ active: currentStep === 1, completed: currentStep > 1 }">
          <span class="step-num">
            <span v-if="currentStep > 1">✓</span>
            <span v-else>1</span>
          </span>
          <span class="step-label">Select Showtime</span>
        </div>
        <div class="step-line" :class="{ 'step-line-active': currentStep === 2, 'step-line-completed': currentStep > 2 }"></div>
        <div class="step-item" :class="{ active: currentStep === 2, completed: currentStep > 2 }">
          <span class="step-num">
            <span v-if="currentStep > 2">✓</span>
            <span v-else>2</span>
          </span>
          <span class="step-label">Select Seats</span>
        </div>
        <div class="step-line" :class="{ 'step-line-active': currentStep === 3, 'step-line-completed': currentStep > 3 }"></div>
        <div class="step-item" :class="{ active: currentStep === 3 }">
          <span class="step-num">3</span>
          <span class="step-label">Payment</span>
        </div>
      </div>
    </div>
  </div>
</template>
