<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { apiUrl } from '../api'

type ParticipantEvent = {
  kind: string
  name?: string
  from?: string
  to?: string
  at: string
}

type Contributor = {
  publicId: string
  events: ParticipantEvent[]
}

const route = useRoute()
const shareCode = computed(() => String(route.params.shareCode || ''))

const loading = ref(false)
const errorMessage = ref<string | null>(null)
const contributors = ref<Contributor[]>([])

const hasItems = computed(() => contributors.value.length > 0)

const formatDate = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '-'
  }
  return date.toLocaleString('tr-TR')
}

const eventLabel = (e: ParticipantEvent) => {
  if (e.kind === 'entry') {
    return `Giriş: ${e.name ?? '—'}`
  }
  if (e.kind === 'rename') {
    return `İsim: ${e.from ?? '—'} → ${e.to ?? '—'}`
  }
  return e.kind
}

const load = async () => {
  if (!shareCode.value) {
    contributors.value = []
    return
  }

  loading.value = true
  errorMessage.value = null

  try {
    const response = await fetch(apiUrl(`/hatims/${encodeURIComponent(shareCode.value)}/contributors`))
    if (!response.ok) {
      if (response.status === 404) {
        errorMessage.value = 'Hatim bulunamadı.'
        contributors.value = []
        return
      }
      errorMessage.value = 'Liste alınamadı.'
      contributors.value = []
      return
    }

    const data = (await response.json()) as { contributors?: Contributor[] }
    contributors.value = Array.isArray(data.contributors) ? data.contributors : []
  } catch {
    errorMessage.value = 'Liste alınamadı.'
    contributors.value = []
  } finally {
    loading.value = false
  }
}

watch(shareCode, () => {
  load()
}, { immediate: true })

onMounted(() => {
  load()
})
</script>

<template>
  <div class="mx-auto w-full max-w-3xl space-y-4">
    <div class="flex items-center justify-between gap-3">
      <div>
        <p class="text-xs uppercase tracking-[0.35em] text-slate-700/80 dark:text-white/50">Liste</p>
        <h1 class="mt-2 text-2xl font-semibold">Katılımcı günlüğü</h1>
      </div>
      <UButton color="primary" :loading="loading" @click="load">
        Yenile
      </UButton>
    </div>

    <UAlert
      v-if="errorMessage"
      color="error"
      variant="soft"
      :title="errorMessage"
    />

    <UCard v-if="!loading && !hasItems && !errorMessage" class="bg-white/5">
      <p class="text-sm text-slate-700 dark:text-white/70">Henüz giriş veya isim değişikliği kaydı yok.</p>
    </UCard>

    <div v-if="hasItems" class="space-y-3">
      <UCard v-for="c in contributors" :key="c.publicId" class="bg-white/5">
        <div class="space-y-2">
          <p class="font-mono text-xs text-slate-500 dark:text-white/45">{{ c.publicId }}</p>
          <ul class="list-inside list-disc space-y-1 text-sm text-slate-800 dark:text-white/90">
            <li v-for="(ev, idx) in c.events" :key="idx">
              <span class="text-xs text-slate-500 dark:text-white/45">{{ formatDate(ev.at) }}</span>
              · <span class="text-sm">{{ eventLabel(ev) }}</span>
            </li>
          </ul>
        </div>
      </UCard>
    </div>
  </div>
</template>

