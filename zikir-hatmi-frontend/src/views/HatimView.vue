<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useZikirSocket } from '../composables/useZikirSocket'

type HatimStateResponse = {
  shareCode: string
  title: string
  count: number
  target: number
  requiresPassword: boolean
}

type JoinResponse = {
  token: string
}

const route = useRoute()
const router = useRouter()

const shareCode = computed(() => String(route.params.shareCode || ''))
const token = ref<string | null>(null)
const hatimTitle = ref('')
const loading = ref(true)
const errorMessage = ref<string | null>(null)
const joinLoading = ref(false)

const tokenKey = computed(() => `hatim-token:${shareCode.value}`)

const { count, target, connected, increment, setState, disconnect } = useZikirSocket({
  hatimId: shareCode,
  token,
})

const radius = 140
const circumference = 2 * Math.PI * radius

const progressRatio = computed(() => {
  if (target.value <= 0) {
    return 0
  }
  return Math.min(count.value / target.value, 1)
})

const dashOffset = computed(() => circumference * (1 - progressRatio.value))

const progressColor = computed(() => {
  const ratio = progressRatio.value
  const hue = 190 + (10 - 190) * ratio
  const saturation = 80
  const lightness = 60 - 8 * ratio
  return `hsl(${hue} ${saturation}% ${lightness}%)`
})

const isCompleted = computed(() => count.value >= target.value && target.value > 0)

const createdLink = computed(() => `${window.location.origin}/h/${shareCode.value}`)
const showCreatedBanner = computed(() => route.query.created === '1')

const loadHatim = async () => {
  if (!shareCode.value) {
    return
  }

  loading.value = true
  errorMessage.value = null
  disconnect()

  try {
    const response = await fetch(`/hatims/${shareCode.value}`)
    if (!response.ok) {
      if (response.status === 404) {
        errorMessage.value = 'Hatim bulunamadı.'
        return
      }
      errorMessage.value = 'Hatim bilgileri alınamadı.'
      return
    }

    const data = (await response.json()) as HatimStateResponse
    hatimTitle.value = data.title
    setState(data.count, data.target)
    const storedToken = localStorage.getItem(tokenKey.value)
    if (data.requiresPassword) {
      if (!storedToken) {
        await router.replace({ name: 'hatim-join', params: { shareCode: shareCode.value } })
        loading.value = false
        return
      }
      token.value = storedToken
      return
    }

    if (storedToken) {
      token.value = storedToken
      return
    }

    joinLoading.value = true
    const joinResponse = await fetch(`/hatims/${shareCode.value}/join`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    })

    if (!joinResponse.ok) {
      errorMessage.value = 'Hatime bağlanılamadı.'
      return
    }

    const joinData = (await joinResponse.json()) as JoinResponse
    token.value = joinData.token
    localStorage.setItem(tokenKey.value, joinData.token)
  } catch (error) {
    errorMessage.value = 'Hatim bilgileri alınamadı.'
  } finally {
    joinLoading.value = false
    loading.value = false
  }
}

const copyCreatedLink = async () => {
  try {
    await navigator.clipboard.writeText(createdLink.value)
  } catch (error) {
    // ignore
  }
}

const dismissCreatedBanner = async () => {
  try {
    await router.replace({
      query: {
        ...route.query,
        created: undefined,
      } as any,
    })
  } catch {
    // ignore
  }
}

watch(shareCode, () => {
  loadHatim()
}, { immediate: true })

onBeforeUnmount(() => {
  disconnect()
})
</script>

<template>
  <div class="relative flex w-full flex-1 items-center justify-center">

    <div class="relative z-10 flex w-full max-w-2xl flex-col items-center gap-6">

      <UCard v-if="showCreatedBanner" class="w-full bg-transparent">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p class="text-sm font-semibold text-white">Paylaşım bağlantısı hazır</p>
            <p class="text-xs text-white/60">{{ createdLink }}</p>
          </div>
          <div class="flex items-center gap-2">
            <UButton color="primary" variant="soft" @click="copyCreatedLink">
              Linki kopyala
            </UButton>
            <UButton
              variant="ghost"
              color="neutral"
              size="sm"
              icon="i-lucide-x"
              aria-label="Kapat"
              @click="dismissCreatedBanner"
            />
          </div>
        </div>
      </UCard>

      <UCard v-if="errorMessage" class="w-full bg-transparent">
        <p class="text-sm text-rose-200">{{ errorMessage }}</p>
      </UCard>

      <div v-else class="w-full">
        <div class="flex flex-col items-center gap-4">
          <div class="text-center">
            <p class="text-5xl font-semibold tracking-tight md:text-6xl">{{ count }}</p>
            <p class="text-sm uppercase tracking-[0.35em] text-white/50">zikir</p>
          </div>

          <div class="relative flex items-center justify-center">
            <svg
              class="h-80 w-80 -rotate-90"
              viewBox="0 0 320 320"
              role="img"
              aria-label="Zikir ilerlemesi"
            >
              <circle
                cx="160"
                cy="160"
                :r="radius"
                stroke-width="18"
                class="stroke-white/10"
                fill="none"
              />
              <circle
                cx="160"
                cy="160"
                :r="radius"
                stroke-width="18"
                class="transition-all duration-500"
                fill="none"
                stroke-linecap="round"
                :stroke-dasharray="circumference"
                :stroke-dashoffset="dashOffset"
                :stroke="progressColor"
              />
            </svg>

            <UButton
              size="lg"
              color="primary"
              class="absolute h-64 w-64 rounded-full text-lg shadow-lg shadow-cyan-500/30"
              :disabled="loading || joinLoading || !connected || isCompleted"
              @click="increment"
            />
          </div>
        </div>

        <div class="mt-6 flex flex-col items-center gap-2 text-center">
          <p class="text-sm text-white/70">
            Hedef: <span class="font-semibold text-white">{{ target }}</span>
          </p>
          <p class="text-xs uppercase tracking-[0.3em] text-white/40">
            {{ connected ? 'Canlı bağlantı' : 'Bağlanıyor...' }}
          </p>
          <p v-if="isCompleted" class="text-sm font-semibold text-emerald-300">
            Hatim tamamlandı.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
