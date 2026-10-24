<template>
  <AppLayout>
    <div class="infinite-canvas-embed">
      <iframe
        src="/infinite-canvas/?embedded=1&entry=canvas"
        title="无限画布"
        class="infinite-canvas-frame"
        allow="clipboard-read; clipboard-write"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue'
</script>

<style scoped>
.infinite-canvas-embed {
  height: calc(100vh - 128px);
  min-height: 620px;
  margin: -1rem;
  overflow: hidden;
  background: #fff;
}

.infinite-canvas-frame {
  display: block;
  width: 100%;
  height: 100%;
  border: 0;
  background: #fff;
}

@media (max-width: 760px) {
  .infinite-canvas-embed {
    height: calc(100dvh - 105px);
    min-height: 520px;
    margin: -1rem;
  }
}
</style>
