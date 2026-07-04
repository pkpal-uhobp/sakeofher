<template>
  <main class="landing-page">
    <div class="noise"></div>
    <div class="glow glow-one"></div>
    <div class="glow glow-two"></div>

    <header class="site-header">
      <a class="brand" href="#top" aria-label="SakeOfHer">
        <span class="brand-mark"></span>
        <span>SakeOfHer</span>
      </a>

      <nav class="nav" aria-label="Основная навигация">
        <a href="#top" class="active">Главная</a>
        <a href="#about">О проекте</a>
        <a href="#features">Возможности</a>
        <a href="#structure">Структура</a>
        <a href="#contacts">Контакты</a>
      </nav>

      <RouterLink class="login-button" to="/login">
        <span class="login-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none">
            <path d="M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Z" stroke="currentColor" stroke-width="1.8" />
            <path d="M4.75 20c.75-3.45 3.52-5.6 7.25-5.6s6.5 2.15 7.25 5.6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
          </svg>
        </span>
        Войти
      </RouterLink>
    </header>

    <section id="top" class="hero">
      <div class="hero-content">
        <p class="eyebrow">Информационная среда нового поколения</p>

        <h1>
          Современное цифровое пространство для идей и управления
        </h1>

        <p class="hero-text">
          SakeOfHer — это удобная информационная среда для организации процессов,
          материалов и коммуникаций. Всё важное — в едином пространстве,
          структурно и наглядно.
        </p>

        <div class="hero-actions">
          <a class="primary-action" href="#about">
            Узнать больше
            <span aria-hidden="true">→</span>
          </a>

          <a class="secondary-action" href="#structure">
            Смотреть структуру
            <span class="mini-grid" aria-hidden="true"></span>
          </a>
        </div>
      </div>

      <div class="hero-visual" aria-hidden="true">
        <div class="network-grid"></div>

        <div class="floating-card chart-card">
          <span>динамика</span>
          <div class="bars">
            <i style="height: 34%"></i>
            <i style="height: 58%"></i>
            <i style="height: 46%"></i>
            <i style="height: 78%"></i>
          </div>
        </div>

        <div class="floating-card folder-card">
          <svg viewBox="0 0 64 64" fill="none">
            <path d="M9 19.5c0-4.14 3.36-7.5 7.5-7.5h11.2c2.25 0 4.37 1.02 5.78 2.76L37 19h10.5c4.14 0 7.5 3.36 7.5 7.5v18c0 4.14-3.36 7.5-7.5 7.5h-31C12.36 52 9 48.64 9 44.5v-25Z" fill="currentColor" opacity=".92"/>
          </svg>
        </div>

        <div class="floating-card profile-card">
          <span class="avatar"></span>
          <span class="profile-line big"></span>
          <span class="profile-line"></span>
          <span class="profile-line short"></span>
        </div>

        <div class="floating-card flow-card">
          <span class="node"></span>
          <span class="line"></span>
          <span class="node"></span>
          <span class="line down"></span>
          <span class="node"></span>
        </div>

        <div class="hub">
          <div class="cube">
            <span class="face front"></span>
            <span class="face top"></span>
            <span class="face side"></span>
          </div>
          <div class="platform platform-one"></div>
          <div class="platform platform-two"></div>
          <div class="platform platform-three"></div>
        </div>
      </div>
    </section>

    <section id="features" class="feature-grid" aria-label="Возможности">
      <article v-for="feature in features" :key="feature.title" class="feature-card">
        <div class="icon-box" v-html="feature.icon"></div>
        <div>
          <h2>{{ feature.title }}</h2>
          <p>{{ feature.text }}</p>
        </div>
      </article>
    </section>

    <section id="about" class="info-grid" aria-label="О проекте">
      <article class="info-card info-card-large">
        <div class="icon-box">
          <svg viewBox="0 0 24 24" fill="none">
            <path d="M4 20h16M6 18V9.5L12 5l6 4.5V18M9 18v-6h6v6" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h2>О проекте</h2>
        <p>
          SakeOfHer объединяет инструменты и подходы для системной работы с информацией.
          Проект помогает упорядочить сложные процессы и сделать их понятными для всех участников.
        </p>
        <span class="card-accent"></span>
      </article>

      <article class="info-card">
        <div class="icon-box">
          <svg viewBox="0 0 24 24" fill="none">
            <path d="M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Z" stroke="currentColor" stroke-width="1.7"/>
            <path d="m15.5 8.5-2.25 5.25L8 16l2.25-5.25L15.5 8.5Z" stroke="currentColor" stroke-width="1.7" stroke-linejoin="round"/>
          </svg>
        </div>
        <h2>Основные направления</h2>
        <p>
          Мы развиваем инструменты для управления знаниями, проектами и коммуникациями.
          Акцент на прозрачности, структурности и полезности каждого элемента.
        </p>
        <span class="card-accent"></span>
      </article>

      <article class="info-card">
        <div class="icon-box">
          <svg viewBox="0 0 24 24" fill="none">
            <path d="m12 3 8 4.5-8 4.5-8-4.5L12 3Z" stroke="currentColor" stroke-width="1.7" stroke-linejoin="round"/>
            <path d="m4 12 8 4.5 8-4.5M4 16.5 12 21l8-4.5" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h2>Подход</h2>
        <p>
          Мы строим среду, где важные детали не теряются, а процессы становятся
          логичными и управляемыми. Простота, качество и устойчивое развитие.
        </p>
        <span class="card-accent"></span>
      </article>
    </section>

    <section id="structure" class="steps-section">
      <div class="section-heading">
        <p>Логика работы</p>
        <h2>Как это устроено</h2>
      </div>

      <div class="steps">
        <article v-for="step in steps" :key="step.number" class="step-card">
          <span class="step-number">{{ step.number }}</span>
          <div>
            <h3>{{ step.title }}</h3>
            <p>{{ step.text }}</p>
          </div>
        </article>
      </div>
    </section>

    <footer id="contacts" class="footer">
      <a class="footer-brand" href="#top">SakeOfHer</a>
      <nav aria-label="Нижняя навигация">
        <a href="#about">Политика</a>
        <a href="#features">Поддержка</a>
        <a href="#contacts">Контакты</a>
      </nav>
    </footer>
  </main>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'

const features = [
  {
    title: 'Единая структура',
    text: 'Все материалы и процессы собраны в одном пространстве.',
    icon: `<svg viewBox="0 0 24 24" fill="none"><path d="m12 3 8 4.5-8 4.5-8-4.5L12 3Z" stroke="currentColor" stroke-width="1.7" stroke-linejoin="round"/><path d="m4 12 8 4.5 8-4.5M4 16.5 12 21l8-4.5" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
  },
  {
    title: 'Наглядная аналитика',
    text: 'Ключевая информация представлена понятно и аккуратно.',
    icon: `<svg viewBox="0 0 24 24" fill="none"><path d="M12 21a9 9 0 1 0-9-9h9v9Z" stroke="currentColor" stroke-width="1.7" stroke-linejoin="round"/><path d="M15 3.5A9 9 0 0 1 20.5 9H15V3.5Z" stroke="currentColor" stroke-width="1.7" stroke-linejoin="round"/></svg>`,
  },
  {
    title: 'Гибкая организация',
    text: 'Удобно выстраивать разделы, сценарии и связи.',
    icon: `<svg viewBox="0 0 24 24" fill="none"><path d="M6 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM18 14a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM6 22a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z" stroke="currentColor" stroke-width="1.7"/><path d="m8.6 6.7 6.8 3.6M8.5 17.2l7-4.4" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/></svg>`,
  },
  {
    title: 'Командная работа',
    text: 'Прозрачное взаимодействие и общий контекст.',
    icon: `<svg viewBox="0 0 24 24" fill="none"><path d="M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8ZM17.5 10a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM3 21c.65-4.2 2.85-6.5 6-6.5s5.35 2.3 6 6.5M14.5 14.5c2.62.2 4.38 2.4 4.9 6.5" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/></svg>`,
  },
]

const steps = [
  {
    number: '1',
    title: 'Сбор информации',
    text: 'Фиксируем идеи, материалы и данные из разных источников в единое пространство.',
  },
  {
    number: '2',
    title: 'Организация структуры',
    text: 'Создаём логичную систему разделов, связей и сценариев для удобной работы.',
  },
  {
    number: '3',
    title: 'Работа и развитие',
    text: 'Используем структуру для достижения целей и постоянного улучшения процессов.',
  },
]
</script>

<style scoped>
:global(*) {
  box-sizing: border-box;
}

.landing-page {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  color: #f8fafc;
  background:
    radial-gradient(circle at 73% 21%, rgba(0, 126, 255, 0.24), transparent 31%),
    radial-gradient(circle at 12% 8%, rgba(2, 132, 199, 0.15), transparent 28%),
    linear-gradient(135deg, #03070f 0%, #070b12 42%, #020409 100%);
  font-family:
    Inter,
    ui-sans-serif,
    system-ui,
    -apple-system,
    BlinkMacSystemFont,
    "Segoe UI",
    sans-serif;
}

.noise {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  opacity: 0.16;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.035) 1px, transparent 1px);
  background-size: 56px 56px;
  mask-image: radial-gradient(circle at 50% 20%, black, transparent 72%);
}

.glow {
  position: fixed;
  z-index: 0;
  width: 420px;
  height: 420px;
  border-radius: 999px;
  background: rgba(0, 145, 255, 0.22);
  filter: blur(90px);
  pointer-events: none;
}

.glow-one {
  top: -210px;
  right: 16%;
}

.glow-two {
  bottom: -260px;
  left: 10%;
  background: rgba(14, 165, 233, 0.13);
}

.site-header,
.hero,
.feature-grid,
.info-grid,
.steps-section,
.footer {
  position: relative;
  z-index: 1;
  width: min(1440px, calc(100% - 64px));
  margin-inline: auto;
}

.site-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 76px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.13);
}

.brand,
.footer-brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  color: #ffffff;
  font-size: 28px;
  font-weight: 800;
  letter-spacing: -0.04em;
  text-decoration: none;
}

.brand-mark {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #0ea5e9;
  box-shadow:
    0 0 18px rgba(14, 165, 233, 0.95),
    0 0 40px rgba(14, 165, 233, 0.45);
}

.nav {
  display: flex;
  align-items: center;
  gap: 44px;
}

.nav a {
  position: relative;
  color: rgba(226, 232, 240, 0.86);
  font-size: 16px;
  text-decoration: none;
  transition: color 0.2s ease;
}

.nav a:hover,
.nav a.active {
  color: #ffffff;
}

.nav a.active::after {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  bottom: -17px;
  height: 3px;
  border-radius: 999px;
  background: #0284ff;
  box-shadow: 0 0 18px rgba(2, 132, 255, 0.9);
}

.login-button {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  height: 42px;
  padding-inline: 20px;
  border: 1px solid rgba(14, 165, 233, 0.75);
  border-radius: 999px;
  color: #f8fafc;
  font-weight: 700;
  text-decoration: none;
  background: rgba(2, 6, 23, 0.55);
  box-shadow: inset 0 0 18px rgba(14, 165, 233, 0.08);
  transition:
    transform 0.2s ease,
    background 0.2s ease,
    border-color 0.2s ease;
}

.login-button:hover {
  transform: translateY(-1px);
  border-color: rgba(125, 211, 252, 0.95);
  background: rgba(14, 165, 233, 0.12);
}

.login-icon svg {
  display: block;
  width: 20px;
  height: 20px;
}

.hero {
  display: grid;
  grid-template-columns: minmax(0, 0.88fr) minmax(460px, 1.12fr);
  gap: 52px;
  align-items: center;
  min-height: 490px;
  padding: 56px 0 34px;
}

.eyebrow {
  margin: 0 0 20px;
  color: #7dd3fc;
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.hero h1 {
  max-width: 760px;
  margin: 0;
  color: #ffffff;
  font-size: clamp(46px, 4.7vw, 78px);
  line-height: 0.96;
  letter-spacing: -0.075em;
}

.hero-text {
  max-width: 620px;
  margin: 26px 0 0;
  color: rgba(203, 213, 225, 0.9);
  font-size: 20px;
  line-height: 1.45;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 22px;
  margin-top: 34px;
}

.primary-action,
.secondary-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  min-width: 220px;
  min-height: 56px;
  padding: 15px 24px;
  border-radius: 16px;
  font-size: 17px;
  font-weight: 800;
  text-decoration: none;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
}

.primary-action {
  color: #ffffff;
  background: linear-gradient(135deg, #0786ff, #0057e5);
  box-shadow:
    0 16px 42px rgba(0, 102, 255, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.22);
}

.secondary-action {
  color: #f8fafc;
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(2, 6, 23, 0.42);
}

.primary-action:hover,
.secondary-action:hover {
  transform: translateY(-2px);
}

.primary-action:hover {
  box-shadow:
    0 22px 54px rgba(0, 102, 255, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.24);
}

.secondary-action:hover {
  border-color: rgba(125, 211, 252, 0.62);
}

.mini-grid {
  width: 18px;
  height: 18px;
  border: 1px solid rgba(226, 232, 240, 0.8);
  border-radius: 4px;
  background:
    linear-gradient(rgba(226, 232, 240, 0.8), rgba(226, 232, 240, 0.8)) 50% 0 / 1px 100% no-repeat,
    linear-gradient(90deg, rgba(226, 232, 240, 0.8), rgba(226, 232, 240, 0.8)) 0 50% / 100% 1px no-repeat;
}

.hero-visual {
  position: relative;
  min-height: 430px;
  isolation: isolate;
}

.network-grid {
  position: absolute;
  inset: 50px 0 0;
  border-radius: 28px;
  opacity: 0.75;
  background:
    radial-gradient(circle at 50% 50%, rgba(56, 189, 248, 0.26), transparent 42%),
    linear-gradient(rgba(14, 165, 233, 0.16) 1px, transparent 1px),
    linear-gradient(90deg, rgba(14, 165, 233, 0.14) 1px, transparent 1px);
  background-size: 100% 100%, 42px 42px, 42px 42px;
  transform: perspective(700px) rotateX(58deg) rotateZ(-6deg);
  transform-origin: 50% 65%;
  filter: drop-shadow(0 0 28px rgba(14, 165, 233, 0.22));
}

.hub {
  position: absolute;
  left: 50%;
  top: 47%;
  width: 310px;
  height: 260px;
  transform: translate(-50%, -50%);
}

.platform {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  border: 1px solid rgba(56, 189, 248, 0.4);
  border-radius: 28px;
  background: linear-gradient(180deg, rgba(14, 165, 233, 0.32), rgba(2, 6, 23, 0.82));
  box-shadow:
    0 0 36px rgba(14, 165, 233, 0.36),
    inset 0 1px 0 rgba(255, 255, 255, 0.14);
}

.platform-one {
  bottom: 42px;
  width: 280px;
  height: 70px;
}

.platform-two {
  bottom: 88px;
  width: 224px;
  height: 58px;
}

.platform-three {
  bottom: 128px;
  width: 154px;
  height: 42px;
}

.cube {
  position: absolute;
  left: 50%;
  top: 35px;
  width: 124px;
  height: 124px;
  transform: translateX(-50%) rotateX(-22deg) rotateY(31deg);
  transform-style: preserve-3d;
  filter: drop-shadow(0 0 34px rgba(56, 189, 248, 0.75));
}

.face {
  position: absolute;
  inset: 0;
  border: 1px solid rgba(186, 230, 253, 0.75);
  background: linear-gradient(135deg, rgba(56, 189, 248, 0.86), rgba(2, 132, 199, 0.4));
  box-shadow: inset 0 0 28px rgba(255, 255, 255, 0.16);
}

.front {
  transform: translateZ(62px);
}

.top {
  transform: rotateX(90deg) translateZ(62px);
  background: linear-gradient(135deg, rgba(125, 211, 252, 0.9), rgba(14, 165, 233, 0.34));
}

.side {
  transform: rotateY(90deg) translateZ(62px);
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.72), rgba(7, 89, 133, 0.52));
}

.floating-card {
  position: absolute;
  z-index: 4;
  border: 1px solid rgba(125, 211, 252, 0.28);
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.88), rgba(2, 6, 23, 0.58));
  box-shadow:
    0 20px 50px rgba(0, 0, 0, 0.28),
    0 0 32px rgba(14, 165, 233, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(14px);
}

.chart-card {
  left: 4%;
  top: 22%;
  width: 172px;
  height: 112px;
  padding: 18px;
}

.chart-card span {
  display: block;
  margin-bottom: 18px;
  color: rgba(203, 213, 225, 0.78);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.bars {
  display: flex;
  align-items: end;
  gap: 10px;
  height: 48px;
}

.bars i {
  display: block;
  flex: 1;
  border-radius: 999px 999px 4px 4px;
  background: linear-gradient(180deg, #7dd3fc, #0066ff);
  box-shadow: 0 0 16px rgba(56, 189, 248, 0.44);
}

.folder-card {
  left: 18%;
  top: 4%;
  width: 88px;
  height: 80px;
  display: grid;
  place-items: center;
  color: #38bdf8;
}

.folder-card svg {
  width: 56px;
  height: 56px;
}

.profile-card {
  right: 4%;
  top: 37%;
  width: 172px;
  height: 116px;
  padding: 18px;
}

.avatar {
  display: block;
  width: 36px;
  height: 36px;
  margin-bottom: 14px;
  border-radius: 50%;
  background: linear-gradient(135deg, #7dd3fc, #2563eb);
  box-shadow: 0 0 18px rgba(56, 189, 248, 0.45);
}

.profile-line {
  display: block;
  width: 72%;
  height: 8px;
  margin-top: 8px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.34);
}

.profile-line.big {
  width: 82%;
}

.profile-line.short {
  width: 46%;
}

.flow-card {
  right: 1%;
  top: 6%;
  width: 220px;
  height: 118px;
  padding: 22px;
}

.flow-card .node {
  display: inline-block;
  width: 38px;
  height: 22px;
  border: 1px solid rgba(125, 211, 252, 0.68);
  border-radius: 8px;
  background: rgba(14, 165, 233, 0.18);
}

.flow-card .line {
  display: inline-block;
  width: 36px;
  height: 1px;
  margin: 0 6px 7px;
  background: rgba(125, 211, 252, 0.72);
}

.flow-card .line.down {
  width: 1px;
  height: 28px;
  margin: 6px 18px -2px;
}

.feature-grid,
.info-grid {
  display: grid;
  gap: 18px;
}

.feature-grid {
  grid-template-columns: repeat(4, 1fr);
  padding: 20px 0 24px;
}

.feature-card,
.info-card,
.step-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 22px;
  background:
    linear-gradient(180deg, rgba(15, 23, 42, 0.72), rgba(2, 6, 23, 0.62)),
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.13), transparent 46%);
  box-shadow:
    0 22px 70px rgba(0, 0, 0, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.055);
}

.feature-card {
  display: flex;
  align-items: flex-start;
  gap: 18px;
  min-height: 132px;
  padding: 22px;
}

.icon-box {
  flex: 0 0 auto;
  display: grid;
  place-items: center;
  width: 58px;
  height: 58px;
  border: 1px solid rgba(125, 211, 252, 0.22);
  border-radius: 18px;
  color: #0ea5e9;
  background: radial-gradient(circle at 40% 30%, rgba(56, 189, 248, 0.32), rgba(15, 23, 42, 0.72));
  box-shadow: inset 0 0 22px rgba(14, 165, 233, 0.12);
}

.icon-box :deep(svg),
.info-card .icon-box svg {
  width: 32px;
  height: 32px;
}

.feature-card h2,
.info-card h2,
.step-card h3 {
  margin: 0;
  color: #ffffff;
  letter-spacing: -0.03em;
}

.feature-card h2 {
  font-size: 19px;
}

.feature-card p,
.info-card p,
.step-card p {
  color: rgba(203, 213, 225, 0.84);
  line-height: 1.48;
}

.feature-card p {
  margin: 9px 0 0;
  font-size: 15px;
}

.info-grid {
  grid-template-columns: 1fr 1fr 1fr;
  padding: 0 0 24px;
}

.info-card {
  min-height: 188px;
  padding: 28px;
}

.info-card-large {
  min-height: 188px;
}

.info-card .icon-box {
  margin-bottom: 18px;
}

.info-card h2 {
  font-size: 22px;
}

.info-card p {
  max-width: 520px;
  margin: 12px 0 0;
  font-size: 16px;
}

.card-accent {
  display: block;
  width: 84px;
  height: 3px;
  margin-top: 20px;
  border-radius: 999px;
  background: linear-gradient(90deg, #0ea5e9, transparent);
  box-shadow: 0 0 14px rgba(14, 165, 233, 0.55);
}

.steps-section {
  padding: 4px 0 34px;
}

.section-heading {
  text-align: center;
  margin-bottom: 16px;
}

.section-heading p {
  margin: 0 0 6px;
  color: #7dd3fc;
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.section-heading h2 {
  margin: 0;
  color: #ffffff;
  font-size: 32px;
  letter-spacing: -0.05em;
}

.steps {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 28px;
  position: relative;
}

.steps::before {
  content: "";
  position: absolute;
  top: 50%;
  left: 18%;
  right: 18%;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(14, 165, 233, 0.86), transparent);
  box-shadow: 0 0 20px rgba(14, 165, 233, 0.65);
}

.step-card {
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 20px;
  min-height: 108px;
  padding: 20px 24px;
}

.step-number {
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  width: 62px;
  height: 62px;
  border: 1px solid rgba(14, 165, 233, 0.64);
  border-radius: 50%;
  color: #dbeafe;
  font-size: 31px;
  font-weight: 900;
  background:
    radial-gradient(circle at 40% 30%, rgba(14, 165, 233, 0.35), rgba(15, 23, 42, 0.66)),
    #020617;
  box-shadow:
    0 0 22px rgba(14, 165, 233, 0.38),
    inset 0 0 18px rgba(14, 165, 233, 0.12);
}

.step-card h3 {
  font-size: 18px;
}

.step-card p {
  margin: 8px 0 0;
  font-size: 14px;
}

.footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 76px;
  border-top: 1px solid rgba(148, 163, 184, 0.12);
}

.footer-brand {
  font-size: 27px;
}

.footer nav {
  display: flex;
  align-items: center;
  gap: 46px;
}

.footer a {
  color: rgba(226, 232, 240, 0.72);
  text-decoration: none;
}

.footer a:hover {
  color: #ffffff;
}

@media (max-width: 1180px) {
  .site-header,
  .hero,
  .feature-grid,
  .info-grid,
  .steps-section,
  .footer {
    width: min(100% - 36px, 960px);
  }

  .nav {
    display: none;
  }

  .hero {
    grid-template-columns: 1fr;
    min-height: auto;
  }

  .hero-visual {
    min-height: 360px;
  }

  .feature-grid,
  .info-grid,
  .steps {
    grid-template-columns: 1fr 1fr;
  }

  .steps::before {
    display: none;
  }
}

@media (max-width: 720px) {
  .brand {
    font-size: 22px;
  }

  .login-button {
    height: 38px;
    padding-inline: 14px;
  }

  .hero {
    padding-top: 34px;
  }

  .hero h1 {
    font-size: 42px;
  }

  .hero-text {
    font-size: 17px;
  }

  .hero-actions {
    flex-direction: column;
  }

  .primary-action,
  .secondary-action {
    width: 100%;
  }

  .hero-visual {
    min-height: 300px;
  }

  .floating-card {
    display: none;
  }

  .feature-grid,
  .info-grid,
  .steps {
    grid-template-columns: 1fr;
  }

  .footer {
    flex-direction: column;
    gap: 18px;
    padding: 24px 0;
  }

  .footer nav {
    gap: 22px;
  }
}
</style>
