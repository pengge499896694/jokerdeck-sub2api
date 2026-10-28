<template>
  <section class="card p-6" aria-labelledby="commerce-links-title">
    <h2 id="commerce-links-title" class="text-lg font-semibold text-gray-900 dark:text-white">闲鱼与店铺链接</h2>
    <p class="mt-2 text-sm text-gray-500">链接保存后立即生效，用户重新进入页面或点击更新链接即可读取。无需重新部署。</p>
    <div class="mt-4 space-y-4">
      <div>
        <label for="commerce-xianyu" class="input-label">闲鱼下单链接</label>
        <input id="commerce-xianyu" v-model.trim="form.xianyu_order_url" type="url" maxlength="2048" class="input w-full" :disabled="loading || saving || !ready" placeholder="粘贴完整的闲鱼商品链接" />
      </div>
      <div>
        <label for="commerce-store" class="input-label">店铺销售链接</label>
        <input id="commerce-store" v-model.trim="form.store_url" type="url" maxlength="2048" class="input w-full" :disabled="loading || saving || !ready" placeholder="https://example.com/shop/..." />
      </div>
      <div>
        <label for="commerce-backup" class="input-label">店铺备用链接（选填）</label>
        <input id="commerce-backup" v-model.trim="form.store_backup_url" type="url" maxlength="2048" class="input w-full" :disabled="loading || saving || !ready" placeholder="店铺主域名无法访问时使用的备用地址" />
        <p class="mt-1 text-xs text-gray-500">请填写同一家店铺的有效备用域名或地址。留空时不显示备用入口。</p>
      </div>
    </div>
    <p v-if="error" role="alert" class="mt-3 text-sm text-red-600">{{ error }}</p>
    <p v-if="saved" role="status" class="mt-3 text-sm text-green-600">链接已保存，前台将使用最新地址。</p>
    <div class="mt-4 flex gap-3">
      <button type="button" class="btn btn-primary" :disabled="loading || saving || !ready" @click="save">{{ saving ? '保存中...' : '保存链接' }}</button>
      <button type="button" class="btn btn-secondary" :disabled="loading || saving" @click="load">{{ loading ? '读取中...' : '重新读取' }}</button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { getCommerceLinks, updateCommerceLinks, type CommerceLinks } from '@/api/commerceLinks'
const form = reactive<CommerceLinks>({ xianyu_order_url: '', store_url: '', store_backup_url: '' })
const loading = ref(false)
const saving = ref(false)
const ready = ref(false)
const saved = ref(false)
const error = ref('')
async function load() {
  loading.value = true; error.value = ''; saved.value = false
  try { Object.assign(form, await getCommerceLinks(true)); ready.value = true }
  catch { ready.value = false; error.value = '无法读取链接配置，请重试。' }
  finally { loading.value = false }
}
async function save() {
  if (saving.value || !ready.value) return
  saving.value = true; error.value = ''; saved.value = false
  try { Object.assign(form, await updateCommerceLinks({ ...form })); saved.value = true }
  catch (err) { error.value = (err as { message?: string })?.message || '链接保存失败，请重试。' }
  finally { saving.value = false }
}
onMounted(load)
</script>
