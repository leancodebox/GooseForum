<script setup lang="ts">
interface Achievement {
  number: string;
  label: string;
}

const achievements: Achievement[] = [
  { number: '5000+', label: '服务用户' },
  { number: '50+', label: '完成项目' },
  { number: '99%', label: '客户满意度' },
  { number: '24/7', label: '技术支持' }
];
</script>

<template>
  <div class="about">
    <div class="about-container">
      <!-- 介绍部分 -->
      <section class="about-intro">
        <h1>关于我们</h1>
        <p class="subtitle">用科技改变世界</p>
        <p class="description">我们是一支充满激情的技术团队，致力于为用户提供最优质的服务和产品体验。通过持续创新和技术突破，我们帮助企业实现数字化转型，创造更大的商业价值。</p>
      </section>

      <!-- 成就数据展示 -->
      <section class="achievements-section">
        <div class="achievements-grid">
          <div v-for="item in achievements" :key="item.label" class="achievement-item">
            <h3>{{ item.number }}</h3>
            <p>{{ item.label }}</p>
          </div>
        </div>
      </section>

      <!-- 我们的使命和愿景 -->
      <section class="mission-section">
        <div class="mission-grid">
          <div class="mission-item">
            <h2>我们的使命</h2>
            <p>通过技术创新推动行业发展，为客户创造价值，为员工提供发展平台。</p>
          </div>
          <div class="mission-item">
            <h2>我们的愿景</h2>
            <p>成为最受信赖的技术解决方案提供商，引领行业发展方向。</p>
          </div>
        </div>
      </section>

      <!-- 团队介绍 -->
      <section class="team-section">
        <h2>核心团队</h2>
        <div class="team-grid">
          <div class="team-member">
            <img src="https://api.dicebear.com/7.x/avataaars/svg?seed=John" alt="张三">
            <h3>张三</h3>
            <p class="role">技术总监</p>
            <p class="description">10年技术开发经验，专注于架构设计和技术创新</p>
          </div>
          <div class="team-member">
            <img src="https://api.dicebear.com/7.x/avataaars/svg?seed=Jane" alt="李四">
            <h3>李四</h3>
            <p class="role">产品经理</p>
            <p class="description">8年产品经验，擅长用户体验设计和产品规划</p>
          </div>
          <div class="team-member">
            <img src="https://api.dicebear.com/7.x/avataaars/svg?seed=Mike" alt="王五">
            <h3>王五</h3>
            <p class="role">研发主管</p>
            <p class="description">12年开发经验，全栈技术专家</p>
          </div>
        </div>
      </section>

      <!-- 技术栈 -->
      <section class="tech-stack">
        <h2>我们的技术栈</h2>
        <div class="tech-grid">
          <div class="tech-item">
            <img src="https://vuejs.org/images/logo.png" alt="Vue">
            <h3>Vue.js</h3>
            <p>现代化的渐进式JavaScript框架</p>
          </div>
          <div class="tech-item">
            <img src="https://www.typescriptlang.org/icons/icon-512x512.png" alt="TypeScript">
            <h3>TypeScript</h3>
            <p>类型安全的JavaScript超集</p>
          </div>
          <div class="tech-item">
            <img src="https://nodejs.org/static/images/logo.svg" alt="Node.js">
            <h3>Node.js</h3>
            <p>高性能服务端运行时</p>
          </div>
        </div>
      </section>

      <!-- 公司文化 -->
      <section class="culture-section">
        <h2>公司文化</h2>
        <div class="culture-grid">
          <div class="culture-item">
            <h3>创新</h3>
            <p>持续探索新技术，追求卓越</p>
          </div>
          <div class="culture-item">
            <h3>协作</h3>
            <p>团队合作，共同成长</p>
          </div>
          <div class="culture-item">
            <h3>诚信</h3>
            <p>诚信为本，客户至上</p>
          </div>
        </div>
      </section>

      <!-- 联系方式 -->
      <section class="contact-section">
        <h2>联系我们</h2>
        <div class="contact-grid">
          <div class="contact-info">
            <h3>联系方式</h3>
            <ul>
              <li><i class="fas fa-envelope"></i> support@leancode.com</li>
              <li><i class="fas fa-phone"></i> +86 123 4567 8901</li>
              <li><i class="fas fa-map-marker-alt"></i> 北京市朝阳区科技园区88号</li>
            </ul>
          </div>
          <div class="social-links">
            <h3>社交媒体</h3>
            <div class="social-icons">
              <a href="#" title="微信"><i class="fab fa-weixin"></i></a>
              <a href="#" title="微博"><i class="fab fa-weibo"></i></a>
              <a href="#" title="GitHub"><i class="fab fa-github"></i></a>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.about {
  padding: 2rem;
  background-color: var(--bg-color);
}

.about-container {
  max-width: 1200px;
  margin: 0 auto;
}

/* 介绍部分样式 */
.about-intro {
  text-align: center;
  margin-bottom: 4rem;
  padding: 2rem 0;
}

.about-intro h1 {
  font-size: 3rem;
  margin-bottom: 1rem;
  color: var(--color-heading);
}

.about-intro .subtitle {
  font-size: 1.5rem;
  color: var(--color-text);
  margin-bottom: 1.5rem;
}

.about-intro .description {
  font-size: 1.1rem;
  color: var(--color-text-light);
  max-width: 800px;
  margin: 0 auto;
  line-height: 1.6;
}

/* 成就展示样式 */
.achievements-section {
  margin: 4rem 0;
  padding: 3rem 0;
  background-color: var(--color-background-soft);
  border-radius: 12px;
}

.achievements-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 2rem;
  padding: 0 2rem;
}

.achievement-item {
  text-align: center;
}

.achievement-item h3 {
  font-size: 2.5rem;
  color: var(--color-primary);
  margin-bottom: 0.5rem;
}

/* 使命愿景样式 */
.mission-section {
  margin: 4rem 0;
}

.mission-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 3rem;
}

.mission-item {
  text-align: center;
  padding: 2rem;
  background-color: var(--color-background-soft);
  border-radius: 12px;
}

/* 团队样式更新 */
.team-section {
  margin: 4rem 0;
}

.team-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2.5rem;
}

.team-member {
  text-align: center;
  padding: 2rem;
  background: var(--color-background-soft);
  border-radius: 12px;
  transition: transform 0.3s ease;
}

.team-member:hover {
  transform: translateY(-5px);
}

.team-member img {
  width: 150px;
  height: 150px;
  border-radius: 50%;
  margin-bottom: 1.5rem;
  border: 4px solid var(--color-border);
}

.team-member .role {
  color: var(--color-primary);
  margin: 0.5rem 0;
}

.team-member .description {
  color: var(--color-text-light);
  font-size: 0.9rem;
}

/* 技术栈样式更新 */
.tech-stack {
  margin: 4rem 0;
}

.tech-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 2rem;
  padding: 1rem;
}

.tech-item {
  text-align: center;
  padding: 2rem;
  border-radius: 12px;
  transition: transform 0.3s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  position: relative;
  overflow: hidden;
}

.tech-item:nth-child(1) {
  background: radial-gradient(ellipse at center,
    rgba(65, 184, 131, 0.2) 0%,
    rgba(65, 184, 131, 0.15) 20%,
    rgba(65, 184, 131, 0.05) 50%,
    rgba(65, 184, 131, 0) 80%
  ),
  radial-gradient(ellipse at center,
    rgba(65, 184, 131, 0.1) 0%,
    rgba(65, 184, 131, 0) 70%
  );
}

.tech-item:nth-child(2) {
  background: radial-gradient(ellipse at center,
    rgba(49, 120, 198, 0.2) 0%,
    rgba(49, 120, 198, 0.15) 20%,
    rgba(49, 120, 198, 0.05) 50%,
    rgba(49, 120, 198, 0) 80%
  ),
  radial-gradient(ellipse at center,
    rgba(49, 120, 198, 0.1) 0%,
    rgba(49, 120, 198, 0) 70%
  );
}

.tech-item:nth-child(3) {
  background: radial-gradient(ellipse at center,
    rgba(104, 160, 99, 0.2) 0%,
    rgba(104, 160, 99, 0.15) 20%,
    rgba(104, 160, 99, 0.05) 50%,
    rgba(104, 160, 99, 0) 80%
  ),
  radial-gradient(ellipse at center,
    rgba(104, 160, 99, 0.1) 0%,
    rgba(104, 160, 99, 0) 70%
  );
}

.tech-item:hover {
  transform: translateY(-5px);
}

.tech-item img {
  width: 80px;
  height: 80px;
  object-fit: contain;
}

.tech-item h3 {
  margin: 0.5rem 0;
  color: var(--color-heading);
  font-weight: 600;
}

.tech-item p {
  color: var(--color-text);
  font-size: 0.9rem;
  line-height: 1.4;
}

/* 公司文化样式 */
.culture-section {
  margin: 4rem 0;
}

.culture-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 2rem;
}

.culture-item {
  text-align: center;
  padding: 2rem;
  background: var(--color-background-soft);
  border-radius: 12px;
}

/* 联系方式样式更新 */
.contact-section {
  margin: 4rem 0;
}

.contact-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 3rem;
}

.contact-info ul {
  list-style: none;
  padding: 0;
}

.contact-info li {
  margin: 1rem 0;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.social-icons {
  display: flex;
  gap: 1.5rem;
  justify-content: center;
  margin-top: 1rem;
}

.social-icons a {
  font-size: 1.5rem;
  color: var(--color-text);
  transition: color 0.3s ease;
}

.social-icons a:hover {
  color: var(--color-primary);
}

/* 响应式调整 */
@media (max-width: 768px) {
  .about {
    padding: 1rem;
  }

  .about-intro h1 {
    font-size: 2rem;
  }

  .achievement-item h3 {
    font-size: 2rem;
  }
}
</style>
