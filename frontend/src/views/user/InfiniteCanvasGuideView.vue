<template>
  <AppLayout>
    <main class="guide-shell">
      <header class="guide-hero">
        <div>
          <p class="eyebrow">SUB2API · 无限画布</p>
          <h1>配置模型，在画布上开始创作</h1>
          <p class="lead">
            从创建分组密钥、接入图片与视频模型，到生成、编辑和整理作品，按这份指南完成首次配置。
          </p>
          <div class="hero-actions">
            <router-link class="primary-action" to="/canvas"
              >打开无限画布 <span aria-hidden="true">↗</span></router-link
            >
            <router-link class="secondary-action" to="/keys"
              >管理 API Key</router-link
            >
          </div>
          <p class="version-note">
            画布渠道配置保存在当前浏览器中；中转分组只会开放已授权的模型。
          </p>
        </div>
        <div class="hero-art" aria-label="无限画布工作区示意图">
          <div class="window-bar">
            <i></i><i></i><i></i><span>INFINITE CANVAS</span><b>工作区</b>
          </div>
          <div class="mock-layout">
            <div class="mock-rail">
              <b>JD</b><span>⌂</span><span>▧</span><span>◉</span><span>⚙</span>
            </div>
            <div class="mock-board">
              <div class="board-title">
                <small>CANVAS</small><strong>灵感工作区</strong>
              </div>
              <div class="board-grid">
                <div class="mock-node prompt-node">
                  <small>提示词</small><strong>清晨的海岸与薄雾</strong
                  ><i>可连接图片或视频节点</i>
                </div>
                <div class="mock-node image-node">
                  <div class="fake-image"><span>IMAGE</span></div>
                  <small>图片生成</small><b>gpt-image-2</b>
                </div>
                <div class="mock-node video-node">
                  <div class="fake-video"><span>▶</span></div>
                  <small>视频生成</small><b>Sora / 视频模型</b>
                </div>
                <span class="connector connector-one"></span
                ><span class="connector connector-two"></span>
              </div>
              <div class="mock-toolbar">
                <span>＋ 文字</span><span>▧ 图片</span><span>▶ 视频</span
                ><span>⚙ 配置</span>
              </div>
            </div>
          </div>
          <div class="art-caption">画布节点与连线示意</div>
        </div>
      </header>

      <div class="guide-grid">
        <aside class="toc">
          <strong>本页目录</strong><a href="#quickstart">快速配置</a
          ><a href="#channels">渠道与模型</a><a href="#workflow">画布生成流程</a
          ><a href="#troubleshoot">问题排查</a><a href="#security">安全提示</a>
        </aside>
        <article class="guide-content">
          <section id="quickstart" class="guide-section">
            <div class="section-heading">
              <span class="section-index">01</span>
              <div>
                <h2>快速配置</h2>
                <p>首次使用，先确保分组有权限，再把中转地址和模型加到画布。</p>
              </div>
            </div>
            <div class="steps-list">
              <div class="quick-step">
                <b>准备分组 API Key</b>
                <p>
                  在左侧「API
                  Key」页面创建或选择一个密钥，确认它关联的分组已开放需要的图片、视频模型，并且账户仍有可用额度。
                </p>
                <router-link to="/keys">前往 API Key 管理 ↗</router-link>
              </div>
              <div class="quick-step">
                <b>进入画布设置</b>
                <p>
                  打开「无限画布」，点击右上角齿轮进入「配置」，默认显示「渠道」页签，再点「新增渠道」。
                </p>
              </div>
              <div class="quick-step">
                <b>填写中转连接</b>
                <p>
                  渠道名称可自定义；「协议」按中转 API 类型选择。接入本站 OpenAI
                  兼容接口时选择 OpenAI；接口地址填写本站域名，API Key
                  填上一步的分组密钥。
                </p>
                <div class="inline-code">
                  <span>接口地址</span><code>https://你的中转站域名</code
                  ><small>可不带 /v1；画布会自动补齐 API 路径</small>
                </div>
              </div>
              <div class="quick-step">
                <b>拉取模型并保存</b>
                <p>
                  点击「选择模型」→「拉取模型列表」，勾选要用的模型并确定。回到渠道编辑页检查模型能力，再点右上角「保存」。如果上游不支持模型列表，可以按中转开放的模型
                  ID 手动添加。
                </p>
              </div>
            </div>
            <div class="notice">
              <b>模型能力要匹配</b
              ><span
                >图片模型标为「生图」，视频模型标为「视频」。模型不会因为写入了
                API Key 就自动获得分组权限。</span
              >
            </div>
          </section>

          <section id="channels" class="guide-section">
            <div class="section-heading">
              <span class="section-index">02</span>
              <div>
                <h2>渠道与模型</h2>
                <p>
                  一条画布渠道对应一组接口地址、API Key
                  和模型列表。图片和视频可以放在同一渠道，也可以分开配置。
                </p>
              </div>
            </div>
            <div class="module-list">
              <section class="module">
                <div class="module-top">
                  <b>图</b>
                  <h3>图片模型</h3>
                  <span>文生图 / 编辑</span>
                </div>
                <p>
                  在模型列表中选择分组实际开放的图片模型，将能力设为「生图」。生成图片后，可继续用局部编辑、裁剪、切图或超分等工具处理，实际可用功能取决于模型支持范围。
                </p>
                <div class="mini-ui">
                  <label>渠道模型</label>
                  <div class="mini-option active">
                    <span>gpt-image-2</span><em>生图</em>
                  </div>
                  <div class="mini-option">
                    <span>[超分] gpt-image-2</span
                    ><small>按中转开放的请求模型 ID 添加</small>
                  </div>
                </div>
                <p class="tip">
                  如果分组提供 `[超分]` 模型别名，请按模型列表展示的 ID
                  原样选择。别自行把请求模型名改成实际上游映射名。
                </p>
              </section>
              <section class="module">
                <div class="module-top">
                  <b>影</b>
                  <h3>视频模型</h3>
                  <span>文生视频 / 参考图生视频</span>
                </div>
                <p>
                  拉取模型后将对应模型能力设为「视频」。若在中转后台添加的是
                  Sora 上游渠道，按上游配置要求选择 Sora
                  协议并将模型归类为视频；画布渠道的「协议」则按它调用的 API
                  兼容格式选择，接入本站 OpenAI 兼容接口通常选 OpenAI。
                </p>
                <div class="mini-ui">
                  <label>两处配置分别确认</label>
                  <div class="mini-option">
                    <span>中转后台 · 上游渠道协议</span
                    ><em>Sora（上游要求时）</em>
                  </div>
                  <div class="mini-option active">
                    <span>无限画布 · API 格式</span><em>OpenAI</em>
                  </div>
                  <div class="mini-option active">
                    <span>无限画布 · 模型能力</span><em>视频</em>
                  </div>
                </div>
                <p class="tip">
                  不要把「上游协议」和画布侧的「API
                  格式」混为一项。本站画布渠道编辑器提供 OpenAI、Gemini
                  两种格式。
                </p>
              </section>
            </div>
            <div class="reference-note">
              <strong>参考教程</strong>
              <p>
                你提供的教程包含服务端添加图片/视频渠道、拉取模型及 Sora
                视频配置截图。不同站点版本的后台字段可能略有差异，请以你当前管理端的渠道表单和上游
                API 文档为准。
              </p>
              <a
                href="https://ai-document.yyapi.cloud/zh/canvas-page"
                target="_blank"
                rel="noopener noreferrer"
                >查看原始无限画布教程 ↗</a
              >
            </div>
          </section>

          <section id="workflow" class="guide-section">
            <div class="section-heading">
              <span class="section-index">03</span>
              <div>
                <h2>画布生成流程</h2>
                <p>
                  画布由文字、图片、视频等节点组成。先选模型和参数，再按需要连接参考素材。
                </p>
              </div>
            </div>
            <div class="steps-list">
              <div class="quick-step">
                <b>添加创作节点</b>
                <p>
                  在画布底部工具栏添加文字、图片或视频节点，也可以先上传已有图片或视频素材。拖动画布空白处平移，滚轮缩放；选中节点后可移动和调整内容。
                </p>
              </div>
              <div class="quick-step">
                <b>输入提示词并选择模型</b>
                <p>
                  在图片/视频节点或生成配置节点中填写提示词，使用模型选择器选一个已配置且能力匹配的模型。点击节点上的「开始生成」；图片生成支持设置尺寸、比例、质量和张数，视频可设置分辨率、比例、时长与生成声音。
                </p>
              </div>
              <div class="quick-step">
                <b>连接参考图或视频</b>
                <p>
                  把图片节点连到视频节点，可用于首尾帧或参考素材。视频设置中的「首尾帧模式」适合指定起始/结束画面；「全能参考模式」适合使用多张参考图。生成时间取决于模型和上游排队情况。
                </p>
              </div>
              <div class="quick-step">
                <b>整理和导出结果</b>
                <p>
                  选中结果节点可查看、下载、重试或加入资产。需要保留多个方案时复制节点或切换为「新建节点」的重新生成方式；重要画布可用配置里的
                  WebDAV 同步备份。
                </p>
              </div>
            </div>
            <div class="workflow-strip" aria-label="画布创作流程">
              <span>提示词</span><i>→</i><span>图片 / 视频节点</span><i>→</i
              ><span>参考素材连线</span><i>→</i><span>生成与整理</span>
            </div>
          </section>

          <section id="troubleshoot" class="guide-section">
            <div class="section-heading">
              <span class="section-index">04</span>
              <div>
                <h2>问题排查</h2>
                <p>先看模型权限、请求模型 ID 和渠道配置，再根据错误码定位。</p>
              </div>
            </div>
            <div class="faq-list">
              <details>
                <summary>提示 Model is not available for this group</summary>
                <p>
                  当前 API Key 所属分组没有开放该请求模型
                  ID，或后台模型映射未配置。回到「API
                  Key」确认分组；再检查模型管理/上游映射，并在画布中按分组可用的
                  ID
                  重新拉取或手动添加。超分别名要按原样请求，不要改成实际上游模型名。
                </p>
              </details>
              <details>
                <summary>拉取模型列表失败</summary>
                <p>
                  检查接口地址是否为本站域名、API Key
                  是否有效，以及该分组能否访问模型列表接口。可先手动添加分组实际开放的模型
                  ID；添加后仍需把能力设为生图或视频。
                </p>
              </details>
              <details>
                <summary>图片可以生成，视频报 404 或协议错误</summary>
                <p>
                  确认中转服务已部署并开放视频接口，视频模型被设为「视频」，并核对后台上游协议与画布侧
                  API 格式。对接本站 OpenAI 兼容视频接口时，画布侧选
                  OpenAI；上游需 Sora 协议时只在中转后台按上游要求设置。
                </p>
              </details>
              <details>
                <summary>视频任务长时间没有结果</summary>
                <p>
                  检查上游任务是否支持查询/轮询、密钥额度和服务状态。长视频任务可能需要等待；超时后查看节点错误信息，再按提示重试或换用当前分组可用的模型。
                </p>
              </details>
              <details>
                <summary>设置换浏览器后不见了</summary>
                <p>
                  渠道配置保存在当前浏览器本地，不会自动跟随账号同步。可从「配置」导出
                  JSON 迁移渠道；WebDAV 可同步画布、资产和生成记录，但不包含 AI
                  API Key。导出的 JSON 含 API Key，请妥善保管。
                </p>
              </details>
            </div>
          </section>

          <section id="security" class="guide-section security-section">
            <div class="section-heading">
              <span class="section-index">05</span>
              <div>
                <h2>安全提示</h2>
                <p>画布在浏览器中直接使用你填入的中转 API Key。</p>
              </div>
            </div>
            <ul>
              <li>
                不要把 API Key 放进提示词、公开截图、插件或不可信的配置文件。
              </li>
              <li>
                分享画布前检查节点内容，避免包含密钥、个人数据或敏感参考图。
              </li>
              <li>
                导出的配置文件包含渠道密钥；只保存到可信位置，分享前先移除凭证。
              </li>
              <li>
                使用浏览器插件或第三方节点插件前，确认其来源可信；插件可能接触本地画布数据。
              </li>
            </ul>
          </section>
          <footer class="guide-footer">
            <span>无限画布 · Sub2API 使用教程</span
            ><router-link to="/canvas">返回无限画布 ↗</router-link>
          </footer>
        </article>
      </div>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import AppLayout from "@/components/layout/AppLayout.vue";
</script>

<style scoped>
.guide-shell {
  max-width: 1320px;
  margin: 0 auto;
  padding: 28px 32px 64px;
  color: #202a34;
}
.guide-hero {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 42px;
  align-items: center;
  padding: 32px 0 40px;
  border-bottom: 1px solid #e2e8e8;
}
.eyebrow {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 1px;
  color: #087e73;
}
.guide-hero h1 {
  font-size: 36px;
  line-height: 1.15;
  margin: 14px 0 12px;
}
.lead {
  max-width: 600px;
  font-size: 16px;
  line-height: 1.75;
  color: #58656d;
}
.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 24px;
}
.primary-action,
.secondary-action {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 11px 16px;
  border-radius: 5px;
  text-decoration: none;
  font-weight: 650;
  font-size: 14px;
}
.primary-action {
  background: #087e73;
  color: white;
}
.secondary-action {
  border: 1px solid #cbd5d5;
  color: #27343b;
}
.version-note {
  font-size: 12px;
  color: #718087;
  margin-top: 18px;
}
.hero-art {
  border: 1px solid #d8e1e1;
  border-radius: 7px;
  overflow: hidden;
  background: #f3f7f6;
  box-shadow: 0 8px 30px #203a3810;
}
.window-bar {
  height: 34px;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 12px;
  background: #e8efed;
  color: #5c686b;
  font-size: 11px;
}
.window-bar i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #aab7b3;
}
.window-bar span {
  margin-left: 8px;
}
.window-bar b {
  margin-left: auto;
  color: #087e73;
  font-weight: 600;
}
.mock-layout {
  display: grid;
  grid-template-columns: 42px 1fr;
  min-height: 280px;
}
.mock-rail {
  display: flex;
  align-items: center;
  flex-direction: column;
  gap: 18px;
  padding: 14px 0;
  background: #fbfdfc;
  color: #82908d;
}
.mock-rail b {
  color: #087e73;
}
.mock-board {
  position: relative;
  min-width: 0;
  padding: 18px;
  background-color: #f9fbfa;
  background-image: radial-gradient(#cbd8d4 1px, transparent 1px);
  background-size: 17px 17px;
}
.board-title small,
.board-title strong {
  display: block;
}
.board-title small {
  font-size: 9px;
  color: #82908d;
  letter-spacing: 1px;
}
.board-title strong {
  font-size: 17px;
  margin-top: 3px;
}
.board-grid {
  position: relative;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: 17px;
}
.mock-node {
  position: relative;
  z-index: 1;
  padding: 9px;
  border: 1px solid #dce5e2;
  border-radius: 5px;
  background: #fff;
  box-shadow: 0 4px 12px #203a380c;
}
.mock-node small,
.mock-node b {
  display: block;
  font-size: 9px;
  color: #6b7875;
}
.mock-node strong {
  display: block;
  font-size: 10px;
  line-height: 1.5;
  margin: 8px 0;
}
.mock-node i {
  font-style: normal;
  font-size: 8px;
  color: #82908d;
}
.prompt-node {
  grid-column: 1 / -1;
  width: 56%;
}
.image-node {
  grid-column: 1;
}
.video-node {
  grid-column: 2;
}
.fake-image,
.fake-video {
  display: grid;
  place-items: center;
  height: 58px;
  border-radius: 3px;
  background: #c8dfd6;
  color: #496961;
  font-size: 8px;
  font-weight: 700;
}
.fake-video {
  background: #586f6b;
  color: #fff;
}
.fake-video span {
  display: grid;
  place-items: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #243c38a8;
}
.connector {
  position: absolute;
  z-index: 0;
  height: 1px;
  background: #13877c;
  transform-origin: left center;
}
.connector-one {
  width: 28px;
  left: 48%;
  top: 61%;
  transform: rotate(26deg);
}
.connector-two {
  width: 34px;
  left: 51%;
  top: 60%;
  transform: rotate(-25deg);
}
.mock-toolbar {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 5px;
  margin-top: 14px;
}
.mock-toolbar span {
  padding: 6px 7px;
  border: 1px solid #dce5e2;
  border-radius: 4px;
  background: #fff;
  color: #52635e;
  font-size: 9px;
}
.art-caption {
  text-align: right;
  padding: 0 12px 10px;
  color: #84918f;
  font-size: 10px;
}
.guide-grid {
  display: grid;
  grid-template-columns: 180px minmax(0, 1fr);
  gap: 52px;
  padding-top: 34px;
}
.toc {
  position: sticky;
  top: 24px;
  align-self: start;
  display: flex;
  flex-direction: column;
  gap: 13px;
  font-size: 13px;
}
.toc strong {
  margin-bottom: 3px;
}
.toc a {
  color: #65737a;
  text-decoration: none;
}
.toc a:hover {
  color: #087e73;
}
.guide-content {
  min-width: 0;
  max-width: 920px;
}
.guide-section {
  padding: 0 0 48px;
  scroll-margin-top: 24px;
}
.section-heading {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  margin-bottom: 24px;
}
.section-index {
  font-size: 12px;
  font-weight: 700;
  color: #087e73;
  background: #eaf5f2;
  padding: 6px 8px;
  border-radius: 3px;
}
.section-heading h2 {
  font-size: 23px;
  margin: 0 0 5px;
}
.section-heading p {
  margin: 0;
  color: #6c797e;
  font-size: 14px;
  line-height: 1.65;
}
.steps-list {
  border-left: 1px solid #d7e2df;
  margin-left: 11px;
}
.quick-step {
  position: relative;
  padding: 0 0 21px 24px;
}
.quick-step:before {
  content: "";
  position: absolute;
  left: -5px;
  top: 5px;
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: #13877c;
  border: 2px solid white;
  box-shadow: 0 0 0 1px #13877c;
}
.quick-step > b {
  font-size: 15px;
}
.quick-step p,
.module p {
  font-size: 13px;
  color: #637177;
  line-height: 1.75;
  margin: 6px 0 0;
}
.quick-step > a {
  display: inline-block;
  margin-top: 7px;
  color: #087e73;
  font-size: 12px;
  font-weight: 650;
  text-decoration: none;
}
.inline-code {
  display: grid;
  grid-template-columns: 90px minmax(0, 1fr);
  gap: 6px 12px;
  align-items: center;
  margin-top: 11px;
  padding: 12px;
  background: #f5f8f7;
  border: 1px solid #e3e9e8;
  border-radius: 4px;
  font-size: 12px;
}
.inline-code span {
  color: #637177;
}
.inline-code code {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #254740;
}
.inline-code small {
  grid-column: 2;
  color: #7a888a;
}
.notice {
  display: flex;
  gap: 14px;
  padding: 13px 15px;
  background: #eef7f5;
  border-left: 3px solid #13877c;
  font-size: 13px;
  color: #405d58;
}
.module-list {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px;
}
.module {
  border-top: 2px solid #dce7e4;
  padding: 14px 0 6px;
}
.module-top {
  display: flex;
  align-items: center;
  gap: 9px;
}
.module-top > b {
  display: grid;
  place-items: center;
  width: 24px;
  height: 24px;
  background: #e9f5f2;
  color: #087e73;
  border-radius: 50%;
  font-size: 11px;
}
.module-top h3 {
  font-size: 16px;
  margin: 0;
}
.module-top span {
  margin-left: auto;
  font-size: 10px;
  color: #7a888a;
}
.module p {
  margin-top: 9px;
}
.mini-ui {
  margin-top: 13px;
  border: 1px solid #e0e7e5;
  border-radius: 4px;
  padding: 11px;
  background: #fbfcfc;
  font-size: 11px;
}
.mini-ui > label {
  display: block;
  font-weight: 700;
  margin-bottom: 9px;
  color: #35423f;
}
.mini-option {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  padding: 8px;
  border: 1px solid #e4e9e8;
  background: white;
  margin-top: 5px;
  border-radius: 3px;
}
.mini-option span {
  font-weight: 650;
}
.mini-option small {
  color: #778481;
}
.mini-option em {
  font-style: normal;
  margin-left: auto;
  color: #168276;
}
.mini-option.active {
  border-color: #70b7ac;
}
.tip {
  padding: 9px 10px;
  background: #f5f7f6;
  border-radius: 3px;
  font-size: 11px !important;
}
.reference-note {
  margin-top: 25px;
  padding: 16px 18px;
  border: 1px solid #d9e7e3;
  border-radius: 5px;
  background: #f7faf9;
}
.reference-note strong {
  font-size: 14px;
}
.reference-note p {
  font-size: 12px;
  line-height: 1.7;
  color: #64716f;
  margin: 6px 0 9px;
}
.reference-note a {
  font-size: 12px;
  color: #087e73;
  font-weight: 650;
  text-decoration: none;
}
.workflow-strip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 23px;
  padding: 14px;
  background: #f4f7f6;
  border: 1px solid #e3e9e8;
  border-radius: 4px;
}
.workflow-strip span {
  font-size: 11px;
  font-weight: 650;
  color: #35514c;
}
.workflow-strip i {
  font-style: normal;
  color: #90a29e;
}
.faq-list {
  border-top: 1px solid #e2e8e7;
}
.faq-list details {
  border-bottom: 1px solid #e2e8e7;
  padding: 14px 2px;
}
.faq-list summary {
  cursor: pointer;
  font-size: 14px;
  font-weight: 650;
}
.faq-list details p {
  font-size: 13px;
  line-height: 1.7;
  color: #68767b;
  margin: 10px 0 2px;
}
.security-section ul {
  padding-left: 20px;
  color: #59676d;
  font-size: 13px;
  line-height: 1.9;
}
.guide-footer {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  border-top: 1px solid #e2e8e7;
  padding-top: 18px;
  font-size: 12px;
  color: #758187;
}
.guide-footer a {
  color: #087e73;
  text-decoration: none;
  font-weight: 650;
}
.dark .guide-shell {
  color: #e5eceb;
}
.dark .guide-hero,
.dark .guide-footer,
.dark .faq-list,
.dark .faq-list details {
  border-color: #34413f;
}
.dark .lead,
.dark .quick-step p,
.dark .module p,
.dark .section-heading p,
.dark .toc a,
.dark .security-section ul,
.dark .reference-note p {
  color: #aab8b5;
}
.dark .hero-art,
.dark .mock-node,
.dark .mock-toolbar span,
.dark .mini-ui,
.dark .mini-option {
  background: #202a29;
  border-color: #394744;
  color: #d5dfdc;
}
.dark .window-bar,
.dark .mock-rail {
  background: #26312f;
  color: #c0ccc8;
}
.dark .mock-board {
  background-color: #26312f;
  background-image: radial-gradient(#41514d 1px, transparent 1px);
}
.dark .fake-image {
  background: #59736a;
}
.dark .fake-video {
  background: #304945;
}
.dark .notice,
.dark .reference-note {
  background: #243b37;
  border-color: #3a5b54;
  color: #cfdfdc;
}
.dark .inline-code,
.dark .workflow-strip,
.dark .tip {
  background: #26312f;
  border-color: #394744;
  color: #d5dfdc;
}
.dark .inline-code code {
  color: #b9e0d4;
}
.dark .mini-option small {
  color: #aab8b5;
}
@media (max-width: 900px) {
  .guide-shell {
    padding: 22px 20px 48px;
  }
  .guide-hero {
    grid-template-columns: 1fr;
    gap: 22px;
  }
  .hero-art {
    max-width: 680px;
  }
  .guide-grid {
    grid-template-columns: 1fr;
    gap: 20px;
  }
  .toc {
    position: static;
    flex-direction: row;
    flex-wrap: wrap;
    align-items: center;
    border-bottom: 1px solid #e2e8e8;
    padding-bottom: 14px;
  }
  .toc strong {
    width: 100%;
  }
}
@media (max-width: 560px) {
  .guide-shell {
    padding: 16px 14px 36px;
  }
  .guide-hero {
    padding-top: 18px;
  }
  .guide-hero h1 {
    font-size: 29px;
  }
  .lead {
    font-size: 14px;
  }
  .mock-layout {
    grid-template-columns: 34px 1fr;
  }
  .mock-board {
    padding: 13px;
  }
  .board-grid {
    gap: 8px;
  }
  .fake-image,
  .fake-video {
    height: 46px;
  }
  .guide-grid {
    padding-top: 22px;
  }
  .section-heading h2 {
    font-size: 20px;
  }
  .module-list {
    grid-template-columns: 1fr;
  }
  .notice {
    flex-direction: column;
    gap: 4px;
  }
  .workflow-strip {
    justify-content: flex-start;
  }
  .workflow-strip i {
    display: none;
  }
  .workflow-strip span {
    padding: 7px;
    background: #fff;
    border: 1px solid #e2e8e7;
    border-radius: 3px;
  }
  .inline-code {
    grid-template-columns: 1fr;
  }
  .inline-code small {
    grid-column: 1;
  }
  .guide-footer {
    flex-direction: column;
  }
}
</style>
