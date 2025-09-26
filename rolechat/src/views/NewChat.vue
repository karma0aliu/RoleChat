<template>
  <div class="chat-page" @dragover.prevent @drop.prevent="onDrop">
  <div ref="messagesEl" class="messages" :class="{ 'is-role-select': hasRoleSelect }">
      <div class="messages-inner" :class="{ 'has-role-select': hasRoleSelect }">
        <template v-for="m in messages" :key="m.id">
          <div v-if="m.kind !== 'roleSelect'" class="message" :class="[m.role, { streaming: m.id === currentStreamingId }]">
            <div v-if="m.role === 'assistant'" class="avatar assistant-avatar" title="Assistant">
              <img :src="currentRoleAvatar" alt="Assistant avatar" />
            </div>
            <div class="bubble">
              <p v-if="m.text" class="text" v-text="m.text"></p>
              <div v-if="m.attachments?.length" class="attachments">
                <template v-for="att in m.attachments" :key="att.id">
                  <a v-if="att.type === 'file' && !att.previewUrl" class="file-chip" :href="att.url" target="_blank" rel="noopener">
                    <span class="icon">📎</span>
                    <span class="name">{{ att.name }}</span>
                    <span class="size">{{ formatSize(att.size) }}</span>
                  </a>
                  <div v-else-if="att.type === 'image'" class="image-thumb">
                    <img :src="att.previewUrl || att.url" :alt="att.name" />
                    <div class="thumb-caption">{{ att.name }}</div>
                  </div>
                </template>
              </div>
            </div>
            <div v-if="m.role === 'user'" class="avatar user-avatar" :title="userName">
              <span class="avatar-initial">{{ userInitial }}</span>
            </div>
          </div>
          <div v-else class="message assistant role-select-turn">
            <div class="avatar assistant-avatar" title="Assistant">
              <img src="/characters/role2.jpg" alt="Assistant avatar" />
            </div>
            <div class="bubble role-select-bubble">
              <div class="intro-assistant-inline">
                <h2>开始新的对话</h2>
                <p>请选择一个角色，我会以该角色的风格陪你聊天。</p>
                <p class="hint">点击下方任意角色卡片继续。</p>
              </div>
              <div class="cards-inline" v-if="!selectedRole">
                <CharacterCard
                  v-for="role in roles"
                  :key="role.id"
                  :avatar-url="role.avatar"
                  :name="role.name"
                  :title="role.title"
                  :handle="role.handle"
                  :status="role.desc"
                  :enable-tilt="true"
                  :show-user-info="false"
                  @select="chooseRole(role)"
                />
              </div>
              <div v-else class="chosen-role-note">你已选择：<strong>{{ selectedRole.name }}</strong>，可以开始聊天。</div>
            </div>
          </div>
        </template>
        <div class="scroll-spacer"></div>
      </div>
    </div>

    <div class="composer-wrap">
      <div class="composer-container">
        <transition name="tray">
          <div v-if="showAttachTray || pendingAttachments.length" class="attach-tray">
            <div class="pending" v-if="pendingAttachments.length">
              <div class="pending-item" v-for="att in pendingAttachments" :key="att.id">
                <div v-if="att.type === 'image'" class="pending-thumb"><img :src="att.previewUrl" :alt="att.name" /></div>
                <div v-else class="pending-file"><span class="icon">📎</span></div>
                <div class="meta">
                  <div class="name" :title="att.name">{{ att.name }}</div>
                  <div class="size">{{ formatSize(att.size) }}</div>
                </div>
                <button class="remove" @click="removePending(att.id)" aria-label="Remove attachment">✕</button>
              </div>
            </div>
            <div class="actions">
              <button class="tray-btn" @click="triggerImagePick" title="Add images" aria-label="Add images">
                <svg viewBox="0 0 24 24" class="i"><path d="M21 19V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v14m18 0a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2m18 0-5.5-6.5L13 16l-3.5-4.5L3 19m8-10a2 2 0 1 1-4 0 2 2 0 0 1 4 0Z"/></svg>
                <span>Images</span>
              </button>
              <button class="tray-btn" @click="triggerFilePick" title="Add files" aria-label="Add files">
                <svg viewBox="0 0 24 24" class="i"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8m-6-6v6h6"/></svg>
                <span>Files</span>
              </button>
              <div class="hint">Drag & drop files here</div>
            </div>
          </div>
        </transition>
        <div class="composer" :class="{ focused: isFocused, disabled: !selectedRole || !isLoggedIn }">
          <textarea
            ref="textareaEl"
            v-model="inputText"
            class="input"
            :placeholder="placeholderText"
            :readonly="!selectedRole || !isLoggedIn"
            rows="1"
            @input="autoResize"
            @keydown.enter.exact.prevent="onEnter"
            @focus="isFocused = true"
            @blur="isFocused = false"
          />
          <div v-if="!isLoggedIn" class="login-blocker">
            <p class="login-hint">需要登录后才能开始对话。</p>
            <button class="login-btn" @click="router.push({ path: '/login', query: { redirect: '/app/new' } })">去登录 / 注册</button>
          </div>
          <div class="action-row">
            <div class="left-tools" role="group" aria-label="Tools">
              <button class="tool-btn" @click="triggerFilePick" title="Import files" aria-label="Import files" :disabled="!selectedRole">
                <svg viewBox="0 0 24 24" class="i"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8m-6-6v6h6" stroke="#111" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </button>
              <button
                class="tool-btn"
                :class="{ recording: isRecording }"
                :disabled="!speechSupported || !selectedRole"
                @click="toggleRecording"
                :title="speechSupported ? (isRecording ? 'Stop voice' : 'Voice input') : 'Voice not supported'"
                aria-label="Voice input"
              >
                <svg viewBox="0 0 24 24" class="i"><path d="M12 14a4 4 0 0 0 4-4V7a4 4 0 1 0-8 0v3a4 4 0 0 0 4 4Zm7-4a7 7 0 0 1-14 0M12 19v4m-4 0h8" stroke="#111" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </button>
            </div>
            <div class="spacer"></div>
            <button class="send-btn" :disabled="!canSend" @click="send" title="Send" aria-label="Send">
              <svg viewBox="0 0 24 24" class="i send-i"><path d="M3 11v2l17 8-5-8 5-8-17 8Z" stroke="#111" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button>
          </div>
          <input ref="imageInput" type="file" accept="image/*" multiple class="hidden-input" @change="onPickImages" />
          <input ref="fileInput" type="file" multiple class="hidden-input" @change="onPickFiles" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, reactive, ref, computed } from 'vue'
import { isLoggedIn } from '../auth'
import { useRouter } from 'vue-router'
import { fetchWithAuth } from '../api'
import CharacterCard from '@/components/CharacterCard.vue'

type Role = 'user' | 'assistant'
type AttachmentType = 'image' | 'file'

interface Attachment {
  id: string
  type: AttachmentType
  name: string
  size: number
  file?: File
  url?: string
  previewUrl?: string
}

interface ChatMessage {
  id: string
  role: Role
  text: string
  attachments?: Attachment[]
  timestamp: number
  kind?: 'roleSelect'
}

// roles definition
interface RoleMeta { id: string; name: string; title: string; handle: string; avatar: string; desc: string }
const roles: RoleMeta[] = [
  { id: 'hermione', name: '赫敏', title: '智慧魔法师', handle: 'hermione', avatar: '/characters/role1.jpg', desc: '聪慧与逻辑' },
  { id: 'harry', name: '哈利', title: '勇敢的巫师', handle: 'harry', avatar: '/characters/role2.jpg', desc: '勇气与行动' },
  { id: 'ron', name: '罗恩', title: '忠诚的伙伴', handle: 'ron', avatar: '/characters/role3.jpg', desc: '幽默与支持' }
]

const ROLE_SELECT_INTRO = `开启一段全新的魔法对话。\n请选择一个角色，他/她将用独特的语气与视角陪你探索、提问、练习与畅想。\n不同角色代表不同的性格、思维方式与表达风格，你可以随时重新选择开始新的分支。\n\n点击下方任意角色卡片，立即开启你的魔法之旅。`

const messages = reactive<ChatMessage[]>([])
const pendingAttachments = reactive<Attachment[]>([])
const inputText = ref('')
const isFocused = ref(false)
const selectedRole = ref<RoleMeta | null>(null)
// 当前正在流式输出的 assistant 消息 id，用于固定宽度防止一行从中间开始视觉抖动
const currentStreamingId = ref<string | null>(null)
// 调试相关：是否开启流式调试（localStorage.setItem('streamDebug','0') 可关闭）
const STREAM_DEBUG = typeof window !== 'undefined' ? (localStorage.getItem('streamDebug') !== '0') : false
// 空闲等待指示（长时间未收到新增量）
const streamWaiting = ref(false)
// 当前会话 topic id（后端首次创建后返回）
const activeTopicId = ref<string | number | null>(null)
const hasRoleSelect = computed(() => messages.some(m => m.kind === 'roleSelect') && !selectedRole.value)

// Auth / user (deferred until login)
const userName = computed(() => (isLoggedIn.value ? (JSON.parse(localStorage.getItem('user')||'{}').username || 'User') : ''))
const userInitial = computed(() => (userName.value?.trim()?.charAt(0) || 'U').toUpperCase())
const router = useRouter()

const messagesEl = ref<HTMLElement | null>(null)
const textareaEl = ref<HTMLTextAreaElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const imageInput = ref<HTMLInputElement | null>(null)

// Textarea sizing constraints
const TEXTAREA_MIN_PX = 72
const TEXTAREA_MAX_PX = 220

// Speech recognition (Web Speech API)
const speechSupported = typeof window !== 'undefined' && (('webkitSpeechRecognition' in window) || ('SpeechRecognition' in window))
let recognition: any = null
const isRecording = ref(false)

function setupRecognition() {
  if (!speechSupported || recognition) return
  const Rec: any = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition
  recognition = new Rec()
  recognition.lang = 'zh-CN'
  recognition.interimResults = true
  recognition.continuous = false
  recognition.onresult = (event: any) => {
    let transcript = ''
    for (let i = event.resultIndex; i < event.results.length; i++) {
      transcript += event.results[i][0].transcript
    }
    if (transcript) inputText.value = (inputText.value + ' ' + transcript).trim()
  }
  recognition.onend = () => { isRecording.value = false }
  recognition.onerror = () => { isRecording.value = false }
}

function toggleRecording() {
  if (!speechSupported) return
  setupRecognition()
  if (!recognition) return
  if (isRecording.value) {
    try { recognition.stop() } catch {}
    isRecording.value = false
  } else {
    try { recognition.start(); isRecording.value = true } catch {}
  }
}

const canSend = computed(() => isLoggedIn.value && !!selectedRole.value && (inputText.value.trim().length > 0 || pendingAttachments.length > 0))
const placeholderText = computed(() => {
  if (!isLoggedIn.value) return '请先登录以开始对话'
  return selectedRole.value ? `和${selectedRole.value.name}聊点什么？` : '请选择一个角色开始对话'
})

function autoResize() {
  const el = textareaEl.value
  if (!el) return
  el.style.height = 'auto'
  const target = Math.max(TEXTAREA_MIN_PX, Math.min(el.scrollHeight, TEXTAREA_MAX_PX))
  el.style.height = target + 'px'
}

function scrollToBottom() {
  nextTick(() => {
    const el = messagesEl.value
    if (!el) return
    el.scrollTop = el.scrollHeight
  })
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function uid(prefix = 'id') { return `${prefix}_${Math.random().toString(36).slice(2, 10)}` }

function handleFiles(files: FileList | File[]) {
  const list = Array.from(files)
  for (const f of list) {
    const isImg = f.type.startsWith('image/')
    const att: Attachment = {
      id: uid('att'),
      type: isImg ? 'image' : 'file',
      name: f.name,
      size: f.size,
      file: f,
      previewUrl: isImg ? URL.createObjectURL(f) : undefined,
      url: !isImg ? URL.createObjectURL(f) : undefined,
    }
    pendingAttachments.push(att)
  }
}

function onPickFiles(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files?.length) handleFiles(target.files)
  target.value = ''
}

function onPickImages(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files?.length) handleFiles(target.files)
  target.value = ''
}

function removePending(id: string) {
  const idx = pendingAttachments.findIndex(a => a.id === id)
  if (idx >= 0) {
    const att = pendingAttachments[idx]
    pendingAttachments.splice(idx, 1)
    if (att?.previewUrl) URL.revokeObjectURL(att.previewUrl)
    if (att?.url) URL.revokeObjectURL(att.url)
  }
}

function triggerFilePick() { fileInput.value?.click() }
function triggerImagePick() { imageInput.value?.click() }

function onDrop(e: DragEvent) {
  if (e.dataTransfer?.files?.length) handleFiles(e.dataTransfer.files)
}

// removed clipboard paste handler and button

async function send() {
  if (!isLoggedIn.value) {
    router.push({ path: '/login', query: { redirect: '/app/new' } });
    return;
  }
  if (!canSend.value) return;

  const userText = inputText.value.trim();
  const userMsg: ChatMessage = {
    id: uid('m'),
    role: 'user',
    text: userText,
    attachments: pendingAttachments.length ? [...pendingAttachments] : undefined,
    timestamp: Date.now()
  };
  messages.push(userMsg);
  // 清空输入区
  inputText.value = '';
  autoResize();
  pendingAttachments.splice(0, pendingAttachments.length);
  scrollToBottom();

  // 发送到后端 /chat/message
  try {
    const payload: any = { content: userText };
    if (activeTopicId.value) payload.topic_id = activeTopicId.value;
    if (selectedRole.value) payload.role_id = selectedRole.value.id; // 若后端需要
    const res = await fetchWithAuth('/chat/message', {
      method: 'POST',
      body: JSON.stringify(payload)
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.detail || data.message || '发送失败');
    // 保存 topic id（首次创建）
    if (data.topic_id && !activeTopicId.value) activeTopicId.value = data.topic_id;
    // 如果后端直接返回 assistant 回复
    if (data.assistant) {
      const assistantText = typeof data.assistant === 'string' ? data.assistant : (data.assistant.text || data.assistant.content || '');
      if (assistantText) {
        messages.push({ id: uid('m'), role: 'assistant', text: assistantText, timestamp: Date.now() });
        scrollToBottom();
        return;
      }
    }
    // 不存在直接回复则尝试流式角色回复（需要已选角色）
    if (selectedRole.value) await streamRoleReply(userText);
  } catch (e: any) {
    messages.push({ id: uid('m'), role: 'assistant', text: '发送失败: ' + (e.message || e), timestamp: Date.now() });
  } finally {
    scrollToBottom();
  }
}

// 解析服务端以 SSE 形式返回的 data: 行；支持纯文本或 JSON
function extractSSEPayload(line: string): string {
  // line 可能是 "data: ..." 或空行
  if (!line) return ''
  if (line.startsWith('data:')) {
    return line.slice(5).trimStart()
  }
  return line
}

function parseMaybeJSON(raw: string): string {
  if (!raw) return ''
  // 某些后端会发送 [DONE] 作为结束标识
  if (raw === '[DONE]') return ''
  try {
    const obj = JSON.parse(raw)
    // 兼容多种字段
    return (
      obj.delta?.content ||
      obj.delta?.text ||
      obj.content ||
      obj.text ||
      obj.message ||
      ''
    )
  } catch {
    return raw // 不是 JSON，当作纯文本增量
  }
}

// 流式角色回复（/chat/role-reply/stream）
async function streamRoleReply(userText: string) {
  if (!selectedRole.value) return
  const assistantId = uid('m')
  messages.push({ id: assistantId, role: 'assistant', text: '', timestamp: Date.now() })
  currentStreamingId.value = assistantId
  scrollToBottom()

  const targetMsg = () => messages.find(m => m.id === assistantId)
  let pending = '' // 未完成的残余行缓冲
  let fullText = ''
  const controller = new AbortController()
  ;(targetMsg() as any)._abort = () => controller.abort()
  const startedAt = performance.now()
  let lastChunkAt = startedAt
  streamWaiting.value = false
  let idleInterval: number | null = null
  const IDLE_THRESHOLD = 3000 // 3s 未收到增量 => 显示等待
  const LONG_IDLE_THRESHOLD = 8000 // 8s 仍无增量 => 控制台告警
  function startIdleWatch() {
    if (idleInterval) return
    idleInterval = window.setInterval(() => {
      const now = performance.now()
      const idle = now - lastChunkAt
      if (idle > IDLE_THRESHOLD) {
        if (!streamWaiting.value) streamWaiting.value = true
      }
      if (STREAM_DEBUG && idle > LONG_IDLE_THRESHOLD) {
        console.warn('[stream][idle]', Math.round(idle) + 'ms with no data')
      }
    }, 1000)
  }
  function stopIdleWatch() {
    if (idleInterval) { clearInterval(idleInterval); idleInterval = null }
    streamWaiting.value = false
  }
  startIdleWatch()
  if (STREAM_DEBUG) console.debug('[stream] start', { assistantId, userText })
  try {
    const res = await fetchWithAuth('/chat/role-reply/stream', {
      method: 'POST',
      headers: { 'Accept': 'text/event-stream' },
      body: JSON.stringify({
        role_id: selectedRole.value.id,
        content: userText,
        topic_id: activeTopicId.value || undefined
      }),
      signal: controller.signal as any
    })
    if (!res.ok || !res.body) {
      const errData = await res.json().catch(() => ({}))
      throw new Error(errData.detail || '流式连接失败')
    }

    const reader = (res.body as ReadableStream).getReader()
    const decoder = new TextDecoder('utf-8')

    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      if (!value) continue
      const chunk = decoder.decode(value, { stream: true })
      if (!chunk) continue
      const now = performance.now()
      const dt = now - lastChunkAt
      lastChunkAt = now
      if (streamWaiting.value) streamWaiting.value = false
      pending += chunk
      const lines = pending.split(/\r?\n/)
      pending = lines.pop() || ''
      if (STREAM_DEBUG) console.debug('[stream] chunk', { size: chunk.length, dt: Math.round(dt) + 'ms', lines: lines.length })
      for (const rawLine of lines) {
        const payload = extractSSEPayload(rawLine.trim())
        if (!payload) continue
        if (payload === '[DONE]') { pending = ''; break }
        const delta = parseMaybeJSON(payload)
        if (!delta) continue
  fullText += delta
  const t = targetMsg()
  if (t) t.text = fullText.replace(/\\n/g, '\n')
      }
      scrollToBottom()
    }
    const last = extractSSEPayload(pending.trim())
    if (last && last !== '[DONE]') {
      const delta = parseMaybeJSON(last)
      if (delta) { fullText += delta; const t = targetMsg(); if (t) t.text = fullText.replace(/\\n/g, '\n') }
    }
    if (STREAM_DEBUG) {
      const dur = performance.now() - startedAt
      const near10s = dur > 9500 && dur < 10500
      console.debug('[stream] done', { durationMs: Math.round(dur), near10s })
      if (near10s) console.warn('[stream] finished around 10s boundary – possible external timeout')
    }
  } catch (err: any) {
    const t = targetMsg()
    const msg = (err?.name === 'AbortError')
      ? '[已取消]'
      : (/INCOMPLETE_CHUNKED_ENCODING/i.test(err?.message || '')
          ? '[连接中断: 可能服务器提前关闭 / 代理超时]'
          : `[流式失败] ${err.message || err}`)
  if (t) t.text = ((t.text || fullText).replace(/\\n/g,'\n')) + (t.text?.endsWith('\n') ? '' : '\n') + msg
    if (STREAM_DEBUG) console.error('[stream] error', err)
  } finally {
    currentStreamingId.value = null
    stopIdleWatch()
    scrollToBottom()
  }
}

function onEnter(e: KeyboardEvent) {
  if (e.shiftKey) return
  send()
}

const showAttachTray = ref(false)


onMounted(() => {
  autoResize()
  setupRecognition()
  const introMsg = messages.find(m => m.kind === 'roleSelect')
  if (introMsg) {
    introMsg.text = ROLE_SELECT_INTRO
  } else {
    messages.push({
      id: uid('m'),
      role: 'assistant',
      text: ROLE_SELECT_INTRO,
      timestamp: Date.now(),
      kind: 'roleSelect'
    })
  }
  scrollToBottom()
})

onBeforeUnmount(() => {
  pendingAttachments.forEach(a => { if (a.previewUrl) URL.revokeObjectURL(a.previewUrl); if (a.url) URL.revokeObjectURL(a.url) })
  try { if (recognition && isRecording.value) recognition.stop() } catch {}
})

const currentRoleAvatar = computed(() => selectedRole.value?.avatar || '/characters/role2.jpg')
function chooseRole(r: RoleMeta) {
  if (selectedRole.value?.id === r.id) return
  selectedRole.value = r
  for (let i = messages.length - 1; i >= 0; i--) {
    const m = messages[i]
    if (m && m.kind === 'roleSelect') messages.splice(i, 1)
  }
  messages.push({
    id: uid('m'),
    role: 'assistant',
    text: `你好，我是${r.name}（${r.title}）。现在以这个视角展开对话，随时向我提问或继续你的想法。`,
    timestamp: Date.now(),
  })
  nextTick(scrollToBottom)
}
</script>

<style scoped>
:root {
  --bg: #faf7f4;
  --card: #ffffff;
  --border: #e6e2de;
  --text: #2b2b2b;
  --muted: #8a8681;
  --primary: #d9a79c; /* soft rose */
  --primary-strong: #cd8f82;
  /* Offsets for sidebars/overlays; adjust these from layout when sidebar changes */
  --sidebar-left: 0px;
  --sidebar-right: 0px;
}

.chat-page {
  height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--bg);
}

.messages {
  flex: 1 1 auto;
  max-width: 820px;
  max-height: 800px;
  overflow-y: auto;
  padding: 16px calc(12px + var(--sidebar-right)) 180px calc(12px + var(--sidebar-left));
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE/Edge */
}

.messages::-webkit-scrollbar { width: 0; height: 0; }

.messages-inner { width: 800px; margin: 0; }

.scroll-spacer { height: 16px; }

.message { display: flex; margin: 8px 0; }
.message.user { justify-content: flex-end; }
.message.assistant { justify-content: flex-start; }
.message.streaming .bubble { position: relative; }
.message.streaming .bubble { min-width: 40%; }
@media (min-width: 900px) { .message.streaming .bubble { min-width: 55%; } }
@media (min-width: 1100px) { .message.streaming .bubble { min-width: 62%; } }

/* avatars */
.avatar { width: 32px; height: 32px; border-radius: 50%; overflow: hidden; flex: 0 0 auto; background: #0f1b2a; color: #fff; display: inline-flex; align-items: center; justify-content: center; font-weight: 700; }
.avatar img { width: 100%; height: 100%; object-fit: cover; display: block; }
.avatar-initial { font-size: 14px; line-height: 1; }
.message.user .avatar { margin-left: 8px; margin-right: 2px; }
.message.assistant .avatar { margin-right: 8px; margin-left: 2px; }

.bubble {
  max-width: min(80ch, 82vw);
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--text);
  border-radius: 14px;
  padding: 10px 12px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.04);
  white-space: pre-wrap;
  line-height: 1.45;
}

.message.user .bubble { background: #f7efe9; border-color: #eadfd6; }

.bubble .text { text-align: left; }


.attachments { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 8px; }

.file-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 10px;
  border: 1px dashed var(--border);
  color: var(--text);
  text-decoration: none;
  background: #fff;
}
.file-chip .icon { font-size: 14px; }
.file-chip .name { max-width: 220px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.file-chip .size { color: var(--muted); font-size: 12px; }

.image-thumb { border: 1px solid var(--border); border-radius: 10px; overflow: hidden; width: 160px; background: #fff; }
.image-thumb img { width: 100%; height: 110px; object-fit: cover; display: block; }
.thumb-caption { padding: 6px 8px; font-size: 12px; color: var(--muted); }

.composer-wrap {
  position: fixed;
  left: var(--sidebar-left); right: var(--sidebar-right); bottom: 0;
  bottom: 20px;
  padding: 0 12px calc(16px + env(safe-area-inset-bottom));
  background: linear-gradient(180deg, rgba(250,247,244,0) 0%, rgba(250,247,244,1) 40%);
}

.composer-container { width: 800px; margin: 0; }

.attach-tray {
  border: 1px solid var(--border);
  background: var(--card);
  border-radius: 14px;
  padding: 10px;
  margin-bottom: 10px;
  box-shadow: 0 6px 24px rgba(0,0,0,0.06);
}
.attach-tray .pending { display: flex; flex-wrap: wrap; gap: 10px; margin-bottom: 10px; }
.pending-item { display: flex; align-items: center; gap: 10px; padding: 8px; border: 1px dashed var(--border); border-radius: 10px; }
.pending-thumb { width: 56px; height: 40px; overflow: hidden; border-radius: 6px; border: 1px solid var(--border); }
.pending-thumb img { width: 100%; height: 100%; object-fit: cover; display: block; }
.pending-file .icon { font-size: 18px; }
.pending-item .meta { max-width: 40vw; }
.pending-item .name { font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.pending-item .size { font-size: 12px; color: var(--muted); }
.pending-item .remove { border: 0; background: transparent; cursor: pointer; color: #b04848; font-size: 16px; padding: 4px; }
.attach-tray .actions { display: flex; align-items: center; gap: 10px; }
.tray-btn { display: inline-flex; align-items: center; gap: 8px; padding: 8px 10px; border: 1px solid var(--border); background: #fff; border-radius: 10px; cursor: pointer; }
.tray-btn .i { width: 18px; height: 18px; fill: none; stroke: #6f6a65; stroke-width: 1.6; }
.attach-tray .hint { margin-left: auto; color: var(--muted); font-size: 12px; }

.composer {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
  /* move colored border to the outer composer */
  border: 1.5px solid #222;
  background: var(--card);
  border-radius: 16px;
  box-shadow: 0 6px 24px rgba(0,0,0,0.06);
}
.composer.focused { box-shadow: 0 8px 28px rgba(0,0,0,0.08); }

.login-blocker { 
  position: absolute; 
  inset: 0; 
  background: rgba(255,255,255,0.85); 
  backdrop-filter: blur(2px); 
  display: flex; 
  flex-direction: column; 
  align-items: center; 
  justify-content: center; 
  gap: 12px; 
  border-radius: 14px; 
}
.login-hint { margin: 0; font-size: 14px; color: #444; }
.login-btn { 
  padding: 8px 14px; 
  border-radius: 8px; 
  border: 1px solid var(--border); 
  background:#d4e8ec; 
  cursor:pointer; 
  font-size:14px; 
}
.login-btn:hover { background:#c2dde3; }

.icon-btn {
  width: 36px; height: 36px; display: inline-flex; align-items: center; justify-content: center;
  border: 1px solid var(--border); background: #fff; border-radius: 10px; cursor: pointer;
}
.icon-btn.optional { display: none; }
@media (min-width: 720px) { .icon-btn.optional { display: inline-flex; } }
.icon-btn .i { width: 18px; height: 18px; fill: none; stroke: #6f6a65; stroke-width: 1.8; }
.icon-btn.recording { border-color: var(--primary); box-shadow: inset 0 0 0 2px rgba(217,167,156,0.25); }

.input {
  width: 100%; border: none; outline: none; resize: none; background: transparent;
  padding: 8px 10px; font-size: 15px; line-height: 1.4; color: var(--text);
  /* increased min height and max height */
  min-height: 72px;
  max-height: 220px;
}
.input::placeholder { color: #b5b0ab; }
.action-row { display: flex; align-items: center; gap: 8px; }
.spacer { flex: 1 1 auto; }
.send-btn { width: 36px; height: 36px; border-radius: 10px; border: 1px solid var(--border); background: #d4e8ec; display: inline-flex; align-items: center; justify-content: center; cursor: pointer; }
.send-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.send-btn .send-i { width: 18px; height: 18px; fill: none; stroke: #111; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; display: block; }
.send-btn .send-i path { stroke: #111; stroke-width: 2; fill: none; stroke-linecap: round; stroke-linejoin: round; }

.right-controls { display: none; }

/* left tools for file + mic with black icons */
.left-tools { display: inline-flex; align-items: center; gap: 8px; padding-left: 4px; }
.tool-btn { width: 36px; height: 36px; border-radius: 10px; border: 1px solid var(--border); background: #d4e8ec; display: inline-flex; align-items: center; justify-content: center; cursor: pointer; }
.tool-btn .i { width: 18px; height: 18px; fill: none; stroke: #111; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; display: block; }
.tool-btn .i path { stroke: #111; stroke-width: 2; fill: none; stroke-linecap: round; stroke-linejoin: round; }
.tool-btn.recording { border-color: #000; box-shadow: inset 0 0 0 2px rgba(0,0,0,0.12); }

.send { width: 36px; height: 36px; border-radius: 10px; border: 0; background: var(--primary); color: #fff; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; }
.send:disabled { opacity: 0.6; cursor: not-allowed; }
.send .i { width: 18px; height: 18px; fill: none; stroke: #fff; stroke-width: 1.8; }

.hidden-input { display: none; }

/* Transitions */
.tray-enter-active, .tray-leave-active { transition: opacity .2s ease, transform .2s ease; }
.tray-enter-from, .tray-leave-to { opacity: 0; transform: translateY(6px); }
.menu-enter-active, .menu-leave-active { transition: opacity .15s ease, transform .15s ease; }
.menu-enter-from, .menu-leave-to { opacity: 0; transform: translateY(4px); }

/* Force icon visibility (defensive overrides) */
.tool-btn svg, .tool-btn svg path,
.send-btn svg, .send-btn svg path {
  stroke: #111 !important;
  opacity: 1 !important;
}

/* Horizontal role cards inside first assistant bubble */
.role-select-bubble { background: var(--card); }
.cards-inline { 
  display: flex;
  gap: 3px;
  justify-content: flex-start;
  align-items: stretch;
  padding: 12px 4px 0;
  margin-top: 4px;
  overflow-x: hidden;
  position: relative;
  scrollbar-width: none;
}
.cards-inline::-webkit-scrollbar { display: none; }
.cards-inline > * { flex: 0 0 280px; max-width:280px; }
.role-select-turn { width:100%; }
.role-select-turn .role-select-bubble { max-width:none; width:100%; }
.messages.is-role-select { max-width: 1320px; }
.messages-inner.has-role-select { width: min(840px,96vw); }
.messages-inner.has-role-select .bubble.role-select-bubble { box-shadow:none; border:0; background:transparent; }
.messages-inner.has-role-select .bubble.role-select-bubble .intro-assistant-inline { text-align:left; max-width:800px; margin:0 auto 4px; }
.messages-inner.has-role-select .cards-inline { max-width:800px; margin:0 auto; }
.messages-inner.has-role-select .bubble.role-select-bubble .intro-assistant-inline h2 { margin:0 0 8px; font-size:30px; }
.messages-inner.has-role-select .bubble.role-select-bubble .intro-assistant-inline p { margin:2px 0; line-height:1.5; }
.chosen-role-note { margin-top:12px; font-size:14px; color:#444; }

@media (max-width: 1400px) { .cards-inline { gap:24px; } .cards-inline > * { flex:0 0 260px; } }
@media (max-width: 1200px) { .cards-inline { gap:22px; } }
@media (max-width: 960px) {
  .cards-inline { gap: 2px; padding-bottom: 16px; scroll-snap-type: x mandatory; overflow-x:auto; }
  .cards-inline > * { flex: 0 0 250px; scroll-snap-align: start; }
  .messages-inner.has-role-select { width: min(1000px,96vw); }
}
@media (max-width: 640px) {
  .cards-inline { gap: 2px; }
  .cards-inline > * { flex: 0 0 220px; }
}

/* additional styles for role selection */
.role-select { max-width: 1100px; margin: 0 auto; padding: 40px 24px 160px; }
.intro-assistant { display: flex; gap: 24px; align-items: center; margin-bottom: 32px; }
.assistant-avatar.big { width: 72px; height: 72px; border-radius: 50%; overflow: hidden; flex-shrink:0; box-shadow:0 4px 16px rgba(0,0,0,0.15); }
.assistant-avatar.big img { width:100%; height:100%; object-fit:cover; }
.intro-text h2 { margin:0 0 8px; font-size:28px; font-weight:600; }
.intro-text p { margin:4px 0; color:#555; }
.intro-text .hint { font-size:13px; color:#888; }
.cards { display:grid; grid-template-columns:repeat(auto-fit,minmax(260px,1fr)); gap:32px; }
@media (max-width:800px){ .cards { gap:20px; } .intro-assistant { flex-direction:column; align-items:flex-start; } }
</style>

<!-- existing <style scoped> from original file continues below without duplication -->
