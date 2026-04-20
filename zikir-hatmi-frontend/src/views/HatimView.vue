<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUsername } from '../composables/useUsername'
import { useZikirSocket } from '../composables/useZikirSocket'
import { apiUrl } from '../api'

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

const { username } = useUsername()

const { count, target, connected, activeUsers, increment, setState, disconnect } = useZikirSocket({
  hatimId: shareCode,
  token,
  username,
})

const bulkIncrementOpen = ref(false)
const bulkIncrementAmount = ref<number | null>(null)
const presenceOpen = ref(false)
const maxVisibleUsers = 4
const visibleUsers = computed(() => activeUsers.value.slice(0, maxVisibleUsers))
const hasMoreUsers = computed(() => activeUsers.value.length > maxVisibleUsers)

const toInitials = (name: string) => {
  const cleaned = (name || '').trim()
  if (!cleaned) return '?'
  const parts = cleaned.split(/\s+/).filter(Boolean)
  const first = parts[0]?.[0] || ''
  const second = parts.length > 1 ? (parts[parts.length - 1]?.[0] || '') : ''
  const initials = (first + second).toUpperCase()
  return initials || cleaned.charAt(0).toUpperCase() || '?'
}

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
    const response = await fetch(apiUrl(`/hatims/${shareCode.value}`))
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
    const joinResponse = await fetch(apiUrl(`/hatims/${shareCode.value}/join`), {
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

const openBulkIncrementModal = () => {
  bulkIncrementAmount.value = null
  bulkIncrementOpen.value = true
}

const applyBulkIncrement = () => {
  const amount = Number.isFinite(bulkIncrementAmount.value) ? Math.floor(Number(bulkIncrementAmount.value)) : 0
  if (amount <= 0) {
    return
  }
  increment(amount)
  bulkIncrementOpen.value = false
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
    <UButton
      icon="i-lucide-plus"
      variant="solid"
      class="fixed left-4 top-4 z-30 bg-white text-slate-900 hover:bg-white/90"
      aria-label="Toplu zikir ekle"
      :disabled="loading || joinLoading || !connected || isCompleted"
      @click="openBulkIncrementModal"
    />

    <div class="relative z-10 flex w-full max-w-2xl flex-col items-center gap-6">

      <UCard v-if="showCreatedBanner" class="w-full bg-transparent">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p class="text-sm font-semibold text-slate-900 dark:text-white">Paylaşım bağlantısı hazır</p>
            <p class="text-xs text-slate-700 dark:text-white/60">{{ createdLink }}</p>
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
          <p class="text-sm text-slate-700 dark:text-white/70">
            Hedef: <span class="font-semibold text-slate-900 dark:text-white">{{ target }}</span>
          </p>
          <p class="text-xs uppercase tracking-[0.3em] text-slate-700/70 dark:text-white/40">
            {{ connected ? 'Canlı bağlantı' : 'Bağlanıyor...' }}
          </p>
          <p v-if="isCompleted" class="text-sm font-semibold text-emerald-300">
            Hatim tamamlandı.
          </p>

          <button
            v-if="connected && activeUsers.length"
            type="button"
            class="mt-3 flex items-center justify-center rounded-full bg-white/5 px-3 py-2 ring-1 ring-white/10 hover:bg-white/10"
            aria-label="Aktif kullanıcıları göster"
            @click="presenceOpen = true"
          >
            <div class="flex -space-x-2">
              <UAvatar
                v-for="u in visibleUsers"
                :key="u.id"
                :text="toInitials(u.name)"
                size="xs"
                class="ring-2 ring-black"
                :title="u.name"
              />
              <UAvatar
                v-if="hasMoreUsers"
                text="…"
                size="xs"
                class="ring-2 ring-black"
                title="Tümünü gör"
              />
            </div>
          </button>
        </div>
      </div>
    </div>

    <UModal
      v-model:open="bulkIncrementOpen"
      :overlay="true"
      :ui="{
        overlay: 'z-[100] bg-black/90 backdrop-blur-md',
        content: 'z-[110] !bg-black text-white ring-1 ring-white/10 rounded-xl shadow-2xl',
      }"
    >
      <template #content>
        <div class="rounded-xl bg-black p-6 text-white shadow-2xl ring-1 ring-white/10">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-xs uppercase tracking-[0.35em] text-white/50">Toplu ekleme</p>
              <h2 class="mt-3 text-xl font-semibold">Zikire kaç tane eklenecek?</h2>
            </div>
            <UButton
              variant="ghost"
              color="neutral"
              size="sm"
              icon="i-lucide-x"
              aria-label="Kapat"
              @click="bulkIncrementOpen = false"
            />
          </div>

          <UFormField class="mt-5" label="Eklenecek sayı">
            <UInput
              v-model.number="bulkIncrementAmount"
              type="number"
              min="1"
              step="1"
              placeholder="Örn: 10"
            />
          </UFormField>

          <div class="mt-5 flex justify-end">
            <UButton
              color="primary"
              :disabled="!bulkIncrementAmount || bulkIncrementAmount < 1 || isCompleted"
              @click="applyBulkIncrement"
            >
              Ekle
            </UButton>
          </div>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="presenceOpen"
      :overlay="true"
      :ui="{
        overlay: 'z-[100] bg-black/90 backdrop-blur-md',
        content: 'z-[110] !bg-black text-white ring-1 ring-white/10 rounded-xl shadow-2xl',
      }"
    >
      <template #content>
        <div class="rounded-xl bg-black p-6 text-white shadow-2xl ring-1 ring-white/10">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-xs uppercase tracking-[0.35em] text-white/50">Aktif kullanıcılar</p>
              <h2 class="mt-3 text-xl font-semibold">Şu an bağlı: {{ activeUsers.length }}</h2>
              <p class="mt-2 text-sm text-white/70">Kullanıcı adını menüden ayarlayabilirsin.</p>
            </div>
            <UButton
              variant="ghost"
              color="neutral"
              size="sm"
              icon="i-lucide-x"
              aria-label="Kapat"
              @click="presenceOpen = false"
            />
          </div>

          <div class="mt-5 space-y-2">
            <div
              v-for="u in activeUsers"
              :key="u.id"
              class="flex items-center gap-3 rounded-lg bg-white/5 px-3 py-2 ring-1 ring-white/10"
            >
              <UAvatar :text="toInitials(u.name)" size="sm" />
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-white">{{ u.name }}</p>
              </div>
            </div>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>
