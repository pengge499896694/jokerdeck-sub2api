<template>
  <AppLayout>
    <main class="guide-shell">
      <header class="guide-hero">
        <div>
          <p class="eyebrow">SUB2API · 无限画布</p>
          <h1>无限画布新手上手指南</h1>
          <p class="lead">
            从填写中转地址和密钥、拉取模型，到新建画布、文生图、图生图和视频创作，跟着步骤完成第一次创作。
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
            本页按当前无限画布界面编写。先完成渠道配置；不想用提示词库时，也可以直接输入提示词。
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
          <strong>本页目录</strong><a href="#quickstart">开始前准备</a
          ><a href="#canvas-start">新建画布</a
          ><a href="#image-workbench">文生图与图生图</a><a href="#video-workbench">视频创作台</a
          ><a href="#network">GitHub 与网络</a><a href="#channels">模型与协议说明</a
          ><a href="#workflow">画布节点流程</a><a href="#troubleshoot">问题排查</a
          ><a href="#security">安全提示</a>
        </aside>
        <article class="guide-content">
          <section id="quickstart" class="guide-section">
            <div class="section-heading">
              <span class="section-index">01</span>
              <div>
                <h2>快速配置</h2>
                <p>准备好中转站分组密钥，并确认对应分组已开通图片或视频模型。</p>
              </div>
            </div>
            <div class="steps-list">
              <div class="quick-step">
                <b>准备分组 API Key</b>
                <p>
                  在本站左侧「API 密钥」页面创建或复制一个密钥。确认密钥所属分组有你要用的图片/视频模型，并且账户有可用额度。后面要把这个密钥填入画布渠道的 API Key。
                </p>
                <router-link to="/keys">前往 API Key 管理 ↗</router-link>
              </div>
              <div class="quick-step">
                <b>进入画布设置</b>
                <p>
                  打开「无限画布」，点击右上角「配置」。在配置弹窗的「渠道」页签点击「新增渠道」；图片、视频可共用一条渠道，也可以分别添加。
                </p>
              </div>
              <div class="quick-step">
                <b>填写中转连接</b>
                <p>
                  渠道名称自己容易认即可。「协议」选择中转接口实际兼容的格式；本站 OpenAI 兼容接口通常选 OpenAI。接口地址填中转站 API 地址，API Key 填上一步的分组密钥。不要把上游厂商地址或代理节点地址填在这里。
                </p>
                <div class="inline-code">
                  <span>接口地址</span><code>https://你的中转站域名</code
                  ><small>填写你平常用于 API 调用的中转站地址；通常可不带 /v1，若服务商明确要求完整路径则按其说明填写</small>
                </div>
              </div>
              <div class="quick-step">
                <b>拉取模型并保存</b>
                <p>
                  点击「选择模型」→「拉取模型列表」，勾选要用的模型后点「确定」。弹窗的搜索框可筛选名称；「新获取的模型」是刚拉取的列表，「已有的模型」是渠道原有列表。「全选当前列表」只选择当前页签可见项目。回到渠道编辑页，把图片模型的能力设为「生图」、视频模型设为「视频」，最后点「保存」。如果拉取不到，在「输入模型名称」框填中转后台提供的模型 ID，再点「增加模型」。
                </p>
              </div>
            </div>
            <figure class="guide-shot">
              <img src="/canvas-guide/01-canvas-home.png" alt="无限画布首页和顶部导航截图" loading="lazy" @click="openGuideImage('/canvas-guide/01-canvas-home.png', '无限画布首页和顶部导航截图')" />
              <figcaption>图 1：从顶部导航进入「无限画布」、生图工作台和视频创作台。</figcaption>
            </figure>
            <figure class="guide-shot guide-shot-compact">
              <img src="/canvas-guide/02-canvas-config.png" alt="无限画布配置入口和新增渠道截图" loading="lazy" @click="openGuideImage('/canvas-guide/02-canvas-config.png', '无限画布配置入口和新增渠道截图')" />
              <figcaption>图 2：点击右上角配置图标，再点击「新增渠道」。</figcaption>
            </figure>
            <div class="notice">
              <b>模型能力要匹配</b
              ><span
                >画布只是保存连接方式，不会替分组开通模型。模型名称和权限以中转站给该密钥开放的列表为准。</span
              >
            </div>
            <div class="guide-shot-grid">
              <figure class="guide-shot">
                <img src="/canvas-guide/03-api-key-group.png" alt="API Key 所属绘图分组截图" loading="lazy" @click="openGuideImage('/canvas-guide/03-api-key-group.png', 'API Key 所属绘图分组截图')" />
                <figcaption>图 3：API Key 要选中已开通生图或视频模型的分组。</figcaption>
              </figure>
              <figure class="guide-shot">
                <img src="/canvas-guide/04-channel-models.png" alt="渠道填写地址密钥和选择模型截图" loading="lazy" @click="openGuideImage('/canvas-guide/04-channel-models.png', '渠道填写地址密钥和选择模型截图')" />
                <figcaption>图 4：填写中转地址和 API Key，拉取模型后全选并确认。</figcaption>
              </figure>
            </div>
          </section>

          <section id="canvas-start" class="guide-section">
            <div class="section-heading">
              <span class="section-index">02</span>
              <div>
                <h2>新建画布</h2>
                <p>无限画布适合把提示词、参考图片和生成结果放在同一张可移动的工作板上。</p>
              </div>
            </div>
            <div class="steps-list">
              <div class="quick-step"><b>进入「我的画布」</b><p>在无限画布左侧或顶部导航点「画布 / 我的画布」，进入画布列表。点击「新建画布」，输入容易辨认的名称后创建并打开。</p></div>
              <div class="quick-step"><b>认识工作区</b><p>中央点阵区域是画布；按住空白处拖动可平移，滚轮可缩放。左侧「画布」查看节点、「资产」选择已有素材、「提示词库」浏览灵感。右上角「配置」管理渠道，右下角缩放控件可调整视野。</p></div>
              <div class="quick-step"><b>添加节点</b><p>底部工具栏的「文本」放提示词，「图片」创建图片节点，「视频」创建视频节点，「音频」创建音频节点，「生成配置」统一设置模型和参数；「上传资产」可把本地素材放进画布。按住节点可移动；「组」用于整理多个节点；「画布外观」切换画布视觉样式；「清空画布」会移除当前画布全部内容，操作前务必确认。新手可先点「图片」开始文生图。</p></div>
              <div class="quick-step"><b>连参考素材并生成</b><p>把图片节点边缘的连接点拖到另一个节点的连接点，即可传递参考内容。图片节点连到图片生成节点用于图生图；视频节点可接图片作为首帧/参考图。填写提示词、选模型后点击节点里的生成按钮。</p></div>
            </div>
            <div class="screen-map" role="img" aria-label="无限画布控件位置图">
              <div class="screen-top"><span>无限画布</span><span>配置 · 快捷键 · Agent</span></div>
              <div class="screen-body"><div class="screen-side">画布<br />资产<br />提示词库</div><div class="screen-board"><span class="screen-node">文本节点<br /><small>描述你想生成的画面</small></span><span class="screen-node image">图片节点<br /><small>选择模型 · 开始生成</small></span><span class="screen-link"></span></div></div>
              <div class="screen-tools">移动  撤销  重做  文本  图片  视频  音频  生成配置  上传资产</div>
              <p>控件位置图 · 依据 Edge 当前版本界面整理，非实拍截图</p>
            </div>
            <figure class="guide-shot">
              <img src="/canvas-guide/05-canvas-image.png" alt="无限画布图片节点上传参考图截图" loading="lazy" @click="openGuideImage('/canvas-guide/05-canvas-image.png', '无限画布图片节点上传参考图截图')" />
              <figcaption>图 5：图片节点可上传参考图；文本节点填写提示词，再连接到生成节点。</figcaption>
            </figure>
          </section>

          <section id="image-workbench" class="guide-section">
            <div class="section-heading"><span class="section-index">03</span><div><h2>文生图与图生图</h2><p>「生图工作台」是快速单次生成页面；不需要先创建画布节点。</p></div></div>
            <div class="steps-list">
              <div class="quick-step"><b>打开生图工作台</b><p>顶部导航点击「生图工作台」。左侧是生成记录；中间填写提示词和参数；右侧显示生成结果。手机或窄屏上，历史记录和设置会收进按钮/抽屉中。</p></div>
              <div class="quick-step"><b>先选图片模型</b><p>「模型」选择已配置且能力为「生图」的模型。若列表为空，回到右上角「配置」检查渠道是否保存、模型能力是否选成「生图」。第一次成功配置后再回来选择。</p></div>
              <div class="quick-step"><b>文生图：只写描述</b><p>在「提示词」输入主体、环境、风格、构图、光线和用途，例如“雨后街道上的红色自行车，电影感摄影，侧逆光，横向构图”。不上传参考图时，点击「开始生成」就是文生图。</p></div>
              <div class="quick-step"><b>图生图：添加参考图</b><p>在「参考图」点「上传」选择照片，也可点「剪切板」读取已复制的图片，或把图片拖入参考区域。确认缩略图已出现，再补充希望怎么修改；上传参考图后点击「开始生成」会走图片编辑/图生图请求。删除不需要的参考图再生成。</p></div>
              <div class="quick-step"><b>理解生成参数</b><p>「质量」控制输出质量档位；「尺寸」设置宽高，旁边的「16 倍数对齐」会把输入调整为 16 的倍数；「分辨率」提供 1k/2k/4k/自动；「宽高比」控制画面横竖形状；「透明背景」只对部分模型有效；「生成张数」一次最多选到 10 张，每张作为单独请求生成。模型不支持的选项可能被忽略或报错，张数增加也会增加耗时和费用。</p></div>
              <div class="quick-step"><b>查看和复用结果</b><p>右侧「生成结果」显示图片，可下载或保存到「我的资产」；结果也会写入左侧「生成记录」。点历史项目可恢复当时的提示词和参数；「新建」用于开始一组新的工作，不会删除历史记录。</p></div>
            </div>
            <div class="screen-map workbench-map" role="img" aria-label="生图工作台控件位置图">
              <div class="screen-top"><span>生图工作台</span><span>生成记录  设置</span></div>
              <div class="screen-body"><div class="screen-side">生成记录<br /><small>历史任务列表</small></div><div class="screen-form"><b>提示词</b><div class="screen-field">描述主体、风格、构图、光线和用途</div><b>参考图</b><div class="screen-field actions">剪切板  上传  拖入图片</div><b>模型  质量  尺寸  分辨率  宽高比</b><div class="screen-field actions">选择已配置模型  自动 / 1k / 2k / 1:1</div><span class="screen-button">开始生成</span></div><div class="screen-result"><b>生成结果</b><div class="screen-placeholder">结果预览区<br />下载 · 加入资产</div></div></div>
              <p>控件位置图 · 依据 Edge 当前版本界面整理，非实拍截图</p>
            </div>
            <figure class="guide-shot">
              <img src="/canvas-guide/06-image-workbench.png" alt="生图工作台提示词参考图模型和参数截图" loading="lazy" @click="openGuideImage('/canvas-guide/06-image-workbench.png', '生图工作台提示词参考图模型和参数截图')" />
              <figcaption>图 6：依次填写提示词、可选上传参考图、选择模型和质量/分辨率/比例，最后点击生成。</figcaption>
            </figure>
            <div class="notice"><b>图生图判断方法</b><span>是否加了参考图是关键：无参考图为文生图，有参考图时工作台调用编辑流程。模型必须支持图像输入；只支持纯文本生图的模型不能可靠地按参考图修改。</span></div>
          </section>

          <section id="video-workbench" class="guide-section">
            <div class="section-heading"><span class="section-index">04</span><div><h2>视频创作台</h2><p>先选视频模型，再写镜头描述；可选图片、视频或音频作为参考。</p></div></div>
            <div class="steps-list">
              <div class="quick-step"><b>打开并选择模型</b><p>顶部导航点「视频创作台」，在「模型」中选择能力为「视频」的模型。若找不到模型，先去「配置 → 渠道」拉取/添加模型并把能力设为「视频」。</p></div>
              <div class="quick-step"><b>写清楚镜头和动作</b><p>在「提示词」写主体做什么、镜头如何移动、场景氛围、画面风格及希望的节奏。点「查看提示词库」可找例子，点「查看我的资产」可插入已保存的文字或图片；提示词库访问 GitHub 失败时，直接手动写也能继续。</p></div>
              <div class="quick-step"><b>可选添加参考图</b><p>「参考图」区域可点「剪切板」读取已复制的图片、点「上传」选择图片，或把图片拖进虚线框；工作台最多保留 7 张图片。缩略图上的序号代表参考顺序，可调整顺序或移除。需要视频首尾帧/多素材流程时，建议在画布中使用视频节点及连线；此工作台没有参考视频/音频上传控件。</p></div>
              <div class="quick-step"><b>设置视频输出</b><p>点「设置」打开参数：清晰度有 480p/720p/1080p，也可输入数值；「尺寸」设置宽高；「比例」可选横屏、竖屏、方形、宽屏、长图或自动；「时长」范围为 4–30 秒。「首尾帧模式」用于指定开头/结束画面，「全能参考模式」适用于参考图创作；还可按需要打开「生成声音」或「添加水印」。模型不支持的模式需改用模型兼容选项。</p></div>
              <div class="quick-step"><b>生成、等待和保存</b><p>确认提示词和参考素材后点「开始生成」。视频通常先创建异步任务，右侧结果区会显示等待时间，完成后可预览、下载或加入「我的资产」。失败时先展开错误提示，核对模型能力、接口支持和账户额度，再重试。</p></div>
            </div>
            <div class="screen-map workbench-map video-map" role="img" aria-label="视频创作台控件位置图">
              <div class="screen-top"><span>视频创作台</span><span>生成记录  设置</span></div>
              <div class="screen-body"><div class="screen-side">生成记录<br /><small>任务状态与历史</small></div><div class="screen-form"><b>提示词</b><div class="screen-field tall">镜头运动 · 主体动作 · 场景氛围 · 画面风格</div><b>参考图</b><div class="screen-field actions">剪切板  上传  最多 7 张图片</div><b>视频参数（点设置）</b><div class="screen-field actions">模型  清晰度  比例  时长  模式</div><span class="screen-button">开始生成</span></div><div class="screen-result"><b>生成结果</b><div class="screen-placeholder video-placeholder">视频预览区<br />下载 · 加入资产</div></div></div>
              <p>控件位置图 · 依据当前界面与源码整理，非实拍截图；参数生效范围取决于所选模型</p>
            </div>
            <div class="guide-shot-grid">
              <figure class="guide-shot">
                <img src="/canvas-guide/07-sora-channel.png" alt="Sora 视频渠道协议和模型能力配置截图" loading="lazy" @click="openGuideImage('/canvas-guide/07-sora-channel.png', 'Sora 视频渠道协议和模型能力配置截图')" />
                <figcaption>图 7：中转后台的上游渠道选择 Sora，并把对应模型能力设为视频。</figcaption>
              </figure>
              <figure class="guide-shot">
                <img src="/canvas-guide/08-video-canvas.png" alt="无限画布上传图片连接视频生成节点截图" loading="lazy" @click="openGuideImage('/canvas-guide/08-video-canvas.png', '无限画布上传图片连接视频生成节点截图')" />
                <figcaption>图 8：在画布上传图片，输入文本提示词，连接到视频生成节点后开始生成。</figcaption>
              </figure>
            </div>
            <div class="notice"><b>Sora 配置位置</b><span>图 7 的 Sora 是中转后台的上游协议；无限画布渠道仍按中转站实际暴露的 API 格式配置。若中转站把 Sora 封装成 OpenAI 兼容视频接口，画布侧选 OpenAI，并将模型能力设为视频；只有中转站明确提供独立 Sora 下游接口时，才需要画布侧增加对应调用格式。</span></div>
          </section>

          <section id="network" class="guide-section">
            <div class="section-heading"><span class="section-index">05</span><div><h2>GitHub 与网络准备</h2><p>网络工具只解决浏览器访问 GitHub 等站点的问题，不要和中转 API 渠道混为一谈。</p></div></div>
            <div class="network-options">
              <div><b>推荐：使用可访问 GitHub 的梯子/代理</b><p>在电脑系统或浏览器中启用你平时使用的代理软件，并确认浏览器能打开 GitHub。它主要用于加载 GitHub 上的提示词源等网页资源；提示词列表打不开时，手动输入提示词仍可生图。</p><strong>不要把梯子节点地址粘贴到无限画布「本地代理」设置中。</strong></div>
              <div><b>备选：配置 GitHub 映射</b><p>如果不能使用梯子，可在「配置 → 提示词来源」中为 GitHub 提示词源填写可访问的映射地址（若当前版本/来源提供该选项）。映射只针对相应资源地址，映射失效时可换映射或改用手动提示词。</p></div>
            </div>
            <div class="notice"><b>不要混淆三类地址</b><span>①「渠道」里的中转 API 地址：用于拉取模型和发送生图/视频请求。②「本地代理」：这是另行运行的画布专用转发服务，不是你的中转站节点，也不等于系统梯子；开启后，拉模型、生图、生视频、文本/音频和 WebDAV 请求会先经过该本机服务再转发到渠道地址。如果不打算让这些请求经本地转发，请保持关闭。③ 系统/浏览器梯子或 GitHub 映射：主要用于解决浏览器访问 GitHub 的网络问题。</span></div>
          </section>

          <section id="channels" class="guide-section">
            <div class="section-heading">
              <span class="section-index">06</span>
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
                  兼容格式选择；如果中转站直接提供 Sora 视频接口，也可以在画布渠道里选择
                  Sora。
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
                  格式」混为一项。当前画布渠道编辑器提供 OpenAI、Gemini、Sora
                  三种格式；Sora 格式用于 `/v1/videos` 创建任务并轮询结果。
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
              <span class="section-index">07</span>
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
              <span class="section-index">08</span>
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
                  检查接口地址是否为你的中转 API 地址、API Key
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
                <summary>GitHub 报错或提示词库空白</summary>
                <p>
                  这是提示词资源访问失败，不等于中转 API 生图失败。可启用系统/浏览器梯子访问 GitHub（推荐），或在「配置 → 提示词来源」编辑对应来源 URL 使用可访问的 GitHub 映射；也可以跳过提示词库，在工作台直接输入提示词。不要为此把中转站地址填进「本地代理」。
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
              <span class="section-index">09</span>
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
    <div v-if="lightbox" class="guide-lightbox" role="dialog" aria-modal="true" :aria-label="lightbox.alt" @click.self="closeGuideImage">
      <button class="guide-lightbox-close" type="button" aria-label="关闭图片预览" @click="closeGuideImage">×</button>
      <img :src="lightbox.src" :alt="lightbox.alt" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import AppLayout from "@/components/layout/AppLayout.vue";

const lightbox = ref<{ src: string; alt: string } | null>(null);

function openGuideImage(src: string, alt: string) {
  lightbox.value = { src, alt };
}

function closeGuideImage() {
  lightbox.value = null;
}

function handleGuideKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") closeGuideImage();
}

onMounted(() => window.addEventListener("keydown", handleGuideKeydown));
onUnmounted(() => window.removeEventListener("keydown", handleGuideKeydown));
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
.screen-map {
  overflow: hidden;
  margin-top: 22px;
  border: 1px solid #d9e2e1;
  border-radius: 5px;
  background: #f5f7f7;
  color: #293638;
  font-size: 11px;
}
.guide-shot,
.guide-shot-grid {
  margin: 22px 0 0;
}
.guide-shot {
  overflow: hidden;
  border: 1px solid #d9e2e1;
  border-radius: 5px;
  background: #f7faf9;
}
.guide-shot img {
  display: block;
  width: 100%;
  height: auto;
  cursor: zoom-in;
}
.guide-shot figcaption {
  padding: 10px 12px;
  color: #60706d;
  font-size: 12px;
  line-height: 1.6;
}
.guide-shot-compact {
  max-width: 760px;
}
.guide-shot-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}
.guide-lightbox {
  position: fixed;
  z-index: 1000;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 34px;
  background: #07100ed9;
  cursor: zoom-out;
}
.guide-lightbox img {
  display: block;
  max-width: min(1400px, 96vw);
  max-height: 92vh;
  width: auto;
  height: auto;
  object-fit: contain;
  box-shadow: 0 18px 60px #0008;
  cursor: default;
}
.guide-lightbox-close {
  position: fixed;
  top: 14px;
  right: 18px;
  width: 36px;
  height: 36px;
  border: 1px solid #ffffff66;
  border-radius: 50%;
  background: #111c1acc;
  color: #fff;
  font-size: 26px;
  line-height: 1;
  cursor: pointer;
}
.screen-top,
.screen-tools {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  padding: 10px 13px;
  background: #fff;
  border-bottom: 1px solid #e1e7e6;
}
.screen-top span:last-child {
  color: #788582;
}
.screen-body {
  display: grid;
  grid-template-columns: minmax(82px, 0.7fr) minmax(0, 2fr) minmax(115px, 1fr);
  gap: 9px;
  min-height: 190px;
  padding: 10px;
}
.screen-side,
.screen-form,
.screen-result {
  min-width: 0;
  padding: 11px;
  border: 1px solid #dfe6e5;
  border-radius: 4px;
  background: #fff;
  line-height: 2;
}
.screen-side {
  color: #50625f;
}
.screen-side small {
  color: #87928f;
}
.screen-form {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.screen-field {
  min-height: 34px;
  padding: 6px 8px;
  border: 1px solid #e0e6e5;
  border-radius: 3px;
  color: #687673;
  line-height: 1.55;
}
.screen-field.tall {
  min-height: 58px;
}
.screen-field.actions {
  background: #f8f9f9;
}
.screen-button {
  align-self: flex-end;
  margin-top: auto;
  padding: 5px 11px;
  border-radius: 3px;
  background: #087e73;
  color: #fff;
  line-height: 1.5;
}
.screen-result {
  display: flex;
  flex-direction: column;
}
.screen-placeholder {
  display: grid;
  flex: 1;
  min-height: 110px;
  margin-top: 7px;
  place-content: center;
  text-align: center;
  color: #788582;
  background: #f0f3f2;
  line-height: 1.8;
}
.video-placeholder {
  background: #e8eeed;
}
.screen-tools {
  justify-content: flex-start;
  flex-wrap: wrap;
  border-top: 1px solid #e1e7e6;
  border-bottom: 0;
  color: #50625f;
}
.screen-map > p {
  margin: 0;
  padding: 0 12px 10px;
  text-align: right;
  color: #788582;
  font-size: 10px;
}
.network-options {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
}
.network-options > div {
  padding: 14px 0;
  border-top: 2px solid #dce7e4;
}
.network-options p,
.network-options strong {
  font-size: 13px;
  line-height: 1.75;
}
.network-options p {
  margin: 7px 0;
  color: #637177;
}
.network-options strong {
  color: #a33c32;
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
.dark .tip,
.dark .screen-map,
.dark .screen-top,
.dark .screen-tools,
.dark .screen-side,
.dark .screen-form,
.dark .screen-result {
  background: #26312f;
  border-color: #394744;
  color: #d5dfdc;
}
.dark .screen-field,
.dark .screen-placeholder {
  background: #202a29;
  border-color: #394744;
  color: #bfcbc8;
}
.dark .network-options p {
  color: #aab8b5;
}
.dark .screen-map > p {
  color: #aab8b5;
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
  .guide-shot-grid {
    grid-template-columns: 1fr;
  }
  .screen-body {
    grid-template-columns: 68px minmax(0, 1fr);
  }
  .screen-side {
    padding: 7px;
  }
  .screen-result {
    grid-column: 1 / -1;
    min-height: 100px;
  }
  .screen-placeholder {
    min-height: 65px;
  }
  .network-options {
    grid-template-columns: 1fr;
    gap: 6px;
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
