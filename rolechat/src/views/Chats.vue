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
      <template v-if="filteredChats.length">
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
              <h3 class="title" :title="chat.topic">{{ chat.topic }}</h3>
            </template>
            <div class="sub">
              开始时间：{{ formatDateTime(chat.startedAt) }}
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

type Chat = {
  id: string
  topic: string
  startedAt: string // ISO string
}

const LS_KEY = 'rolechat.chats'

const chats = ref<Chat[]>([])
const query = ref('')

// rename state
const editingId = ref<string | null>(null)
const editTitle = ref('')
const editInputRef = ref<HTMLInputElement | null>(null)

const filteredChats = computed(() => {
  const q = query.value.toLowerCase()
  const arr = !q
    ? chats.value
    : chats.value.filter(c => c.topic.toLowerCase().includes(q))
  // sort by startedAt desc
  return [...arr].sort((a, b) => b.startedAt.localeCompare(a.startedAt))
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

function loadChats(): Chat[] {
  try {
    const raw = localStorage.getItem(LS_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as Chat[]
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function saveChats(list: Chat[]) {
  localStorage.setItem(LS_KEY, JSON.stringify(list))
}

function ensureSeedData() {
  const now = Date.now()
  const seedTopics = [
    '和小助手的第一次对话',
    '项目需求讨论',
    '旅行计划',
    '晚餐菜单头脑风暴',
    '英语口语练习',
    '健身计划与打卡',
    'OKR 月度回顾',
    '读书笔记：小王子',
    '周末亲子活动安排',
    '预算与记账优化',
    '新功能命名讨论',
    '学习路线规划：前端进阶'
  ]

  // 如果已有数据，尽量补足到 10 条；否则从空开始种子到 10 条
  const existing = chats.value
  const existingTitles = new Set(existing.map(c => c.topic))
  const needed = Math.max(0, 10 - existing.length)
  if (needed === 0 && existing.length > 0) return

  const seeds: Chat[] = []
  let offset = 0
  for (const t of seedTopics) {
    if (seeds.length >= needed) break
    if (existingTitles.has(t)) continue
    // 生成不同时间：从最近往前，错开分钟数
    const startedAt = new Date(now - (offset + 1) * 1000 * 60 * 37).toISOString()
    seeds.push({ id: cryptoRandomId(), topic: t, startedAt })
    offset++
  }

  // 如果种子不够 needed，就继续生成占位主题
  while (seeds.length < needed) {
    const i = seeds.length + 1
    const title = `新的对话 ${i}`
    const startedAt = new Date(now - (offset + 1) * 1000 * 60 * 29).toISOString()
    seeds.push({ id: cryptoRandomId(), topic: title, startedAt })
    offset++
  }

  const updated = existing.concat(seeds)
  chats.value = updated
  saveChats(updated)
}

function cryptoRandomId(): string {
  // Prefer crypto if available; fall back to Math.random
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    // @ts-ignore - randomUUID exists in modern browsers
    return crypto.randomUUID()
  }
  return 'id-' + Math.random().toString(36).slice(2, 10)
}

function startRename(chat: Chat) {
  editingId.value = chat.id
  editTitle.value = chat.topic
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
      const updated: Chat = { id: current.id, topic: name, startedAt: current.startedAt }
      chats.value.splice(idx, 1, updated)
      saveChats(chats.value)
    }
  }
  editingId.value = null
}

function removeChat(chat: Chat) {
  const ok = window.confirm(`确定要删除“${chat.topic}”吗？此操作不可恢复。`)
  if (!ok) return
  chats.value = chats.value.filter(c => c.id !== chat.id)
  saveChats(chats.value)
}

onMounted(() => {
  chats.value = loadChats()
  ensureSeedData()
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
  background: #fff;
}

.chat-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-radius: 8px;
}

.chat-item + .chat-item { border-top: 1px solid #f0f2f5; }

.meta { min-width: 0; }
.title {
  margin: 0 0 4px;
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: left; /* main topic title left-aligned */
}
.sub {
  color: #6b7280;
  font-size: 12px;
}

.actions { display: flex; gap: 8px; flex-shrink: 0; }

.btn {
  padding: 6px 10px;
  font-size: 13px;
  line-height: 1;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid transparent;
}
.btn.ghost {
  background: #f3f4f6;
  border-color: #e5e7eb;
}
.btn.ghost:hover { background: #e5e7eb; }
.btn.danger {
  color: #b91c1c;
  background: #fef2f2;
  border-color: #fecaca;
}
.btn.danger:hover { background: #fee2e2; }

.title-input {
  width: clamp(160px, 40vw, 520px);
  padding: 6px 8px;
  font-size: 16px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
}

.empty { padding: 24px; text-align: center; color: #6b7280; }
</style>
