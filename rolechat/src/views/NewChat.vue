<template>
  <div class="chat-page" @dragover.prevent @drop.prevent="onDrop">
    <div ref="messagesEl" class="messages">
      <div class="messages-inner">
        <div
          v-for="m in messages"
          :key="m.id"
          class="message"
          :class="m.role"
        >
          <div v-if="m.role === 'assistant'" class="avatar assistant-avatar" title="Assistant">
            <img src="/characters/role2.jpg" alt="Assistant avatar" />
          </div>
          <div class="bubble">
            <p v-if="m.text" class="text" v-text="m.text"></p>
            <div v-if="m.attachments?.length" class="attachments">
              <template v-for="att in m.attachments" :key="att.id">
                <a
                  v-if="att.type === 'file' && !att.previewUrl"
                  class="file-chip"
                  :href="att.url"
                  target="_blank"
                  rel="noopener"
                >
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
        <div class="scroll-spacer"></div>
      </div>
    </div>

    <!-- Composer (fixed bottom) -->
    <div class="composer-wrap">
      <div class="composer-container">
        <!-- Attachment tray -->
        <transition name="tray">
          <div v-if="showAttachTray || pendingAttachments.length" class="attach-tray">
            <div class="pending" v-if="pendingAttachments.length">
              <div
                class="pending-item"
                v-for="att in pendingAttachments"
                :key="att.id"
              >
                <div v-if="att.type === 'image'" class="pending-thumb">
                  <img :src="att.previewUrl" :alt="att.name" />
                </div>
                <div v-else class="pending-file">
                  <span class="icon">📎</span>
                </div>
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

        <div class="composer" :class="{ focused: isFocused }">
          <!-- Top: typing area -->
          <textarea
            ref="textareaEl"
            v-model="inputText"
            class="input"
            placeholder="How can I help you today?"
            rows="1"
            @input="autoResize"
            @keydown.enter.exact.prevent="onEnter"
            @focus="isFocused = true"
            @blur="isFocused = false"
          />

          <!-- Bottom: action row -->
          <div class="action-row">
            <div class="left-tools" role="group" aria-label="Tools">
              <!-- 1st: File import -->
              <button class="tool-btn" @click="triggerFilePick" title="Import files" aria-label="Import files">
                <svg viewBox="0 0 24 24" class="i">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8m-6-6v6h6"
                        stroke="#111" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </button>
              <!-- 2nd: Mic -->
              <button
                class="tool-btn"
                :class="{ recording: isRecording }"
                :disabled="!speechSupported"
                @click="toggleRecording"
                :title="speechSupported ? (isRecording ? 'Stop voice' : 'Voice input') : 'Voice not supported'"
                aria-label="Voice input"
              >
                <svg viewBox="0 0 24 24" class="i">
                  <path d="M12 14a4 4 0 0 0 4-4V7a4 4 0 1 0-8 0v3a4 4 0 0 0 4 4Zm7-4a7 7 0 0 1-14 0M12 19v4m-4 0h8"
                        stroke="#111" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </button>
            </div>
            <div class="spacer"></div>
            <button class="send-btn" :disabled="!canSend" @click="send" title="Send" aria-label="Send">
              <svg viewBox="0 0 24 24" class="i send-i">
                <path d="M3 11v2l17 8-5-8 5-8-17 8Z"
                      stroke="#111" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
          </div>

          <!-- Hidden inputs -->
          <input ref="imageInput" type="file" accept="image/*" multiple class="hidden-input" @change="onPickImages" />
          <input ref="fileInput" type="file" multiple class="hidden-input" @change="onPickFiles" />
        </div>

        <!-- Removed model menu -->
      </div>
    </div>
  </div>
  
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, reactive, ref, computed } from 'vue'

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
}

const messages = reactive<ChatMessage[]>([])
const pendingAttachments = reactive<Attachment[]>([])
const inputText = ref('')
const isFocused = ref(false)
// removed model selector state

// simple current user state (replace with real auth/user store if available)
const userName = ref('joshua')
const userInitial = computed(() => (userName.value?.trim()?.charAt(0) || 'U').toUpperCase())

const messagesEl = ref<HTMLElement | null>(null)
const textareaEl = ref<HTMLTextAreaElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const imageInput = ref<HTMLInputElement | null>(null)

// Textarea sizing constraints
const TEXTAREA_MIN_PX = 72 // increase default height
const TEXTAREA_MAX_PX = 220

// Speech recognition (Web Speech API)
const speechSupported = typeof window !== 'undefined' && (('webkitSpeechRecognition' in window) || ('SpeechRecognition' in window))
let recognition: any = null
const isRecording = ref(false)

function setupRecognition() {
  if (!speechSupported || recognition) return
  const Rec: any = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition
  recognition = new Rec()
  recognition.lang = 'zh-CN' // default to Chinese; adjust as needed
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

const canSend = computed(() => inputText.value.trim().length > 0 || pendingAttachments.length > 0)

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

function send() {
  if (!canSend.value) return
  const msg: ChatMessage = {
    id: uid('m'),
    role: 'user',
    text: inputText.value.trim(),
    attachments: pendingAttachments.length ? [...pendingAttachments] : undefined,
    timestamp: Date.now(),
  }
  messages.push(msg)
  // clear composer
  inputText.value = ''
  autoResize()
  pendingAttachments.splice(0, pendingAttachments.length)
  scrollToBottom()

  // demo assistant echo
  setTimeout(() => {
    messages.push({
      id: uid('m'),
      role: 'assistant',
      text: `收到：${msg.text || '(附件)'}`,
      timestamp: Date.now(),
    })
    scrollToBottom()
  }, 600)
}

function onEnter(e: KeyboardEvent) {
  if (e.shiftKey) return // allow Shift+Enter for newline
  send()
}

// attachment tray toggle no longer used in the compact composer
const showAttachTray = ref(false)

// removed model selection functions

onMounted(() => {
  autoResize()
  scrollToBottom()
  setupRecognition()
})

onBeforeUnmount(() => {
  pendingAttachments.forEach(a => { if (a.previewUrl) URL.revokeObjectURL(a.previewUrl); if (a.url) URL.revokeObjectURL(a.url) })
  try { if (recognition && isRecording.value) recognition.stop() } catch {}
})
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

.messages-inner { width: 800px; margin: 0 auto; }

.scroll-spacer { height: 16px; }

.message { display: flex; margin: 8px 0; }
.message.user { justify-content: flex-end; }
.message.assistant { justify-content: flex-start; }

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
  /* move the composer slightly up from the bottom */
  bottom: 20px;
  padding: 0 12px calc(16px + env(safe-area-inset-bottom));
  background: linear-gradient(180deg, rgba(250,247,244,0) 0%, rgba(250,247,244,1) 40%);
}

.composer-container { width: 800px; margin: 0 auto; }

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
</style>
