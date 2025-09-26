<template>
  <div
    ref="wrapRef"
    :class="['pc-card-wrapper', className, { selectable }]"
    :style="cardStyle"
    @click="onSelect"
  >
    <section ref="cardRef" class="pc-card">
      <div class="pc-inside">
        <div class="pc-shine" />
        <div class="pc-glare" />
        <div class="pc-content pc-avatar-content">
          <img
            class="avatar"
            :src="avatarUrl"
            :alt="`${name || '角色'} avatar`"
            loading="lazy"
            @error="handleAvatarError"
          />
          <div v-if="showUserInfo" class="pc-user-info">
            <div class="pc-user-details">
              <div class="pc-mini-avatar">
                <img
                  :src="miniAvatarUrl || avatarUrl"
                  :alt="`${name || '角色'} mini avatar`"
                  loading="lazy"
                  @error="handleMiniAvatarError"
                />
              </div>
              <div class="pc-user-text">
                <div class="pc-handle">@{{ handle }}</div>
                <div class="pc-status">{{ status }}</div>
              </div>
            </div>
            <button
              v-if="contactText"
              class="pc-contact-btn"
              @click.stop="handleContactClick"
              style="pointer-events: auto"
              type="button"
              :aria-label="`Contact ${name || 'user'}`"
            >
              {{ contactText }}
            </button>
          </div>
        </div>
        <div class="pc-content">
          <div class="pc-details">
            <h3>{{ name }}</h3>
            <p>{{ title }}</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, useTemplateRef } from 'vue'

interface Props {
  avatarUrl?: string
  iconUrl?: string
  grainUrl?: string
  behindGradient?: string
  innerGradient?: string
  showBehindGradient?: boolean
  className?: string
  enableTilt?: boolean
  miniAvatarUrl?: string
  name?: string
  title?: string
  handle?: string
  status?: string
  contactText?: string
  showUserInfo?: boolean
  selectable?: boolean
  textTopOffset?: string
  titleGap?: string
}

const props = withDefaults(defineProps<Props>(), {
  avatarUrl: '/characters/role1.jpg',
  iconUrl: undefined,
  grainUrl: undefined,
  behindGradient: undefined,
  innerGradient: undefined,
  showBehindGradient: true,
  className: '',
  enableTilt: true,
  miniAvatarUrl: undefined,
  name: '角色',
  title: '',
  handle: 'role',
  status: '在线',
  contactText: '',
  showUserInfo: false,
  selectable: true,
  textTopOffset: '1em',
  titleGap: '6px'
})

const emit = defineEmits<{
  contactClick: []
  select: []
}>()

const wrapRef = useTemplateRef<HTMLDivElement>('wrapRef')
const cardRef = useTemplateRef<HTMLElement>('cardRef')

const DEFAULT_BEHIND_GRADIENT =
  'radial-gradient(farthest-side circle at var(--pointer-x) var(--pointer-y),hsla(266,100%,90%,var(--card-opacity)) 4%,hsla(266,50%,80%,calc(var(--card-opacity)*0.75)) 10%,hsla(266,25%,70%,calc(var(--card-opacity)*0.5)) 50%,hsla(266,0%,60%,0) 100%),radial-gradient(35% 52% at 55% 20%,#00ffaac4 0%,#073aff00 100%),radial-gradient(100% 100% at 50% 50%,#00c1ffff 1%,#073aff00 76%),conic-gradient(from 124deg at 50% 50%,#c137ffff 0%,#07c6ffff 40%,#07c6ffff 60%,#c137ffff 100%)'

const DEFAULT_INNER_GRADIENT = 'linear-gradient(145deg,#60496e8c 0%,#71C4FF44 100%)'

const ANIMATION_CONFIG = {
  SMOOTH_DURATION: 600,
  INITIAL_DURATION: 1500,
  INITIAL_X_OFFSET: 70,
  INITIAL_Y_OFFSET: 60
} as const

const clamp = (value: number, min = 0, max = 100): number => Math.min(Math.max(value, min), max)
const round = (value: number, precision = 3): number => parseFloat(value.toFixed(precision))
const adjust = (value: number, fromMin: number, fromMax: number, toMin: number, toMax: number): number =>
  round(toMin + ((toMax - toMin) * (value - fromMin)) / (fromMax - fromMin))
const easeInOutCubic = (x: number): number => (x < 0.5 ? 4 * x * x * x : 1 - Math.pow(-2 * x + 2, 3) / 2)

let rafId: number | null = null

const updateCardTransform = (offsetX: number, offsetY: number, card: HTMLElement, wrap: HTMLElement) => {
  const width = card.clientWidth
  const height = card.clientHeight
  const percentX = clamp((100 / width) * offsetX)
  const percentY = clamp((100 / height) * offsetY)
  const centerX = percentX - 50
  const centerY = percentY - 50
  const properties = {
    '--pointer-x': `${percentX}%`,
    '--pointer-y': `${percentY}%`,
    '--background-x': `${adjust(percentX, 0, 100, 35, 65)}%`,
    '--background-y': `${adjust(percentY, 0, 100, 35, 65)}%`,
    '--pointer-from-center': `${clamp(Math.hypot(percentY - 50, percentX - 50) / 50, 0, 1)}`,
    '--pointer-from-top': `${percentY / 100}`,
    '--pointer-from-left': `${percentX / 100}`,
    '--rotate-x': `${round(-(centerX / 5))}deg`,
    '--rotate-y': `${round(centerY / 4)}deg`
  } as Record<string, string>
  Object.entries(properties).forEach(([property, value]) => {
    wrap.style.setProperty(property, value)
  })
}

const createSmoothAnimation = (
  duration: number,
  startX: number,
  startY: number,
  card: HTMLElement,
  wrap: HTMLElement
) => {
  const startTime = performance.now()
  const targetX = wrap.clientWidth / 2
  const targetY = wrap.clientHeight / 2
  const animationLoop = (currentTime: number) => {
    const elapsed = currentTime - startTime
    const progress = clamp(elapsed / duration)
    const easedProgress = easeInOutCubic(progress)
    const currentX = adjust(easedProgress, 0, 1, startX, targetX)
    const currentY = adjust(easedProgress, 0, 1, startY, targetY)
    updateCardTransform(currentX, currentY, card, wrap)
    if (progress < 1) {
      rafId = requestAnimationFrame(animationLoop)
    }
  }
  rafId = requestAnimationFrame(animationLoop)
}

const cancelAnimation = () => {
  if (rafId) {
    cancelAnimationFrame(rafId)
    rafId = null
  }
}

const handlePointerMove = (event: PointerEvent) => {
  const card = cardRef.value
  const wrap = wrapRef.value
  if (!card || !wrap || !props.enableTilt) return
  const rect = card.getBoundingClientRect()
  updateCardTransform(event.clientX - rect.left, event.clientY - rect.top, card, wrap)
}

const handlePointerEnter = () => {
  const card = cardRef.value
  const wrap = wrapRef.value
  if (!card || !wrap || !props.enableTilt) return
  cancelAnimation()
  wrap.classList.add('active')
  card.classList.add('active')
}

const handlePointerLeave = (event: PointerEvent) => {
  const card = cardRef.value
  const wrap = wrapRef.value
  if (!card || !wrap || !props.enableTilt) return
  createSmoothAnimation(ANIMATION_CONFIG.SMOOTH_DURATION, event.offsetX, event.offsetY, card, wrap)
  wrap.classList.remove('active')
  card.classList.remove('active')
}

const cardStyle = computed(() => ({
  '--icon': props.iconUrl ? `url(${props.iconUrl})` : 'none',
  '--grain': props.grainUrl ? `url(${props.grainUrl})` : 'none',
  '--behind-gradient': props.showBehindGradient ? (props.behindGradient ?? DEFAULT_BEHIND_GRADIENT) : 'none',
  '--inner-gradient': props.innerGradient ?? DEFAULT_INNER_GRADIENT,
  '--card-text-top': props.textTopOffset,
  '--card-title-gap': props.titleGap
}))

const handleContactClick = () => emit('contactClick')

const handleAvatarError = (event: Event) => {
  const target = event.target as HTMLImageElement
  target.style.display = 'none'
}

const handleMiniAvatarError = (event: Event) => {
  const target = event.target as HTMLImageElement
  target.style.opacity = '0.5'
  target.src = props.avatarUrl
}

function onSelect() {
  if (props.selectable) emit('select')
}

onMounted(() => {
  if (!props.enableTilt) return
  const card = cardRef.value
  const wrap = wrapRef.value
  if (!card || !wrap) return
  card.addEventListener('pointerenter', handlePointerEnter)
  card.addEventListener('pointermove', handlePointerMove)
  card.addEventListener('pointerleave', handlePointerLeave)
  const initialX = wrap.clientWidth - ANIMATION_CONFIG.INITIAL_X_OFFSET
  const initialY = ANIMATION_CONFIG.INITIAL_Y_OFFSET
  updateCardTransform(initialX, initialY, card, wrap)
  createSmoothAnimation(ANIMATION_CONFIG.INITIAL_DURATION, initialX, initialY, card, wrap)
})

onUnmounted(() => {
  const card = cardRef.value
  if (card) {
    card.removeEventListener('pointerenter', handlePointerEnter)
    card.removeEventListener('pointermove', handlePointerMove)
    card.removeEventListener('pointerleave', handlePointerLeave)
  }
  cancelAnimation()
})
</script>

<style scoped>
/* (Card styles copied from provided code, trimmed only for minor additions) */
.pc-card-wrapper { perspective: 500px; transform: translate3d(0,0,0.1px); position: relative; touch-action: none; cursor: pointer; }
.pc-card-wrapper.selectable:active { scale: .98; }
/* full original styles retained */
.pc-card-wrapper { --pointer-x:50%;--pointer-y:50%;--pointer-from-center:0;--pointer-from-top:.5;--pointer-from-left:.5;--card-opacity:0;--rotate-x:0deg;--rotate-y:0deg;--background-x:50%;--background-y:50%;--grain:none;--icon:none;--behind-gradient:none;--inner-gradient:none;--sunpillar-1:hsl(2,100%,73%);--sunpillar-2:hsl(53,100%,69%);--sunpillar-3:hsl(93,100%,69%);--sunpillar-4:hsl(176,100%,76%);--sunpillar-5:hsl(228,100%,74%);--sunpillar-6:hsl(283,100%,73%);--sunpillar-clr-1:var(--sunpillar-1);--sunpillar-clr-2:var(--sunpillar-2);--sunpillar-clr-3:var(--sunpillar-3);--sunpillar-clr-4:var(--sunpillar-4);--sunpillar-clr-5:var(--sunpillar-5);--sunpillar-clr-6:var(--sunpillar-6);--card-radius:30px; }
.pc-card-wrapper::before { content:''; position:absolute; inset:-10px; background:inherit; background-position:inherit; border-radius:inherit; transition:all .5s ease; filter:contrast(2) saturate(2) blur(36px); transform:scale(.8) translate3d(0,0,0.1px); background-size:100% 100%; background-image:var(--behind-gradient); }
.pc-card-wrapper:hover,.pc-card-wrapper.active { --card-opacity:1; }
.pc-card-wrapper:hover::before,.pc-card-wrapper.active::before { filter:contrast(1) saturate(2) blur(40px) opacity(1); transform:scale(.9) translate3d(0,0,0.1px); }
.pc-card { height:320px; display:grid; aspect-ratio:0.718; border-radius:var(--card-radius); position:relative; background-blend-mode:color-dodge,normal,normal,normal; animation:glow-bg 12s linear infinite; box-shadow:rgba(0,0,0,.8) calc((var(--pointer-from-left)*10px)-3px) calc((var(--pointer-from-top)*20px)-6px) 20px -5px; transition:transform 1s ease; transform:translate3d(0,0,0.1px) rotateX(0deg) rotateY(0deg); background-size:100% 100%; background-position:0 0,0 0,50% 50%,0 0; background-image:radial-gradient(farthest-side circle at var(--pointer-x) var(--pointer-y),hsla(266,100%,90%,var(--card-opacity)) 4%,hsla(266,50%,80%,calc(var(--card-opacity)*0.75)) 10%,hsla(266,25%,70%,calc(var(--card-opacity)*0.5)) 50%,hsla(266,0%,60%,0) 100%),radial-gradient(35% 52% at 55% 20%,#00ffaac4 0%,#073aff00 100%),radial-gradient(100% 100% at 50% 50%,#00c1ffff 1%,#073aff00 76%),conic-gradient(from 124deg at 50% 50%,#c137ffff 0%,#07c6ffff 40%,#07c6ffff 60%,#c137ffff 100%); overflow:hidden; }
.pc-card:hover,.pc-card.active { transition:none; transform:translate3d(0,0,0.1px) rotateX(var(--rotate-y)) rotateY(var(--rotate-x)); }
.pc-card * { display:grid; grid-area:1/-1; border-radius:var(--card-radius); transform:translate3d(0,0,0.1px); pointer-events:none; }
.pc-inside { inset:1px; position:absolute; background-image:var(--inner-gradient); background-color:rgba(0,0,0,0.9); transform:translate3d(0,0,0.01px); }
.pc-shine { mask-image:var(--icon); mask-mode:luminance; mask-repeat:repeat; mask-size:150%; mask-position:top calc(200% - (var(--background-y) * 5)) left calc(100% - var(--background-x)); -webkit-mask-image:var(--icon); -webkit-mask-mode:luminance; -webkit-mask-repeat:repeat; -webkit-mask-size:150%; -webkit-mask-position:top calc(200% - (var(--background-y) * 5)) left calc(100% - var(--background-x)); transition:filter .6s ease; filter:brightness(.66) contrast(1.33) saturate(.33) opacity(.5); animation:holo-bg 18s linear infinite; mix-blend-mode:color-dodge; }
.pc-glare { transform:translate3d(0,0,1.1px); overflow:hidden; background-image:radial-gradient(farthest-corner circle at var(--pointer-x) var(--pointer-y),hsl(248,25%,80%) 12%,hsla(207,40%,30%,0.8) 90%); mix-blend-mode:overlay; filter:brightness(.8) contrast(1.2); z-index:4; }
.pc-avatar-content { mix-blend-mode:screen; overflow:hidden; }
.pc-avatar-content .avatar { width:100%; position:absolute; left:50%; transform:translateX(-50%) scale(1); bottom:2px; opacity:calc(1.75 - var(--pointer-from-center)); }
.pc-avatar-content::before { content:''; position:absolute; inset:0; z-index:1; backdrop-filter:blur(30px); mask:linear-gradient(to bottom,rgba(0,0,0,0) 0%,rgba(0,0,0,0) 60%,rgba(0,0,0,1) 90%,rgba(0,0,0,1) 100%); pointer-events:none; }
.pc-content { max-height:100%; overflow:hidden; text-align:center; position:relative; transform:translate3d(calc(var(--pointer-from-left)*-6px + 3px),calc(var(--pointer-from-top)*-6px + 3px),0.1px) !important; z-index:5; mix-blend-mode:luminosity; }
.pc-details { width:100%; position:absolute; top:var(--card-text-top,3em); display:flex; flex-direction:column; align-items:center; }
.pc-details h3 { font-weight:600; margin:0; font-size:1.6rem; background-image:linear-gradient(to bottom,#fff,#6f6fbe); background-size:1em 1.5em; -webkit-text-fill-color:transparent; background-clip:text; -webkit-background-clip:text; }
.pc-details p { font-weight:600; position:relative; margin: var(--card-title-gap,6px) 0 0; white-space:nowrap; font-size:14px; width:min-content; background-image:linear-gradient(to bottom,#fff,#4a4ac0); background-size:1em 1.5em; -webkit-text-fill-color:transparent; background-clip:text; -webkit-background-clip:text; }
@keyframes glow-bg { 0% { --bgrotate:0deg;} 100% { --bgrotate:360deg;} }
@keyframes holo-bg { 0% { background-position:0 var(--background-y),0 0,center;} 100% { background-position:0 var(--background-y),90% 90%,center;} }
</style>
