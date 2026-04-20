<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUsername } from './composables/useUsername'
import { apiUrl } from './api'

type CreateHatimResponse = {
  shareCode: string
  token: string
  target: number
}

const route = useRoute()
const router = useRouter()

const createOpen = ref(false)
const usernameOpen = ref(false)
const title = ref('')
const target = ref(50)
const password = ref('')
const creating = ref(false)
const createError = ref<string | null>(null)

const { username, setUsername } = useUsername()
const usernameDraft = ref('')

watch(usernameOpen, (open) => {
  if (open) {
    usernameDraft.value = username.value || ''
  }
})

const shareCode = computed(() => String(route.params.shareCode || ''))
const shareLink = computed(() =>
  shareCode.value ? `${window.location.origin}/h/${shareCode.value}` : ''
)

const canShare = computed(() => shareCode.value.length > 0)

const menuItems = computed(() => [
  [
    {
      label: 'Paylaş',
      icon: 'i-lucide-share-2',
      disabled: !canShare.value,
      onSelect: copyShareLink,
    },
    {
      label: 'Kullanıcı adı',
      icon: 'i-lucide-user',
      onSelect: () => {
        usernameOpen.value = true
      },
    },
    {
      label: 'Yeni hatim',
      icon: 'i-lucide-plus',
      onSelect: () => {
        createOpen.value = true
      },
    },
  ],
])

const headerTitle = ref('')

watch(shareCode, async (code) => {
  headerTitle.value = ''
  if (!code) return
  try {
    const res = await fetch(apiUrl(`/hatims/${code}`))
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
    const response = await fetch(apiUrl('/hatims'), {
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

const submitUsername = () => {
  setUsername(usernameDraft.value)
  usernameOpen.value = false
}

const clearUsername = () => {
  setUsername('')
  usernameDraft.value = ''
  usernameOpen.value = false
}
</script>

<template>
  <UApp class="ramadan-app min-h-dvh bg-transparent">
    <main class="relative flex min-h-dvh flex-col">
      <header class="sticky top-0 z-20 grid grid-cols-[auto_1fr_auto] items-center gap-2 px-4 py-4">
        <div class="w-9"></div>

        <p
          v-if="headerTitle"
          class="truncate text-center text-lg font-semibold text-slate-900 dark:text-white/90"
          :title="headerTitle"
        >
          {{ headerTitle }}
        </p>
        <div v-else></div>

        <div class="flex items-center justify-end">
          <UDropdownMenu
            :items="menuItems"
            :content="{ align: 'end' }"
          >
            <UButton
              variant="solid"
              size="md"
              icon="i-lucide-ellipsis-vertical"
              aria-label="Menü"
              class="bg-white text-slate-900 hover:bg-white/90"
            />
          </UDropdownMenu>
        </div>
      </header>

      <div class="flex flex-1 flex-col px-4 py-4 sm:px-6 sm:py-10">
        <div class="flex flex-1 flex-col bg-transparent">
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

    <UModal
      v-model:open="usernameOpen"
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
              <p class="text-xs uppercase tracking-[0.35em] text-white/50">Kullanıcı</p>
              <h2 class="mt-3 text-xl font-semibold">Kullanıcı adın</h2>
              <p class="mt-2 text-sm text-white/70">
                Opsiyonel. Bu cihazda saklanır ve hatimde aktif kullanıcılar arasında görünür.
              </p>
            </div>

            <UInput
              v-model="usernameDraft"
              type="text"
              placeholder="Adınız (ör. Ahmet)"
              size="lg"
            />

            <div class="flex items-center gap-2">
              <UButton color="primary" size="lg" class="flex-1" @click="submitUsername">
                Kaydet
              </UButton>
              <UButton color="neutral" variant="soft" size="lg" @click="clearUsername">
                Temizle
              </UButton>
            </div>
          </div>
        </div>
      </template>
    </UModal>
  </UApp>
</template>
