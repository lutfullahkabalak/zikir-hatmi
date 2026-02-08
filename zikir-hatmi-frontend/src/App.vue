<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

type CreateHatimResponse = {
  shareCode: string
  token: string
  target: number
}

const route = useRoute()
const router = useRouter()

const createOpen = ref(false)
const title = ref('')
const target = ref(50)
const password = ref('')
const creating = ref(false)
const createError = ref<string | null>(null)

const shareCode = computed(() => String(route.params.shareCode || ''))
const shareLink = computed(() =>
  shareCode.value ? `${window.location.origin}/h/${shareCode.value}` : ''
)

const canShare = computed(() => shareCode.value.length > 0)

const headerTitle = ref('')

watch(shareCode, async (code) => {
  headerTitle.value = ''
  if (!code) return
  try {
    const res = await fetch(`/hatims/${code}`)
    if (res.ok) {
      const data = await res.json()
      headerTitle.value = data.title || ''
    }
  } catch { /* ignore */ }
}, { immediate: true })

const copyShareLink = async () => {
  if (!shareLink.value) {
    return
  }
  try {
    await navigator.clipboard.writeText(shareLink.value)
  } catch (error) {
    // ignore
  }
}

const resetForm = () => {
  title.value = ''
  target.value = 50
  password.value = ''
  createError.value = null
}

const submitCreate = async () => {
  creating.value = true
  createError.value = null

  try {
    const response = await fetch('/hatims', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title: title.value.trim(),
        target: target.value,
        password: password.value.trim(),
      }),
    })

    if (!response.ok) {
      createError.value = 'Hatim oluşturulamadı.'
      return
    }

    const data = (await response.json()) as CreateHatimResponse
    localStorage.setItem(`hatim-token:${data.shareCode}`, data.token)
    createOpen.value = false
    resetForm()
    await router.push({
      name: 'hatim',
      params: { shareCode: data.shareCode },
      query: { created: '1' },
    })
  } catch (error) {
    createError.value = 'Hatim oluşturulamadı.'
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <UApp class="min-h-screen bg-transparent text-white">
    <main class="relative flex min-h-screen flex-col">
      <header class="sticky top-0 z-20 grid grid-cols-[auto_1fr_auto] items-center gap-2 px-4 py-4">
        <div class="flex items-center">
          <UButton
            v-if="canShare"
            variant="ghost"
            color="neutral"
            size="sm"
            icon="i-lucide-share-2"
            @click="copyShareLink"
          />
          <div v-else class="w-16"></div>
        </div>

        <p
          v-if="headerTitle"
          class="truncate text-center text-lg font-semibold text-white/90"
          :title="headerTitle"
        >
          {{ headerTitle }}
        </p>
        <div v-else></div>

        <div class="flex items-center justify-end">
          <UButton
            variant="ghost"
            color="neutral"
            size="sm"
            icon="i-lucide-plus"
            @click="createOpen = true"
          />
        </div>
      </header>

      <div class="flex flex-1 px-6 py-10">
        <div class="flex-1 bg-transparent">
          <RouterView />
        </div>
      </div>
    </main>

    <UModal
      v-model:open="createOpen"
      :overlay="true"
      :ui="{
        overlay: 'z-[100] bg-black/90 backdrop-blur-md',
        content: 'z-[110] !bg-black text-white ring-1 ring-white/10 rounded-xl shadow-2xl',
      }"
    >
      <template #content>
        <div class="rounded-xl bg-black p-6 text-white shadow-2xl ring-1 ring-white/10">
          <div class="space-y-4">
            <div>
              <p class="text-xs uppercase tracking-[0.35em] text-white/50">Yeni Hatim</p>
              <h2 class="mt-3 text-xl font-semibold">Hatim oluştur</h2>
              <p class="mt-2 text-sm text-white/70">
                Hedef belirleyin ve dilerseniz şifre ekleyin.
              </p>
            </div>

            <UInput
              v-model="title"
              type="text"
              placeholder="Hatim başlığı (ör. Yasin-i Şerif)"
              size="lg"
            />

            <UInput
              v-model.number="target"
              type="number"
              min="1"
              placeholder="Hedef (ör. 50)"
              size="lg"
            />

            <UInput
              v-model="password"
              type="password"
              placeholder="Şifre (opsiyonel)"
              size="lg"
            />

            <UButton
              color="primary"
              size="lg"
              class="w-full"
              :loading="creating"
              @click="submitCreate"
            >
              Oluştur
            </UButton>

            <p v-if="createError" class="text-sm text-rose-300">
              {{ createError }}
            </p>
          </div>
        </div>
      </template>
    </UModal>
  </UApp>
</template>
