<template>
  <div class="chats-page">
    <header class="header">
      <h1>历史对话记录</h1>
      <div class="search-box">
        <input
          v-model.trim="query"
          class="search-input"
          type="search"
          placeholder="搜索对话主题…"
          @keydown.escape="query = ''"
        />
      </div>
    </header>

    <section class="list-wrap" role="list" aria-label="历史对话">
      <div v-if="loading" class="loading">加载中...</div>
      <div v-else-if="error" class="error">{{ error }}</div>
      <template v-else-if="filteredChats.length">
        <article
          v-for="chat in filteredChats"
          :key="chat.id"
          class="chat-item"
          role="listitem"
        >
          <div class="meta">
            <template v-if="editingId === chat.id">
              <input
                ref="editInputRef"
                v-model="editTitle"
                class="title-input"
                type="text"
                :maxlength="80"
                @keyup.enter="confirmRename(chat)"
                @blur="confirmRename(chat)"
              />
            </template>
            <template v-else>
              <h3 class="title" :title="chat.title">{{ chat.title }}</h3>
            </template>
            <div class="sub">
              更新时间：{{ formatDateTime(chat.updated_at) }}
            </div>
          </div>
          <div class="actions">
            <button class="btn ghost" @click="startRename(chat)">重命名</button>
            <button class="btn danger" @click="removeChat(chat)">删除</button>
          </div>
        </article>
      </template>
      <div v-else class="empty">
        <p>暂无对话记录。</p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { getTopicsWithLimit } from '../api'

type Chat = {
  id: number
  title: string
  updated_at: string // ISO string
}

const chats = ref<Chat[]>([])
const query = ref('')
const loading = ref(false)
const error = ref('')

// rename state
const editingId = ref<number | null>(null)
const editTitle = ref('')
const editInputRef = ref<HTMLInputElement | null>(null)

const filteredChats = computed(() => {
  const q = query.value.toLowerCase()
  const arr = !q
    ? chats.value
    : chats.value.filter(c => c.title.toLowerCase().includes(q))
  // sort by updated_at desc
  return [...arr].sort((a, b) => b.updated_at.localeCompare(a.updated_at))
})

function formatDateTime(iso: string): string {
  try {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return iso
    return d.toLocaleString()
  } catch {
    return iso
  }
}

async function loadChats() {
  loading.value = true
  error.value = ''
  
  try {
    const response = await getTopicsWithLimit(30)
    chats.value = response.topics.map((topic: any) => ({
      id: topic.id,
      title: topic.title,
      updated_at: topic.updated_at
    }))
  } catch (err) {
    console.error('Failed to load chats:', err)
    error.value = '加载对话记录失败，请重试'
    chats.value = []
  } finally {
    loading.value = false
  }
}

function startRename(chat: Chat) {
  editingId.value = chat.id
  editTitle.value = chat.title
  nextTick(() => {
    editInputRef.value?.focus()
    editInputRef.value?.select()
  })
}

function confirmRename(chat: Chat) {
  if (editingId.value !== chat.id) return
  const name = editTitle.value.trim()
  if (!name) {
    // keep original if empty
    editingId.value = null
    return
  }
  const idx = chats.value.findIndex(c => c.id === chat.id)
  if (idx !== -1) {
    const current = chats.value[idx]
    if (current) {
      const updated: Chat = { id: current.id, title: name, updated_at: current.updated_at }
      chats.value.splice(idx, 1, updated)
      // Note: In a real app, you'd want to call an API to update the topic title on the server
    }
  }
  editingId.value = null
}

function removeChat(chat: Chat) {
  const ok = window.confirm(`确定要删除"${chat.title}"吗？此操作不可恢复。`)
  if (!ok) return
  chats.value = chats.value.filter(c => c.id !== chat.id)
  // Note: In a real app, you'd want to call an API to delete the topic on the server
}

onMounted(() => {
  loadChats()
})
</script>

<style scoped>
.chats-page {
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  box-sizing: border-box;
  text-align: left; /* ensure left alignment within page scope */
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.header h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.search-box { flex: 1; display: flex; justify-content: flex-end; }

.search-input {
  width: min(420px, 100%);
  padding: 10px 12px;
  border: 1px solid #d0d7de;
  border-radius: 8px;
  outline: none;
}
.search-input:focus {
  border-color: #409eff;
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.15);
}

.list-wrap {
  overflow: auto;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 8px;
  max-height: calc(100vh - 200px);
}

.loading, .error, .empty {
  padding: 40px 20px;
  text-align: center;
  color: #6b7280;
}

.error {
  color: #dc2626;
}

.chat-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border-radius: 8px;
  transition: background-color 0.15s ease;
}

.chat-item:hover {
  background-color: #f3f4f6;
}

.chat-item + .chat-item {
  border-top: 1px solid #e5e7eb;
}

.meta {
  flex: 1;
  min-width: 0; /* allow flex child to shrink */
}

.title {
  margin: 0 0 4px 0;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.title-input {
  width: 100%;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  padding: 2px 6px;
  outline: none;
}

.title-input:focus {
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.1);
}

.sub {
  font-size: 12px;
  color: #6b7280;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.btn {
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid transparent;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn.ghost {
  color: #6b7280;
  border-color: #d1d5db;
}

.btn.ghost:hover {
  color: #374151;
  border-color: #9ca3af;
  background-color: #f9fafb;
}

.btn.danger {
  color: #dc2626;
  border-color: #fecaca;
  background-color: #fef2f2;
}

.btn.danger:hover {
  color: #b91c1c;
  border-color: #f87171;
  background-color: #fee2e2;
}
</style>