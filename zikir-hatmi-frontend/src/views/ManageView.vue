<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

type HatimSummary = {
  shareCode: string
  title: string
  count: number
  target: number
  requiresPassword: boolean
  createdAt: string
  updatedAt: string
}

const items = ref<HatimSummary[]>([])
const loading = ref(false)
const errorMessage = ref<string | null>(null)
const savingCode = ref<string | null>(null)
const deletingCode = ref<string | null>(null)
const adminKey = ref('')
const authenticated = ref(false)

const editState = ref<Record<string, { title: string; count: number; target: number }>>({})

const hasItems = computed(() => items.value.length > 0)

const getAuthHeaders = (): HeadersInit => {
  const headers: HeadersInit = { 'Content-Type': 'application/json' }
  if (adminKey.value) {
    headers['Authorization'] = `Bearer ${adminKey.value}`
  }
  return headers
}

const ensureDraft = (shareCode: string) => {
  if (!editState.value[shareCode]) {
    editState.value[shareCode] = {
      title: '',
      count: 0,
      target: 1,
    }
  }
  return editState.value[shareCode]
}

const setDraftTitle = (shareCode: string, value: string | number | undefined) => {
  const draft = ensureDraft(shareCode)
  draft.title = String(value ?? '')
}

const setDraftCount = (shareCode: string, value: string | number | undefined) => {
  const draft = ensureDraft(shareCode)
  draft.count = Math.max(0, Math.floor(Number(value) || 0))
}

const setDraftTarget = (shareCode: string, value: string | number | undefined) => {
  const draft = ensureDraft(shareCode)
  draft.target = Math.max(1, Math.floor(Number(value) || 1))
}

const formatDate = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '-'
  }
  return date.toLocaleString('tr-TR')
}

const shareLink = (shareCode: string) => `${window.location.origin}/h/${shareCode}`

const hydrateEditState = (source: HatimSummary[]) => {
  const next: Record<string, { title: string; count: number; target: number }> = {}
  for (const item of source) {
    next[item.shareCode] = {
      title: item.title || '',
      count: item.count,
      target: item.target,
    }
  }
  editState.value = next
}

const loadHatims = async () => {
  if (!adminKey.value) {
    errorMessage.value = 'Admin anahtarı gerekli.'
    return
  }

  loading.value = true
  errorMessage.value = null

  try {
    const response = await fetch('/hatims', {
      headers: getAuthHeaders(),
    })
    if (!response.ok) {
      if (response.status === 401 || response.status === 403) {
        errorMessage.value = 'Geçersiz admin anahtarı.'
        authenticated.value = false
        return
      }
      errorMessage.value = 'Hatimler alınamadı.'
      return
    }
    const data = (await response.json()) as HatimSummary[]
    hydrateEditState(data)
    items.value = data
    authenticated.value = true
  } catch {
    errorMessage.value = 'Hatimler alınamadı.'
  } finally {
    loading.value = false
  }
}

const saveHatim = async (shareCode: string) => {
  const draft = editState.value[shareCode]
  if (!draft) {
    return
  }

  savingCode.value = shareCode
  errorMessage.value = null

  try {
    const response = await fetch(`/hatims/${shareCode}`, {
      method: 'PATCH',
      headers: getAuthHeaders(),
      body: JSON.stringify({
        title: draft.title.trim(),
        count: Math.max(0, Math.floor(Number(draft.count) || 0)),
        target: Math.max(1, Math.floor(Number(draft.target) || 1)),
      }),
    })

    if (!response.ok) {
      if (response.status === 401) {
        errorMessage.value = 'Yetkilendirme hatası. Token gerekli.'
        return
      }
      errorMessage.value = 'Hatim güncellenemedi.'
      return
    }

    await loadHatims()
  } catch {
    errorMessage.value = 'Hatim güncellenemedi.'
  } finally {
    savingCode.value = null
  }
}

const resetCount = async (shareCode: string) => {
  const draft = editState.value[shareCode]
  if (!draft) {
    return
  }
  draft.count = 0
  await saveHatim(shareCode)
}

const deleteHatim = async (shareCode: string) => {
  const confirmed = window.confirm('Bu hatim silinsin mi?')
  if (!confirmed) {
    return
  }

  deletingCode.value = shareCode
  errorMessage.value = null

  try {
    const response = await fetch(`/hatims/${shareCode}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    })

    if (!response.ok) {
      if (response.status === 401) {
        errorMessage.value = 'Yetkilendirme hatası. Token gerekli.'
        return
      }
      errorMessage.value = 'Hatim silinemedi.'
      return
    }

    items.value = items.value.filter((item) => item.shareCode !== shareCode)
    const copy = { ...editState.value }
    delete copy[shareCode]
    editState.value = copy
  } catch {
    errorMessage.value = 'Hatim silinemedi.'
  } finally {
    deletingCode.value = null
  }
}

const copyLink = async (shareCode: string) => {
  try {
    await navigator.clipboard.writeText(shareLink(shareCode))
  } catch {
    // ignore
  }
}

onMounted(() => {
  // Don't auto-load, require authentication first
})
</script>

<template>
  <div class="mx-auto w-full max-w-5xl space-y-4">
    <div class="flex items-center justify-between gap-3">
      <div>
        <p class="text-xs uppercase tracking-[0.35em] text-slate-700/80 dark:text-white/50">Yönetim</p>
        <h1 class="mt-2 text-2xl font-semibold">Hatim paneli</h1>
      </div>
      <UButton v-if="authenticated" color="primary" :loading="loading" @click="loadHatims">
        Yenile
      </UButton>
    </div>

    <!-- Authentication form -->
    <UCard v-if="!authenticated" class="bg-white/5">
      <div class="space-y-4">
        <p class="text-sm text-slate-700 dark:text-white/70">
          Bu panele erişmek için admin anahtarı gereklidir.
        </p>
        <UInput
          v-model="adminKey"
          type="password"
          placeholder="Admin anahtarı"
          size="lg"
        />
        <UButton
          color="primary"
          :loading="loading"
          @click="loadHatims"
        >
          Giriş
        </UButton>
      </div>
    </UCard>

    <UAlert
      v-if="errorMessage"
      color="error"
      variant="soft"
      :title="errorMessage"
    />

    <UCard v-if="authenticated && !loading && !hasItems" class="bg-white/5">
      <p class="text-sm text-slate-700 dark:text-white/70">Kayıtlı hatim yok.</p>
    </UCard>

    <div v-if="authenticated" class="space-y-3">
      <UCard
        v-for="item in items"
        :key="item.shareCode"
        class="bg-white/5"
      >
        <div class="space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <p class="text-sm font-semibold text-slate-900 dark:text-white">
              {{ item.shareCode }}
            </p>
            <div class="flex items-center gap-2 text-xs text-slate-700 dark:text-white/60">
              <span>{{ formatDate(item.updatedAt) }}</span>
              <UBadge color="neutral" variant="soft">
                {{ item.requiresPassword ? 'Şifreli' : 'Açık' }}
              </UBadge>
            </div>
          </div>

          <div class="grid gap-2 md:grid-cols-3">
            <UInput
              :model-value="editState[item.shareCode]?.title ?? ''"
              placeholder="Başlık"
              @update:model-value="(value: string | number | undefined) => setDraftTitle(item.shareCode, value)"
            />
            <UInput
              :model-value="editState[item.shareCode]?.count ?? 0"
              type="number"
              min="0"
              placeholder="Sayı"
              @update:model-value="(value: string | number | undefined) => setDraftCount(item.shareCode, value)"
            />
            <UInput
              :model-value="editState[item.shareCode]?.target ?? 1"
              type="number"
              min="1"
              placeholder="Hedef"
              @update:model-value="(value: string | number | undefined) => setDraftTarget(item.shareCode, value)"
            />
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <UButton
              color="primary"
              :loading="savingCode === item.shareCode"
              @click="saveHatim(item.shareCode)"
            >
              Kaydet
            </UButton>
            <UButton
              variant="soft"
              color="neutral"
              :disabled="savingCode === item.shareCode"
              @click="resetCount(item.shareCode)"
            >
              Sayacı sıfırla
            </UButton>
            <UButton variant="soft" color="neutral" @click="copyLink(item.shareCode)">
              Link kopyala
            </UButton>
            <UButton
              variant="soft"
              color="error"
              :loading="deletingCode === item.shareCode"
              @click="deleteHatim(item.shareCode)"
            >
              Sil
            </UButton>
          </div>

          <p class="text-xs text-slate-700/80 dark:text-white/50">
            Oluşturulma: {{ formatDate(item.createdAt) }} · Link: {{ shareLink(item.shareCode) }}
          </p>
        </div>
      </UCard>
    </div>
  </div>
</template>
