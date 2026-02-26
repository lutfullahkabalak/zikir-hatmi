<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

type JoinResponse = {
  token: string
}

const route = useRoute()
const router = useRouter()

const shareCode = computed(() => String(route.params.shareCode || ''))
const password = ref('')
const loading = ref(false)
const errorMessage = ref<string | null>(null)

const tokenKey = computed(() => `hatim-token:${shareCode.value}`)

const submit = async () => {
  if (!shareCode.value) {
    return
  }

  loading.value = true
  errorMessage.value = null

  try {
    const response = await fetch(`/hatims/${shareCode.value}/join`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: password.value }),
    })

    if (!response.ok) {
      if (response.status === 401) {
        errorMessage.value = 'Şifre hatalı. Lütfen tekrar deneyin.'
        return
      }
      if (response.status === 404) {
        errorMessage.value = 'Hatim bulunamadı.'
        return
      }
      errorMessage.value = 'Katılım sırasında hata oluştu.'
      return
    }

    const data = (await response.json()) as JoinResponse
    localStorage.setItem(tokenKey.value, data.token)
    await router.replace({ name: 'hatim', params: { shareCode: shareCode.value } })
  } catch (error) {
    errorMessage.value = 'Katılım sırasında hata oluştu.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-[70vh] items-center justify-center px-6 py-12">
    <UCard class="w-full max-w-md bg-white/5">
      <div class="space-y-4">
        <div>
          <p class="text-xs uppercase tracking-[0.35em] text-slate-700/80 dark:text-white/50">Zikir Hatmi</p>
          <h1 class="mt-3 text-2xl font-semibold">Şifre ile katıl</h1>
          <p class="mt-2 text-sm text-slate-700 dark:text-white/70">
            Bu hatime katılmak için şifre girmeniz gerekiyor.
          </p>
        </div>

        <UInput
          v-model="password"
          type="password"
          placeholder="Hatim şifresi"
          size="lg"
        />

        <UButton
          color="primary"
          size="lg"
          class="w-full"
          :loading="loading"
          @click="submit"
        >
          Katıl
        </UButton>

        <p v-if="errorMessage" class="text-sm text-rose-300">
          {{ errorMessage }}
        </p>
      </div>
    </UCard>
  </div>
</template>
