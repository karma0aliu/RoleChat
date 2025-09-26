<template>
    <div class="login-body">
        <div class="main">
        <div class="container a-container" id="a-container" ref="aContainer">
            <form class="form" id="a-form" @submit.prevent="handleSignUp">
            <h2 class="form_title title">创建账号</h2>
            <div class="form__icons">
                </div>
            <span class="form__span">使用邮箱注册</span>
            <input class="form__input" type="text" placeholder="Name" v-model="signUpName" required>
            <input class="form__input" type="email" placeholder="Email" v-model="signUpEmail" required>
            <input class="form__input" type="password" placeholder="Password (>=6 chars)" v-model="signUpPassword" required>
            <div v-if="signUpError" class="form__error">{{ signUpError }}</div>
            <button class="form__button button submit" :disabled="signUpLoading">{{ signUpLoading ? '注册中...' : 'SIGN UP' }}</button>
            </form>
        </div>

        <div class="container b-container" id="b-container" ref="bContainer">
            <form class="form" id="b-form" @submit.prevent="handleSignIn">
            <h2 class="form_title title">登入账号</h2>
            <div class="form__icons">
                </div>
            <span class="form__span">使用您的邮箱和密码登录</span>
            <input class="form__input" type="email" placeholder="Email" v-model="signInEmail" required>
            <input class="form__input" type="password" placeholder="Password" v-model="signInPassword" required>
            <div v-if="errorMessage" class="form__error">
                {{ errorMessage }}
            </div>

            <button class="form__button button submit" :disabled="isLoading">
                {{ isLoading ? '登录中...' : 'SIGN IN' }}
            </button>
            </form>
        </div>

        <div class="switch" id="switch-cnt" ref="switchCtn">
            <div class="switch__circle" ref="switchCircle1"></div>
            <div class="switch__circle switch__circle--t" ref="switchCircle2"></div>
            <div class="switch__container" id="switch-c1" ref="switchC1">
            <h2 class="switch__title title">欢迎回来！</h2>
            <button class="switch__button button switch-btn" @click="changeForm">SIGN IN</button>
            </div>
            <div class="switch__container is-hidden" id="switch-c2" ref="switchC2">
            <h2 class="switch__title title">你好, 用户！</h2>
            <button class="switch__button button switch-btn" @click="changeForm">SIGN UP</button>
            </div>
        </div>
        </div>
    </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { setAuth } from '../auth';

const router = useRouter();
const route = useRoute();

const signUpName = ref('');
const signUpEmail = ref('');
const signUpPassword = ref('');
const signUpLoading = ref(false);
const signUpError = ref('');
const signInEmail = ref('');
const signInPassword = ref('');
const isLoading = ref(false);
const errorMessage = ref('');

const switchCtn = ref(null);
const switchC1 = ref(null);
const switchC2 = ref(null);
const switchCircle1 = ref(null);
const switchCircle2 = ref(null);
const aContainer = ref(null);
const bContainer = ref(null);

const changeForm = () => {
switchCtn.value.classList.add("is-gx");
setTimeout(() => {
    switchCtn.value.classList.remove("is-gx");
}, 1500);

switchCtn.value.classList.toggle("is-txr");
switchCircle1.value.classList.toggle("is-txr");
switchCircle2.value.classList.toggle("is-txr");

switchC1.value.classList.toggle("is-hidden");
switchC2.value.classList.toggle("is-hidden");
aContainer.value.classList.toggle("is-txl");
bContainer.value.classList.toggle("is-txl");
bContainer.value.classList.toggle("is-z200");
};

const API_BASE = 'http://127.0.0.1:8080/api';

const handleSignUp = async () => {
    signUpError.value = '';
    if (!signUpName.value.trim()) { signUpError.value = '请输入名称'; return; }
    if (!/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(signUpEmail.value)) { signUpError.value = '邮箱格式不正确'; return; }
    if (signUpPassword.value.length < 6) { signUpError.value = '密码至少 6 位'; return; }
    signUpLoading.value = true;
    try {
        const REGISTER_ENDPOINT = API_BASE + '/auth/register';
        const payload = {
            username: signUpName.value,
            email: signUpEmail.value,
            password: signUpPassword.value,
            nickname: signUpName.value
        };
        const res = await fetch(REGISTER_ENDPOINT, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
            throw new Error(data.detail || data.message || '注册失败');
        }
        let access = data.access_token || data.access || data.token || '';
        let refresh = data.refresh_token || data.refresh || '';
        let user = data.user;

        if (!access) {
            const loginRes = await fetch(API_BASE + '/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email: signUpEmail.value, password: signUpPassword.value })
            });
            const loginData = await loginRes.json();
            if (!loginRes.ok) throw new Error(loginData.detail || '注册成功但自动登录失败');
            access = loginData.access_token;
            refresh = loginData.refresh_token;
            user = loginData.user;
        }
        if (!access) throw new Error('没有收到访问令牌');
                if (refresh) localStorage.setItem('refreshToken', refresh);
                // 构造最终用户对象：优先后端返回，其次使用注册表单中的名称
                const finalUser = user
                    ? {
                            username: user.username || user.name || signUpName.value,
                            email: user.email || signUpEmail.value,
                            nickname: user.nickname || user.display_name || user.name || signUpName.value
                        }
                    : { username: signUpName.value, email: signUpEmail.value, nickname: signUpName.value };
                setAuth(access, finalUser);
        const redirect = '/app/new';
        router.push(redirect);
    } catch (e) {
        signUpError.value = e.message || '发生未知错误';
    } finally {
        signUpLoading.value = false;
    }
};

const handleSignIn = async () => {
    isLoading.value = true;
    errorMessage.value = '';

    try {
    const apiUrl = API_BASE + '/auth/login';

        const payload = {
            email: signInEmail.value,
            password: signInPassword.value
        };

        const response = await fetch(apiUrl, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.detail || data.message || '登录失败，请检查您的凭据。');
        }


        const accessToken = data.access_token;
        const refreshToken = data.refresh_token;
        if (!accessToken) throw new Error('服务器返回的数据中不包含 access_token。');

                if (refreshToken) localStorage.setItem('refreshToken', refreshToken);
                        const loginUser = data.user
                            ? {
                                    username: data.user.username || data.user.name || signInEmail.value,
                                    email: data.user.email || signInEmail.value,
                                    nickname: data.user.nickname || data.user.display_name || data.user.name || data.user.username || signInEmail.value.split('@')[0]
                                }
                            : { username: signInEmail.value, email: signInEmail.value, nickname: signInEmail.value.split('@')[0] };
                setAuth(accessToken, loginUser);

    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/app/new';
        router.push(redirect);

    } catch (error) {
        console.error("登录过程中发生错误:", error);
        errorMessage.value = error.message;
    } finally {
        isLoading.value = false;
    }
};
</script>

<style lang="scss" scoped>
$neu-1: #ecf0f3;
$neu-2: #d1d9e6;
$white: #f9f9f9;
$gray: #a0a5a8;
$black: #181818;
$purple: #4B70E2;
$danger: #e57373;
$transition: 1.25s;

*, *::after, *::before {
margin: 0;
padding: 0;
box-sizing: border-box;
user-select: none;
}

.login-body {
width: 100%;
height: 100%;
display: flex;
justify-content: center;
align-items: center;
font-family: 'Montserrat', sans-serif;
font-size: 12px;
color: $gray;
}

.main {
position: relative;
width: 1000px;
min-width: 1000px;
min-height: 600px;
height: 600px;
padding: 25px;
background-color: $neu-1;
box-shadow: 10px 10px 10px $neu-2, -10px -10px 10px $white;
border-radius: 12px;
overflow: hidden;

@media (max-width: 1200px) { transform: scale(0.7); }
@media (max-width: 1000px) { transform: scale(0.6); }
@media (max-width: 800px) { transform: scale(0.5); }
@media (max-width: 600px) { transform: scale(0.4); }
}

.container {
display: flex;
justify-content: center;
align-items: center;
position: absolute;
top: 0;
width: 600px;
height: 100%;
padding: 25px;
background-color: $neu-1;
transition: $transition;
}

.a-container {
z-index: 100;
left: calc(100% - 600px);
}

.b-container {
left: calc(100% - 600px);
z-index: 0;
}

.form {
display: flex;
justify-content: center;
align-items: center;
flex-direction: column;
width: 100%;
height: 100%;

&__input {
    width: 350px;
    height: 40px;
    margin: 4px 0;
    padding-left: 25px;
    font-size: 13px;
    letter-spacing: .15px;
    border: none;
    outline: none;
    font-family: 'Montserrat', sans-serif;
    background-color: $neu-1;
    transition: .25s ease;
    border-radius: 8px;
    box-shadow: inset 2px 2px 4px $neu-2, inset -2px -2px 4px $white;

    &:focus {
    box-shadow: inset 4px 4px 4px $neu-2, inset -4px -4px 4px $white;
    }
}

&__span {
    margin-top: 30px;
    margin-bottom: 12px;
}

&__error {
        color: $danger;
        font-size: 13px;
        font-weight: 500;
        margin-top: 15px;
        max-width: 350px;
        text-align: center;
}

&__link {
    color: $black;
    font-size: 15px;
    margin-top: 25px;
    border-bottom: 1px solid $gray;
    line-height: 2;
}
}

.title {
font-size: 34px;
font-weight: 700;
line-height: 3;
color: $black;
}

.description {
font-size: 14px;
letter-spacing: .25px;
text-align: center;
line-height: 1.6;
}

.button {
width: 180px;
height: 50px;
border-radius: 25px;
margin-top: 50px;
font-weight: 700;
font-size: 14px;
letter-spacing: 1.15px;
background-color: $purple;
color: $white;
box-shadow: 8px 8px 16px $neu-2, -8px -8px 16px $white;
border: none;
outline: none;
cursor: pointer;
}

.switch {
display: flex;
justify-content: center;
align-items: center;
position: absolute;
top: 0;
left: 0;
height: 100%;
width: 400px;
padding: 50px;
z-index: 200;
transition: $transition;
background-color: $neu-1;
overflow: hidden;
box-shadow: 4px 4px 10px $neu-2, -4px -4px 10px $white;

&__circle {
    position: absolute;
    width: 500px;
    height: 500px;
    border-radius: 50%;
    background-color: $neu-1;
    box-shadow: inset 8px 8px 12px $neu-2, inset -8px -8px 12px $white;
    bottom: -60%;
    left: -60%;
    transition: $transition;

    &--t {
    top: -30%;
    left: 60%;
    width: 300px;
    height: 300px;
    }
}

&__container {
    display: flex;
    justify-content: center;
    align-items: center;
    flex-direction: column;
    position: absolute;
    width: 400px;
    padding: 50px 55px;
    transition: $transition;
}

&__button {
    cursor: pointer;
    &:hover {
    box-shadow: 6px 6px 10px $neu-2, -6px -6px 10px $white;
    transform: scale(.985);
    transition: .25s;
    }
    &:active, &:focus {
    box-shadow: 2px 2px 6px $neu-2, -2px -2px 6px $white;
    transform: scale(.97);
    transition: .25s;
    }
}
}

.is-txr {
left: calc(100% - 400px);
transition: $transition;
transform-origin: left;
}

.is-txl {
left: 0;
transition: $transition;
transform-origin: right;
}

.is-z200 {
z-index: 200;
transition: $transition;
}

.is-hidden {
visibility: hidden;
opacity: 0;
position: absolute;
transition: $transition;
}

.is-gx {
animation: is-gx $transition;
}

@keyframes is-gx {
0%, 10%, 100% { width: 400px; }
30%, 50% { width: 500px; }
}

</style>