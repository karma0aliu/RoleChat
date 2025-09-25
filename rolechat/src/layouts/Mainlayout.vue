<template>
<div class="main-layout">
  <aside :class="{ collapsed: isSidebarCollapsed }">
  <div class="sidebar-header">
    <svg class="role-logo" width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="12" cy="8" r="4" stroke="currentColor" stroke-width="2"/>
      <path d="M4 20c0-4 4-6 8-6s8 2 8 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
    </svg>
    <span class="text">RoleChat</span>
  </div>

  <div class="sidebar-main">
    <div class="sidebar-top">
  <router-link to="/api/new" class="new-chat-btn">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 5V19M5 12H19" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
      <span class="text">{{ t('newChat') }}</span>
    </router-link>
        </div>

        <nav class="main-nav">
  <router-link to="/api/chats" class="nav-item">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v10z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
      <span class="text">{{ t('chats') }}</span>
    </router-link>
        </nav>

        <div class="history">
    <div class="history-title">
      <span class="text">{{ t('recents') }}</span>
        </div>
    <div
      v-for="topic in topics"
      :key="topic.id"
      :class="['history-item-wrapper', { 'has-open-menu': openMenuTopic === topic.id }]"
      @mouseenter="hoveredTopic = topic.id"
      @mouseleave="hoveredTopic = null"
    >
            <button
                @click="navigate({ view: 'chat', topicId: topic.id })"
                class="history-item"
                :class="{ active: isActive(topic.id) }"
            >
                <span class="text">{{ topic.title }}</span>
            </button>
            <div class="actions" v-show="hoveredTopic === topic.id || openMenuTopic === topic.id">
                <button @click.stop="toggleMenu(topic.id)" class="more-btn" :class="{ active: openMenuTopic === topic.id }">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12C11 12.5523 11.4477 13 12 13Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12C18 12.5523 18.4477 13 19 13Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12C4 12.5523 4.44772 13 5 13Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                </button>
                <div v-if="openMenuTopic === topic.id" class="dropdown-menu">
          <button @click.stop="renameTopic(topic.id)">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M16.86 4.49998L19.5 7.13998L9.14 17.5H6.5V14.86L16.86 4.49998Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M14.5 6.49998L17.5 9.49998" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            {{ t('rename') }}
                    </button>
          <button @click.stop="deleteTopic(topic.id)">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M3 6H5H21" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M8 6V4C8 3.46957 8.21071 2.96086 8.58579 2.58579C8.96086 2.21071 9.46957 2 10 2H14C14.5304 2 15.0391 2.21071 15.4142 2.58579C15.7893 2.96086 16 3.46957 16 4V6M19 6V20C19 20.5304 18.7893 21.0391 18.4142 21.4142C18.0391 21.7893 17.5304 22 17 22H7C6.46957 22 5.96086 21.7893 5.58579 21.4142C5.21071 21.0391 5 20.5304 5 20V6H19Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            {{ t('delete') }}
                    </button>
                </div>
            </div>
        </div>
        </div>
    </div>

  <div class="sidebar-footer">
        <div class="user-area">
      <div class="user-profile" @click="toggleUserMenu">
        <div class="avatar">J</div>
        <span class="text username">{{ user.username }}</span>
      </div>
      <div class="user-menu" v-show="userMenuOpen">
        <div class="user-info">
          <div class="avatar large">J</div>
          <div class="meta">
            <div class="name">{{ user.username }}</div>
            <div class="email">{{ user.email }}</div>
          </div>
        </div>
        <div class="user-actions">
          <button class="user-action" @click="toggleLanguage">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 21C16.9706 21 21 16.9706 21 12C21 7.02944 16.9706 3 12 3C7.02944 3 3 7.02944 3 12C3 16.9706 7.02944 21 12 21Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M2 12H22" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M12 2C14.5 5.5 15.5 8.5 15.5 12C15.5 15.5 14.5 18.5 12 22C9.5 18.5 8.5 15.5 8.5 12C8.5 8.5 9.5 5.5 12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            <span>{{ t('language') }}: {{ languageLabel }}</span>
          </button>
          <button class="user-action" @click="logout">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M10 17L15 12L10 7" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M15 12H3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M21 21V3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            <span>{{ t('logout') }}</span>
          </button>
        </div>
      </div>
    </div>
    <button @click="toggleSidebar" class="toggle-btn">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="currentColor" d="m14 17l-5-5l5-5l1.41 1.41L11.83 12l3.58 3.59z"/></svg>
        </button>
    </div>

    </aside>

  <main class="content-area" :style="contentStyle">
    <div class="content-wrap">
      <router-view />
    </div>
  </main>
</div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
activeTopicId: [Number, String, null]
})
const emit = defineEmits(['navigate'])

const isSidebarCollapsed = ref(false);
const userMenuOpen = ref(false);
const user = ref({ username: 'joshua', email: 'joshua@example.com' });
const language = ref('zh');
const i18n = {
  en: {
    newChat: 'New chat',
    chats: 'Chats',
    recents: 'Recents',
    rename: 'Rename',
    delete: 'Delete',
    language: 'Language',
    logout: 'Logout',
    username: 'Username',
    email: 'Email',
  },
  zh: {
    newChat: '新对话',
    chats: '历史对话',
    recents: '最近',
    rename: '重命名',
    delete: '删除',
    language: '语言',
    logout: '登出',
    username: '用户名',
    email: '邮箱',
  }
};
const dict = computed(() => i18n[language.value]);
const t = (key) => dict.value[key] || key;
const languageLabel = computed(() => (language.value === 'zh' ? '中文' : 'English'));

const topics = ref([
  { id: 1, title: 'CSS priority tag styling issue - advanced layout', mode: 'chat' },
  { id: 2, title: 'Web page layout optimization - advanced', mode: 'chat' },
  { id: 3, title: 'Advanced web layout optimization', mode: 'chat' },
  { id: 4, title: 'Django database thread safety', mode: 'chat' },
  { id: 5, title: 'Untitled', mode: 'chat' },
]);

const hoveredTopic = ref(null);
const openMenuTopic = ref(null);

const toggleMenu = (topicId) => {
  if (openMenuTopic.value === topicId) {
    openMenuTopic.value = null;
  } else {
    openMenuTopic.value = topicId;
  }
};

const renameTopic = (topicId) => {
  console.log('Rename topic:', topicId);
  openMenuTopic.value = null;
};

const deleteTopic = (topicId) => {
  console.log('Delete topic:', topicId);
  topics.value = topics.value.filter(t => t.id !== topicId);
  openMenuTopic.value = null;
};

const openUserMenu = () => { userMenuOpen.value = true; };
const closeUserMenu = () => { userMenuOpen.value = false; };
const toggleUserMenu = (e) => { e?.stopPropagation?.(); userMenuOpen.value = !userMenuOpen.value; };
const toggleLanguage = () => { language.value = language.value === 'zh' ? 'en' : 'zh'; };
const logout = () => { console.log('logout clicked'); };

const onDocumentClick = (e) => {
  const target = e.target;
  if (!(target instanceof Element)) return;
  if (!target.closest('.history-item-wrapper')) {
    openMenuTopic.value = null;
  }
  if (!target.closest('.user-area')) {
    userMenuOpen.value = false;
  }
};

onMounted(() => {
  document.addEventListener('click', onDocumentClick);
});
onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick);
});

const navigate = (payload) => {
emit('navigate', payload);
}

const isActive = (topicId) => {
return props.activeTopicId === topicId;
}

const toggleSidebar = () => {
isSidebarCollapsed.value = !isSidebarCollapsed.value;
}

const contentStyle = computed(() => ({
paddingLeft: isSidebarCollapsed.value ? 'var(--sidebar-collapsed-width)' : 'var(--sidebar-width)'
}));
</script>

<style scoped>
.main-layout {
display: flex;
height: 100%;
width: 100%;
overflow: hidden;
--sidebar-width: 15vw;
--sidebar-collapsed-width: 64px;
}

.content-area {
  flex-grow: 1;
  height: 100vh;
  overflow: hidden; /* let inner views manage their own scroll */
  transition: padding-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
}

.content-wrap {
  flex: 1;
  min-height: 0; /* allow children to shrink for internal scroll containers */
  display: flex;
  flex-direction: column;
}

aside {
width: var(--sidebar-width);
height: 100vh;
position: fixed;
left: 0;
top: 0;
background-color: #eceee8;;
border-right: 1px solid var(--border-color);
display: flex;
flex-direction: column;
transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
z-index: 10;
padding: 16px;
box-sizing: border-box;
}
aside.collapsed {
width: var(--sidebar-collapsed-width);
padding: 16px 8px;
}

.sidebar-header {
flex-shrink: 0;
padding: 4px 10px;
margin-bottom: 24px;
font-size: 1.125rem; /* 18px */
font-weight: 600;
text-align: left;
display: flex;
align-items: center;
gap: 10px;
}

.role-logo {
  width: 20px;
  height: 20px;
  color: var(--text-primary);
}

.collapsed .sidebar-header .text {
display: none;
}

.sidebar-main {
flex-grow: 1;
overflow-y: auto;
overflow-x: hidden;
}

.sidebar-footer {
flex-shrink: 0;
padding-top: 16px;
border-top: 1px solid var(--border-color);
display: flex;
align-items: center;
justify-content: space-between;
}
.collapsed .sidebar-footer {
flex-direction: column;
gap: 16px;
align-items: center;
}

.text {
transition: opacity 0.2s ease, width 0.2s ease, display 0.2s;
white-space: nowrap;
}
.collapsed .text {
opacity: 0;
width: 0;
overflow: hidden;
}

.sidebar-top { margin-bottom: 12px; }
.new-chat-btn, .nav-item, .history-item {
display: flex;
align-items: center;
gap: 12px;
padding: 10px;
border-radius: 8px;
width: 100%;
text-align: left;
cursor: pointer;
font-family: inherit;
background: none;
border: none;
}
.new-chat-btn, .nav-item { text-decoration: none; }
.collapsed .new-chat-btn,
.collapsed .nav-item {
justify-content: center;
}
.new-chat-btn:hover, .nav-item:hover {
  background-color: #e3e3d9;
  color: inherit;
}

.history-item:hover,
.history-item:focus-visible {
  background-color: transparent;
  color: inherit;
}
.history-item:active {
  background-color: transparent;
  color: inherit;
}

.new-chat-btn { font-weight: 500; color: var(--text-primary); }
.nav-item { font-size: 14px; font-weight: 500; color: var(--text-secondary); }
.main-nav { display: flex; flex-direction: column; gap: 12px; }

.history { margin-top: 24px; }
.history-title {
font-size: 12px;
font-weight: 600;
color: var(--text-secondary);
text-transform: uppercase;
padding: 0 10px 12px;
text-align: left;
}
.collapsed .history-title {
padding: 0;
}
.collapsed .history-title .text {
display: none;
}

aside.collapsed .history { display: none; }

.history-item {
padding: 8px 10px;
font-size: 14px;
white-space: nowrap;
overflow: hidden;
text-overflow: ellipsis;
color: var(--text-secondary);
flex-grow: 1;
width: 0;
}
.history-item.active {
color: var(--text-primary);
font-weight: 600;
background-color: transparent;
}

.history-item.active:hover,
.history-item.active:focus-visible,
.history-item.active:active {
  background-color: transparent;
  color: inherit;
}

.history-item:focus { outline: none; }

.history-item-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  z-index: 0;
  border-radius: 8px;
}

.history-item-wrapper:hover {
  background-color: #e3e3d9;
}

.history-item-wrapper:hover::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  background-color: #e3e3d9;
  border-top-right-radius: 8px;
  border-bottom-right-radius: 8px;
}

/* In collapsed mode, the actions use -8px; keep the extension in sync */
.collapsed .history-item-wrapper:hover::after {
  right: -8px;
  width: 8px;
}

.history-item-wrapper .history-item {
  flex-grow: 1;
  padding-right: 30px;
  width: 12vw;
  margin-top: 10px;
  max-width: 12vw;
  min-width: 0;
  box-sizing: border-box;
}

.actions {
  position: absolute;
  right: -7px;
  display: flex;
  align-items: center;
  top: 50%;
  transform: translateY(-25%);
}

aside.collapsed .actions {
  right: -8px;
}

.more-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  color: var(--text-secondary);
}
.more-btn:focus { outline: none; }
.more-btn:focus-visible { outline: none; color: var(--text-primary); }

.more-btn.active,
.more-btn:hover {
  background-color: transparent;
  color: var(--text-primary);
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  background-color: #fff;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  z-index: 1000;
  width: 120px;
  padding: 8px 0;
}

.history-item-wrapper.has-open-menu {
  z-index: 100;
}

.dropdown-menu button {
  display: block;
  width: 100%;
  padding: 8px 16px;
  text-align: left;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dropdown-menu button:hover {
  background-color: var(--accent-hover-bg);
}

.user-area { position: relative; }
.user-profile { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.avatar { width: 28px; height: 28px; border-radius: 50%; background-color: #111827; color: white; display: flex; align-items: center; justify-content: center; font-weight: bold; font-size: 14px; flex-shrink: 0; }
.toggle-btn { background: transparent; border: none; cursor: pointer; color: var(--text-secondary); padding: 5px; border-radius: 50%; display: flex; align-items: center; justify-content: center; box-shadow: none; outline: none; appearance: none; -webkit-appearance: none; }
.toggle-btn:focus, .toggle-btn:active { background: transparent; box-shadow: none; outline: none; }
.toggle-btn:hover { background-color: var(--accent-hover-bg); color: var(--text-primary); }
.toggle-btn svg { transition: transform 0.3s ease; }
.collapsed .toggle-btn svg { transform: rotate(180deg); }

.collapsed .sidebar-footer .user-profile {
gap: 0;
}
.collapsed .sidebar-footer .username {
display: none;
}

.user-menu {
  position: absolute;
  bottom: 42px;
  width: 240px;
  background: #fff;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12);
  padding: 10px;
  z-index: 30;
}
.user-info { display: flex; gap: 10px; align-items: center; padding: 6px 6px 10px; border-bottom: 1px solid var(--border-color); }
.avatar.large { width: 36px; height: 36px; font-size: 16px; }
.user-info .meta { display: flex; flex-direction: column; }
.user-info .name { font-weight: 600; color: var(--text-primary); }
.user-info .email { font-size: 12px; color: var(--text-secondary); }
.user-actions { display: flex; flex-direction: column; padding-top: 6px; }
.user-action { display: flex; align-items: center; gap: 8px; padding: 8px; border: none; background: none; text-align: left; cursor: pointer; border-radius: 8px; color: var(--text-secondary); }
.user-action:hover { background-color: var(--accent-hover-bg); color: var(--text-primary); }

aside.collapsed .user-menu { left: 50%; transform: translateX(-50%); }

</style>