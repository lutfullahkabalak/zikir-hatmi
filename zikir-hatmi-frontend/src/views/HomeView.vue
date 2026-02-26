<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const hasQueryMessage = computed(() => route.query.created === '1')
</script>

<template>
  <div class="flex min-h-[70vh] flex-col items-center justify-center gap-6 text-center">
    <div>
      <p class="text-xs uppercase tracking-[0.4em] text-slate-700/80 dark:text-white/50">Zikir Hatmi</p>
      <h1 class="mt-4 text-3xl font-semibold md:text-4xl">Dinamik hatim paylaşımı</h1>
      <p class="mt-3 text-base text-slate-700 dark:text-white/70">
        Sağ üstteki + ile yeni hatim oluşturun, bağlantıyı paylaşın ve birlikte zikre devam edin.
      </p>
    </div>

    <UCard class="w-full max-w-lg bg-white/5 text-left">
      <div class="space-y-3">
        <h2 class="text-lg font-semibold">Nasıl çalışır?</h2>
        <ul class="list-inside list-disc text-sm text-slate-700 dark:text-white/70">
          <li>Hatim oluşturun ve paylaşım linkini kopyalayın.</li>
          <li>Katılımcılar linke girip gerekirse şifreyle katılır.</li>
          <li>Sayım herkes için gerçek zamanlı güncellenir.</li>
        </ul>
        <p v-if="hasQueryMessage" class="text-sm text-emerald-300">
          Hatim bağlantısı hazır. Yukarıdaki paylaş butonunu kullanabilirsiniz.
        </p>
      </div>
    </UCard>
  </div>
</template>
