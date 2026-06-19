<template>
  <main class="min-h-screen overflow-hidden bg-[#05080d] text-slate-100">
    <section v-if="!authenticated" class="gate-enter relative grid min-h-screen place-items-center px-6">
      <div class="absolute inset-0 opacity-70">
        <div class="absolute left-[18%] top-[14%] h-48 w-48 rounded-full bg-teal-400/10 blur-3xl" />
        <div class="absolute bottom-[18%] right-[20%] h-64 w-64 rounded-full bg-cyan-300/10 blur-3xl" />
      </div>
      <form class="relative w-full max-w-sm border border-white/10 bg-white/[0.045] p-8 shadow-2xl shadow-teal-950/30 backdrop-blur-xl" @submit.prevent="performLogin">
        <div class="mb-10 flex items-center gap-3">
          <div class="grid h-11 w-11 place-items-center border border-teal-300/30 bg-teal-300/10 text-teal-200">
            <ShieldCheck :size="22" />
          </div>
          <div>
            <p class="text-sm uppercase tracking-[0.28em] text-teal-200/70">gofurry-rag</p>
            <h1 class="mt-1 text-xl font-semibold text-white">控制台入口</h1>
          </div>
        </div>
        <label class="block text-sm text-slate-300" for="passcode">唯一口令</label>
        <input id="passcode" v-model="password" class="mt-3 h-12 w-full border border-white/10 bg-black/30 px-4 text-white outline-none transition focus:border-teal-300/70" type="password" autocomplete="current-password" autofocus />
        <button class="mt-6 flex h-12 w-full items-center justify-center gap-2 bg-teal-300 px-4 text-sm font-semibold text-slate-950 transition hover:bg-teal-200 disabled:cursor-not-allowed disabled:opacity-60" :disabled="busy" type="submit">
          <LogIn :size="18" />进入
        </button>
        <p v-if="notice" class="mt-4 text-sm text-rose-300">{{ notice }}</p>
      </form>
    </section>

    <section v-else class="grid min-h-screen grid-cols-[260px_minmax(0,1fr)]">
      <aside class="relative border-r border-white/10 bg-black/20 px-5 py-6 backdrop-blur-xl">
        <div class="mb-10 flex items-center gap-3">
          <div class="grid h-10 w-10 place-items-center border border-teal-300/30 bg-teal-300/10 text-teal-200">
            <Database :size="21" />
          </div>
          <div>
            <strong class="block text-sm tracking-[0.2em] text-white">GOFURRY RAG</strong>
            <span class="text-xs text-slate-500">Knowledge Observatory</span>
          </div>
        </div>
        <nav class="space-y-2">
          <button v-for="item in menuItems" :key="item.key" class="flex h-11 w-full items-center gap-3 border px-3 text-left text-sm transition" :class="activeMenu === item.key ? 'translate-x-1 border-teal-300/35 bg-teal-300/10 text-teal-100' : 'border-transparent text-slate-400 hover:border-white/10 hover:bg-white/[0.045] hover:text-slate-100'" @click="activeMenu = item.key">
            <component :is="item.icon" :size="18" />{{ item.label }}
          </button>
        </nav>
        <div class="absolute bottom-6 left-5 w-[220px]">
          <button class="flex h-10 w-full items-center justify-center gap-2 border border-white/10 text-sm text-slate-300 transition hover:border-rose-300/30 hover:bg-rose-300/10 hover:text-rose-100" @click="performLogout">
            <LogOut :size="17" />退出
          </button>
        </div>
      </aside>

      <section class="thin-scrollbar max-h-screen overflow-y-auto px-8 py-7">
        <header class="mb-7 flex items-end justify-between gap-6">
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-teal-200/60">{{ currentKicker }}</p>
            <h2 class="mt-2 text-3xl font-semibold tracking-tight text-white">{{ currentTitle }}</h2>
          </div>
          <div class="flex items-center gap-3 text-sm text-slate-400">
            <span class="h-2 w-2 rounded-full bg-teal-300 shadow-[0_0_22px_rgba(45,212,191,0.9)]" />{{ healthState.status || 'unknown' }}
          </div>
        </header>

        <transition name="fade-slide" mode="out-in">
          <section v-if="activeMenu === 'overview'" key="overview" class="space-y-7">
            <div class="grid gap-px overflow-hidden border border-white/10 bg-white/10 md:grid-cols-5">
              <MetricCell label="文档" :value="overviewData?.document_total ?? 0" />
              <MetricCell label="Chunks" :value="overviewData?.chunk_total ?? 0" />
              <MetricCell label="已向量化" :value="overviewData?.embedded_chunk_total ?? 0" />
              <MetricCell label="可检索" :value="overviewData?.ready_documents ?? 0" />
              <MetricCell label="失败文档" :value="overviewData?.failed_documents ?? 0" />
            </div>
            <div class="grid gap-6 lg:grid-cols-[1fr_360px]">
              <div class="border border-white/10 bg-white/[0.035] p-6">
                <div class="mb-5 flex items-center gap-2 text-slate-300">
                  <Activity :size="18" class="text-teal-200" />状态分布
                </div>
                <div class="space-y-4">
                  <StatusBar label="待处理" :value="overviewData?.pending_documents ?? 0" :total="overviewData?.document_total ?? 0" tone="bg-slate-400" />
                  <StatusBar label="处理中" :value="overviewData?.processing_documents ?? 0" :total="overviewData?.document_total ?? 0" tone="bg-amber-300" />
                  <StatusBar label="可检索" :value="overviewData?.ready_documents ?? 0" :total="overviewData?.document_total ?? 0" tone="bg-teal-300" />
                  <StatusBar label="失败" :value="overviewData?.failed_documents ?? 0" :total="overviewData?.document_total ?? 0" tone="bg-rose-300" />
                </div>
              </div>
              <div class="border border-white/10 bg-white/[0.035] p-6">
                <div class="mb-6 flex items-center gap-2 text-slate-300">
                  <Database :size="18" class="text-teal-200" />连接信息
                </div>
                <div class="space-y-5">
                  <section class="border border-white/10 bg-black/20 p-4">
                    <div class="mb-3 flex items-center justify-between">
                      <span class="text-sm text-slate-300">数据库</span>
                      <span class="status-pill" :class="healthState.database?.connected ? 'border-teal-300/30 bg-teal-300/10 text-teal-100' : 'border-rose-300/30 bg-rose-300/10 text-rose-100'">{{ healthState.database?.connected ? 'connected' : 'degraded' }}</span>
                    </div>
                    <dl class="space-y-2 text-sm">
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">类型</dt><dd class="text-slate-200">{{ healthState.database?.type || '-' }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">库名</dt><dd class="text-slate-200">{{ healthState.database?.name || '-' }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">地址</dt><dd class="text-right text-slate-200">{{ databaseAddress }}</dd></div>
                    </dl>
                    <p v-if="healthState.database?.error" class="mt-3 text-xs leading-5 text-rose-300">{{ healthState.database.error }}</p>
                  </section>
                  <section class="border border-white/10 bg-black/20 p-4">
                    <div class="mb-3 flex items-center justify-between">
                      <span class="text-sm text-slate-300">Ollama</span>
                      <span class="status-pill" :class="healthState.ollama?.healthy ? 'border-teal-300/30 bg-teal-300/10 text-teal-100' : 'border-rose-300/30 bg-rose-300/10 text-rose-100'">{{ healthState.ollama?.healthy ? 'healthy' : 'degraded' }}</span>
                    </div>
                    <dl class="space-y-2 text-sm">
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">模型</dt><dd class="text-right text-slate-200">{{ healthState.ollama?.model || healthState.embedding_model || '-' }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">维度</dt><dd class="text-slate-200">{{ healthState.ollama?.embed_dim || '-' }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">地址</dt><dd class="text-right text-slate-200">{{ healthState.ollama?.base_url || '-' }}</dd></div>
                    </dl>
                    <p v-if="healthState.ollama?.error" class="mt-3 text-xs leading-5 text-rose-300">{{ healthState.ollama.error }}</p>
                  </section>
                  <section v-if="overviewData?.ollama_queue" class="grid gap-3 border border-white/10 bg-black/20 p-4">
                    <div class="flex items-center justify-between">
                      <span class="text-sm text-slate-300">Ollama 队列</span>
                      <span class="status-pill" :class="ollamaQueueClass">{{ ollamaQueueLabel }}</span>
                    </div>
                    <dl class="space-y-2 text-sm">
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">当前活跃</dt><dd class="text-slate-200">{{ overviewData?.ollama_queue?.active ?? 0 }}/{{ overviewData?.ollama_queue?.max_concurrency ?? 4 }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">问答队列</dt><dd class="text-slate-200">{{ overviewData?.ollama_queue?.queued_query ?? 0 }}/{{ overviewData?.ollama_queue?.query_queue_size ?? 0 }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">入库队列</dt><dd class="text-slate-200">{{ overviewData?.ollama_queue?.queued_ingest ?? 0 }}/{{ overviewData?.ollama_queue?.ingest_queue_size ?? 0 }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">最近等待</dt><dd class="text-slate-200">{{ formatDuration(overviewData?.ollama_queue?.oldest_wait_ms) }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">等待上限</dt><dd class="text-slate-200">{{ overviewData?.ollama_queue?.wait_timeout_seconds ?? 0 }} s</dd></div>
                    </dl>
                    <p class="text-xs leading-5 text-slate-500">rejected {{ overviewData?.ollama_queue?.rejected ?? 0 }}</p>
                  </section>
                  <section class="grid gap-3 border border-white/10 bg-black/20 p-4">
                    <div class="flex items-center justify-between">
                      <span class="text-sm text-slate-300">Worker 状态</span>
                      <span class="status-pill" :class="workerStatusClass">{{ overviewData?.worker_state || 'idle' }}</span>
                    </div>
                    <dl class="space-y-2 text-sm">
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">当前活跃</dt><dd class="text-slate-200">{{ overviewData?.worker_active_workers ?? 0 }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">已处理</dt><dd class="text-slate-200">{{ overviewData?.worker_total_processed ?? 0 }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">已失败</dt><dd class="text-slate-200">{{ overviewData?.worker_total_failed ?? 0 }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">上次耗时</dt><dd class="text-slate-200">{{ formatDuration(overviewData?.worker_last_duration_ms) }}</dd></div>
                      <div class="flex justify-between gap-4"><dt class="text-slate-500">平均耗时</dt><dd class="text-slate-200">{{ formatDuration(overviewData?.worker_average_duration_ms) }}</dd></div>
                    </dl>
                    <div class="border border-white/10 bg-white/[0.03] p-3">
                      <div class="flex items-center justify-between gap-3">
                        <span class="text-sm text-slate-300">最近失败</span>
                        <span class="text-xs text-slate-500">{{ formatDate(overviewData?.worker_recent_error_at) }}</span>
                      </div>
                      <p v-if="overviewData?.worker_recent_error" class="mt-2 text-sm leading-6 text-rose-200">
                        #{{ overviewData?.worker_current_document_id || overviewData?.worker_last_document_id }} {{ overviewData?.worker_recent_error }}
                      </p>
                      <p v-else class="mt-2 text-sm text-slate-500">Worker 当前没有活跃任务。</p>
                    </div>
                  </section>
                  <section class="grid gap-3 border border-white/10 bg-black/20 p-4">
                    <div class="flex items-center justify-between">
                      <span class="text-sm text-slate-300">待处理队列</span>
                      <span class="text-lg font-semibold text-white">{{ overviewData?.queue_documents ?? 0 }}</span>
                    </div>
                    <div class="border border-white/10 bg-white/[0.03] p-3">
                      <div class="flex items-center justify-between gap-3">
                        <span class="text-sm text-slate-300">最近失败</span>
                        <span class="text-xs text-slate-500">{{ formatDate(overviewData?.recent_failure_at) }}</span>
                      </div>
                      <p v-if="overviewData?.recent_failure_message" class="mt-2 text-sm leading-6 text-rose-200">
                        #{{ overviewData?.recent_failed_document_id }} {{ overviewData?.recent_failure_message }}
                      </p>
                      <p v-else class="mt-2 text-sm text-slate-500">当前没有失败中的文档。</p>
                    </div>
                  </section>
                  <p class="text-xs text-slate-500">整体态势每 5 秒自动刷新</p>
                </div>
              </div>
            </div>
          </section>

          <section v-else-if="activeMenu === 'documents'" key="documents" class="space-y-6">
            <div class="flex gap-2 border-b border-white/10">
              <button v-for="tab in documentTabs" :key="tab.key" class="px-4 py-3 text-sm transition" :class="documentTab === tab.key ? 'border-b border-teal-300 text-teal-100' : 'text-slate-500 hover:text-slate-200'" @click="switchDocumentTab(tab.key)">
                {{ tab.label }}
              </button>
            </div>

            <div v-if="documentTab === 'ingest'" class="grid gap-6 xl:grid-cols-[minmax(0,560px)_1fr]">
              <form class="space-y-5 border border-white/10 bg-white/[0.035] p-6" @submit.prevent="submitText">
                <Field label="标题"><input v-model="form.title" class="control" placeholder="gofurry 网站介绍" /></Field>
                <Field label="正文"><textarea v-model="form.content" class="control min-h-56 resize-none py-3" /></Field>
                <button class="flex items-center gap-2 text-sm text-slate-400 transition hover:text-teal-100" type="button" @click="sourceFieldsOpen = !sourceFieldsOpen">
                  <ChevronDown :size="16" class="transition" :class="sourceFieldsOpen ? 'rotate-180 text-teal-200' : ''" />来源信息（可选）
                </button>
                <div v-if="sourceFieldsOpen" class="space-y-4 border border-white/10 bg-black/20 p-4">
                  <p class="text-xs leading-5 text-slate-500">来源信息用于回溯文档来自哪里；纯手动录入可以不填，系统会按 manual 处理。</p>
                  <div class="grid gap-4 md:grid-cols-2">
                    <Field label="来源类型"><input v-model="form.source_type" class="control" placeholder="manual / website / nav / game" /></Field>
                    <Field label="来源 ID"><input v-model="form.source_id" class="control" placeholder="about-page" /></Field>
                  </div>
                  <Field label="URL"><input v-model="form.url" class="control" placeholder="https://example.com/about" /></Field>
                  <div class="grid gap-4 md:grid-cols-2">
                    <Field label="分类"><input v-model="form.category" class="control" placeholder="faq / intro / docs" /></Field>
                    <Field label="语言"><input v-model="form.language" class="control" placeholder="zh-CN / en-US" /></Field>
                  </div>
                  <div class="grid gap-4 md:grid-cols-2">
                    <Field label="作者"><input v-model="form.author" class="control" placeholder="内容负责人" /></Field>
                    <Field label="发布时间"><input v-model="form.published_at" class="control" placeholder="2026-05-10" /></Field>
                  </div>
                  <Field label="标签"><input v-model="form.tags" class="control" placeholder="多个标签用逗号分隔" /></Field>
                </div>
                <button class="primary-button" :disabled="busy" type="submit"><Send :size="17" />提交入库</button>
              </form>

              <section class="border border-white/10 bg-black/10 p-6">
                <input ref="fileInput" class="hidden" type="file" multiple :accept="fileAccept" @change="onFileInputChange" />
                <div class="drop-zone" :class="dragActive ? 'border-teal-300/60 bg-teal-300/10' : 'border-white/10 bg-white/[0.025]'" @dragenter.prevent="dragActive = true" @dragover.prevent="dragActive = true" @dragleave.prevent="dragActive = false" @drop.prevent="onFileDrop">
                  <UploadCloud :size="30" class="text-teal-200" />
                  <div>
                    <p class="text-base font-medium text-white">拖入文件或选择文件</p>
                    <p class="mt-2 max-w-xl text-sm leading-6 text-slate-500">单文件最大 10 MiB；支持 txt、md、csv、json、yaml、log、html。文件名去掉后缀作为标题，内容作为正文。</p>
                  </div>
                  <button class="ghost-button" type="button" @click="fileInput?.click()">导入文件</button>
                </div>

                <div v-if="fileIssues.length" class="mt-4 border border-amber-300/20 bg-amber-300/10 p-3">
                  <div class="mb-2 flex items-center gap-2 text-sm text-amber-100"><AlertTriangle :size="15" />被跳过的文件</div>
                  <ul class="space-y-1 text-xs leading-5 text-amber-100/80">
                    <li v-for="issue in fileIssues" :key="issue">{{ issue }}</li>
                  </ul>
                </div>

                <div class="mt-6 border border-white/10">
                  <div class="flex items-center justify-between border-b border-white/10 px-4 py-3">
                    <span class="text-sm text-slate-300">待入库文件 {{ pendingFiles.length }}</span>
                    <button class="primary-button h-10" :disabled="busy || pendingFiles.length === 0" type="button" @click="submitFiles"><UploadCloud :size="16" />批量提交入库</button>
                  </div>
                  <div v-if="pendingFiles.length === 0" class="py-12 text-center text-sm text-slate-500">通过拖拽或导入按钮添加文件</div>
                  <ul v-else class="thin-scrollbar max-h-[340px] divide-y divide-white/10 overflow-auto">
                    <li v-for="file in pendingFiles" :key="file.id" class="flex items-center gap-4 px-4 py-3">
                      <FileText :size="18" class="shrink-0 text-teal-200" />
                      <div class="min-w-0 flex-1">
                        <p class="truncate text-sm text-slate-100">{{ file.title }}</p>
                        <p class="mt-1 text-xs text-slate-500">{{ file.name }} · {{ formatBytes(file.size) }}</p>
                      </div>
                      <button class="icon-button" title="移除" type="button" @click="removePendingFile(file.id)"><X :size="15" /></button>
                    </li>
                  </ul>
                </div>
              </section>
            </div>

            <div v-else-if="documentTab === 'list'" class="border border-white/10 bg-white/[0.03]">
              <div class="space-y-4 border-b border-white/10 p-4">
                <div class="grid gap-3 xl:grid-cols-[180px_200px_180px_180px_minmax(0,1fr)_auto]">
                  <div class="relative">
                    <button class="custom-select" type="button" @click="statusOpen = !statusOpen">
                      <span>{{ selectedStatusLabel }}</span><ChevronDown :size="16" :class="statusOpen ? 'rotate-180 text-teal-200' : 'text-slate-500'" />
                    </button>
                    <div v-if="statusOpen" class="absolute left-0 top-12 z-20 w-full border border-white/10 bg-[#090e15] p-1 shadow-2xl shadow-black/40">
                      <button v-for="option in statusOptions" :key="option.value || 'all'" class="flex h-9 w-full items-center justify-between px-3 text-left text-sm transition hover:bg-white/[0.06]" :class="filters.status === option.value ? 'text-teal-100' : 'text-slate-400'" type="button" @click="selectStatus(option.value)">
                        {{ option.label }}<Check v-if="filters.status === option.value" :size="14" />
                      </button>
                    </div>
                  </div>
                  <input v-model="filters.sourceType" class="control h-10" placeholder="来源类型，逗号分隔" @keyup.enter="reloadDocumentsFromFirstPage" />
                  <input v-model="filters.category" class="control h-10" placeholder="分类" @keyup.enter="reloadDocumentsFromFirstPage" />
                  <input v-model="filters.language" class="control h-10" placeholder="语言" @keyup.enter="reloadDocumentsFromFirstPage" />
                  <input v-model="filters.keyword" class="control h-10" placeholder="标题关键字" @keyup.enter="reloadDocumentsFromFirstPage" />
                  <button class="ghost-button" @click="reloadDocumentsFromFirstPage"><RefreshCw :size="16" />刷新</button>
                </div>
                <div class="flex flex-wrap items-center gap-3">
                  <button class="ghost-button" type="button" @click="askBatchReindex"><RotateCcw :size="16" />按当前过滤批量重建</button>
                  <button class="ghost-button" type="button" @click="askRetryFailed"><AlertTriangle :size="16" />重试失败文档</button>
                  <span class="text-xs text-slate-500">批量操作使用状态、来源类型、分类、语言过滤；关键词搜索仅用于列表查看。</span>
                  <span class="ml-auto text-xs text-slate-500">每 3 秒自动刷新</span>
                </div>
              </div>
              <div class="min-h-[452px]">
                <table class="w-full border-collapse text-sm">
                  <thead class="bg-[#080d14] text-left text-xs uppercase tracking-[0.16em] text-slate-500">
                    <tr><th class="px-4 py-3">ID</th><th class="px-4 py-3">标题</th><th class="px-4 py-3">状态</th><th class="px-4 py-3">Chunks</th><th class="px-4 py-3">重试</th><th class="px-4 py-3">索引信息</th><th class="px-4 py-3"></th></tr>
                  </thead>
                  <tbody>
                    <tr v-for="doc in documents.items" :key="doc.id" class="border-t border-white/10 transition hover:bg-white/[0.04]">
                      <td class="px-4 py-4 text-slate-500">#{{ doc.id }}</td>
                      <td class="px-4 py-4">
                        <p class="font-medium text-slate-100">{{ doc.title || 'Untitled' }}</p>
                        <p class="mt-1 text-xs text-slate-500">{{ documentSourceLine(doc) }}</p>
                        <p v-if="documentMetaLine(doc)" class="mt-1 text-xs text-slate-600">{{ documentMetaLine(doc) }}</p>
                      </td>
                      <td class="px-4 py-4"><span class="status-pill" :class="statusClass(doc.status)">{{ statusLabel(doc.status) }}</span><p v-if="doc.error_message" class="mt-2 text-xs text-rose-300">{{ doc.error_message }}</p></td>
                      <td class="px-4 py-4 text-slate-300">{{ doc.chunk_count }}</td>
                      <td class="px-4 py-4 text-slate-300">{{ doc.retry_count }}</td>
                      <td class="px-4 py-4 text-xs leading-6 text-slate-500">
                        <p>完成：{{ formatDate(doc.last_indexed_at || doc.processed_at) }}</p>
                        <p>请求：{{ formatDate(doc.reindex_requested_at) }}</p>
                        <p>失败：{{ formatDate(doc.last_error_at) }}</p>
                      </td>
                      <td class="px-4 py-4">
                        <div class="flex justify-end gap-2">
                          <button class="ghost-button h-9" title="查看 Chunks" @click="openChunksForDocument(doc)"><Layers :size="15" />查看</button>
                          <button class="ghost-button h-9" title="重新索引" @click="askReindexDocument(doc)"><RotateCcw :size="15" />重建</button>
                          <button class="icon-button text-rose-200 hover:border-rose-300/40 hover:bg-rose-300/10" title="删除" @click="askDeleteDocument(doc)"><Trash2 :size="16" /></button>
                        </div>
                      </td>
                    </tr>
                    <tr v-if="documents.items.length === 0"><td class="px-4 py-16 text-center text-sm text-slate-500" colspan="7">暂无文档</td></tr>
                  </tbody>
                </table>
              </div>
              <PaginationBar :buttons="documentPageButtons" :current-page="documentsPage" :total-pages="documentTotalPages" v-model:jump="documentJump" @go="goDocumentPage" @jump="jumpDocumentPage" />
            </div>

            <div v-else class="grid gap-6 xl:grid-cols-[360px_minmax(0,1fr)]">
              <aside class="border border-white/10 bg-white/[0.03]">
                <div class="border-b border-white/10 p-4">
                  <Field label="按文档搜索 Chunks"><input v-model="chunkDocumentKeyword" class="control h-10" placeholder="输入文档标题或 ID" @keyup.enter="searchChunkDocuments(1)" /></Field>
                  <button class="ghost-button mt-3 w-full" type="button" @click="searchChunkDocuments(1)"><Search :size="16" />搜索文档</button>
                </div>
                <div class="thin-scrollbar max-h-[560px] overflow-auto">
                  <button v-for="doc in chunkDocuments.items" :key="doc.id" class="w-full border-b border-white/10 px-4 py-3 text-left transition hover:bg-white/[0.04]" :class="selectedDocument?.id === doc.id ? 'bg-teal-300/10 text-teal-100' : 'text-slate-300'" type="button" @click="openChunksForDocument(doc, false)">
                    <span class="block truncate text-sm font-medium">#{{ doc.id }} {{ doc.title || 'Untitled' }}</span>
                    <span class="mt-1 block text-xs text-slate-500">{{ statusLabel(doc.status) }} · {{ doc.chunk_count }} chunks</span>
                  </button>
                  <p v-if="chunkDocuments.items.length === 0" class="px-4 py-12 text-center text-sm text-slate-500">搜索文档后查看 chunks</p>
                </div>
                <PaginationBar :buttons="chunkDocumentPageButtons" :current-page="chunkDocumentPage" :total-pages="chunkDocumentTotalPages" v-model:jump="chunkDocumentJump" @go="goChunkDocumentPage" @jump="jumpChunkDocumentPage" />
              </aside>
              <section class="border border-white/10 bg-white/[0.03]">
                <div class="flex items-center justify-between border-b border-white/10 p-4">
                  <div>
                    <p class="text-sm text-slate-300">Chunks {{ selectedDocumentLabel }}</p>
                    <p class="mt-1 text-xs text-slate-500">{{ selectedDocument?.title || '请选择一个文档' }}</p>
                  </div>
                  <button class="ghost-button" :disabled="!selectedDocument || operation === 'reload-chunks'" @click="reloadChunks"><RefreshCw :size="16" />刷新</button>
                </div>
                <div v-if="!selectedDocument" class="py-24 text-center text-sm text-slate-500">从左侧选择文档，或在文档页点击“查看”</div>
                <div v-else class="thin-scrollbar max-h-[640px] divide-y divide-white/10 overflow-auto">
                  <article v-for="chunk in chunks.items" :key="chunk.id" class="p-5 transition hover:bg-white/[0.035]">
                    <div class="mb-3 flex items-center justify-between gap-4">
                      <div class="text-xs text-slate-500">#{{ chunk.chunk_index }} · {{ chunk.token_count }} chars · <span :class="chunk.has_embedding ? 'text-teal-200' : 'text-amber-200'">{{ chunk.has_embedding ? chunk.embedding_dim + 'd' : '未向量化' }}</span></div>
                      <div class="flex gap-2">
                        <button v-if="editingChunkId !== chunk.id" class="icon-button" title="编辑" :disabled="operation !== ''" @click="startEditChunk(chunk)"><Pencil :size="15" /></button>
                        <button v-else class="icon-button text-teal-100" title="保存" :disabled="operation === 'save-chunk'" @click="saveChunk(chunk.id)"><Save :size="15" /></button>
                        <button class="icon-button text-rose-200 hover:border-rose-300/40 hover:bg-rose-300/10" title="删除" :disabled="operation !== ''" @click="askDeleteChunk(chunk)"><Trash2 :size="15" /></button>
                      </div>
                    </div>
                    <textarea v-if="editingChunkId === chunk.id" v-model="editingChunkContent" class="control min-h-40 resize-y py-3 leading-6" />
                    <p v-else class="whitespace-pre-wrap break-words text-sm leading-7 text-slate-300">{{ chunk.content }}</p>
                  </article>
                  <p v-if="chunks.items.length === 0" class="py-20 text-center text-sm text-slate-500">这个文档暂时没有 chunks。可以等待入库完成，或重新索引文档。</p>
                </div>
              </section>
            </div>
          </section>

          <section v-else-if="activeMenu === 'sync'" key="sync" class="space-y-6">
            <section class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_auto]">
              <div class="border border-white/10 bg-white/[0.035] p-6">
                <div class="flex flex-wrap items-start justify-between gap-4">
                  <div>
                    <div class="flex items-center gap-2 text-slate-300">
                      <RefreshCw :size="18" class="text-teal-200" />
                      <span>同步调度</span>
                    </div>
                    <p class="mt-2 text-sm leading-6 text-slate-500">
                      服务内定时同步站点导航、游戏详情和游戏新闻。控制台可以查看最近一次结果，也可以手动触发。
                    </p>
                  </div>
                  <div class="flex flex-wrap items-center gap-2">
                    <span class="status-pill" :class="syncEnabledClass">{{ syncEnabledLabel }}</span>
                    <span class="status-pill" :class="syncRunningClass">{{ syncRunningLabel }}</span>
                  </div>
                </div>
                <dl class="mt-5 grid gap-4 sm:grid-cols-3">
                  <div class="border border-white/10 bg-black/20 p-4">
                    <dt class="text-xs uppercase tracking-[0.16em] text-slate-500">自动同步</dt>
                    <dd class="mt-3 text-lg font-semibold text-white">{{ syncState?.enabled ? '已开启' : '已关闭' }}</dd>
                    <p class="mt-2 text-xs text-slate-500">interval {{ syncState?.interval_minutes ?? '-' }} min</p>
                  </div>
                  <div class="border border-white/10 bg-black/20 p-4">
                    <dt class="text-xs uppercase tracking-[0.16em] text-slate-500">当前任务</dt>
                    <dd class="mt-3 text-lg font-semibold text-white">{{ currentSyncSourceLabel }}</dd>
                    <p class="mt-2 text-xs text-slate-500">{{ currentSyncTriggerText }}</p>
                  </div>
                  <div class="border border-white/10 bg-black/20 p-4">
                    <dt class="text-xs uppercase tracking-[0.16em] text-slate-500">开始时间</dt>
                    <dd class="mt-3 text-lg font-semibold text-white">{{ formatDate(syncState?.current_started_at) }}</dd>
                    <p class="mt-2 text-xs text-slate-500">运行中会每 3 秒自动刷新</p>
                  </div>
                </dl>
              </div>
              <div class="flex flex-col gap-3">
                <button class="ghost-button h-11 w-full min-w-[180px]" :disabled="operation === 'sync-refresh'" type="button" @click="refreshSyncStatus">
                  <RefreshCw :size="16" :class="operation === 'sync-refresh' ? 'animate-spin' : ''" />刷新状态
                </button>
                <button class="primary-button h-11 w-full min-w-[180px]" :disabled="syncRunDisabled('all')" type="button" @click="runSyncNow('all')">
                  <RefreshCw :size="16" :class="operation === 'sync:all' ? 'animate-spin' : ''" />立即同步全部
                </button>
              </div>
            </section>

            <section class="grid gap-5 xl:grid-cols-2">
              <article v-for="card in syncCards" :key="card.source" class="border border-white/10 bg-white/[0.03] p-6">
                <div class="flex items-start justify-between gap-4">
                  <div>
                    <div class="flex items-center gap-2">
                      <span class="status-pill border-teal-300/30 bg-teal-300/10 text-teal-100">{{ card.badge }}</span>
                      <strong class="text-base text-white">{{ card.label }}</strong>
                    </div>
                    <p class="mt-3 text-sm leading-6 text-slate-500">{{ card.description }}</p>
                  </div>
                  <span class="status-pill" :class="syncLastRunClass(card.last_run?.status)">{{ syncLastRunLabel(card.last_run?.status) }}</span>
                </div>

                <dl class="mt-5 space-y-3 text-sm">
                  <div class="flex items-center justify-between gap-4 border-b border-white/10 pb-3">
                    <dt class="text-slate-500">来源服务</dt>
                    <dd class="text-right text-slate-200">{{ card.service }}</dd>
                  </div>
                  <div class="flex items-center justify-between gap-4 border-b border-white/10 pb-3">
                    <dt class="text-slate-500">自动同步</dt>
                    <dd class="text-right text-slate-200">{{ card.auto_enabled ? '启用' : '关闭' }}</dd>
                  </div>
                  <div class="flex items-center justify-between gap-4 border-b border-white/10 pb-3">
                    <dt class="text-slate-500">上次同步</dt>
                    <dd class="text-right text-slate-200">{{ formatDate(card.last_run?.completed_at || card.last_run?.started_at) }}</dd>
                  </div>
                  <div class="flex items-center justify-between gap-4 border-b border-white/10 pb-3">
                    <dt class="text-slate-500">触发方式</dt>
                    <dd class="text-right text-slate-200">{{ syncTriggerLabel(card.last_run?.trigger) }}</dd>
                  </div>
                  <div class="flex items-center justify-between gap-4 border-b border-white/10 pb-3">
                    <dt class="text-slate-500">当前文档数</dt>
                    <dd class="text-right text-slate-200">{{ card.current_document_count ?? 0 }}</dd>
                  </div>
                  <div class="flex items-center justify-between gap-4 border-b border-white/10 pb-3">
                    <dt class="text-slate-500">上次扫描总量</dt>
                    <dd class="text-right text-slate-200">{{ card.last_run?.source_total_count ?? 0 }}</dd>
                  </div>
                </dl>

                <div class="mt-5 grid gap-px overflow-hidden border border-white/10 bg-white/10 sm:grid-cols-4">
                  <MetricCell label="新增" :value="card.last_run?.added_count ?? 0" />
                  <MetricCell label="更新" :value="card.last_run?.updated_count ?? 0" />
                  <MetricCell label="跳过" :value="card.last_run?.skipped_count ?? 0" />
                  <MetricCell label="失败" :value="card.last_run?.failed_count ?? 0" />
                </div>

                <p v-if="card.last_run?.message" class="mt-4 border border-amber-300/20 bg-amber-300/10 p-3 text-sm leading-6 text-amber-100">
                  {{ card.last_run?.message }}
                </p>

                <div class="mt-5 flex justify-end">
                  <button class="primary-button h-11" :disabled="syncRunDisabled(card.source)" type="button" @click="runSyncNow(card.source)">
                    <RefreshCw :size="16" :class="operation === `sync:${card.source}` ? 'animate-spin' : ''" />立即同步
                  </button>
                </div>
              </article>
            </section>
          </section>

          <section v-else-if="activeMenu === 'ai'" key="ai" class="space-y-6">
            <AiChatPanel />
          </section>

          <section v-else key="search" class="space-y-6">
            <div class="grid gap-6 xl:grid-cols-[460px_minmax(0,1fr)]">
              <form class="border border-white/10 bg-white/[0.035] p-6" @submit.prevent="runQuery">
                <Field label="问题"><textarea v-model="question" class="control min-h-36 resize-none py-3" /></Field>
                <Field label="Top K"><input v-model="topKText" class="control" inputmode="numeric" pattern="[0-9]*" @input="sanitizeTopK" /></Field>
                <div class="mt-5 space-y-4 border border-white/10 bg-black/20 p-4">
                  <div>
                    <p class="text-xs uppercase tracking-[0.18em] text-slate-500">检索范围</p>
                    <p class="mt-2 text-xs leading-5 text-slate-500">支持按来源类型、分类、语言和文档 ID 精确收窄检索范围。</p>
                  </div>
                  <Field label="来源类型"><input v-model="queryFilters.sourceType" class="control h-10" placeholder="site, faq, manual" /></Field>
                  <div class="grid gap-4 md:grid-cols-2">
                    <Field label="分类"><input v-model="queryFilters.category" class="control h-10" placeholder="intro, faq" /></Field>
                    <Field label="语言"><input v-model="queryFilters.language" class="control h-10" placeholder="zh-CN, en-US" /></Field>
                  </div>
                  <Field label="文档 ID"><input v-model="queryFilters.documentIds" class="control h-10" inputmode="numeric" placeholder="1,2,3" /></Field>
                </div>
                <button class="primary-button mt-5" :disabled="busy" type="submit"><Search :size="17" />检索</button>
              </form>
              <div class="border border-white/10 bg-white/[0.03] p-6">
                <div class="mb-5 flex items-center justify-between gap-4">
                  <div class="flex items-center gap-2 text-slate-300"><BookOpen :size="18" class="text-teal-200" />Sources 调试</div>
                  <span v-if="queryResult" class="text-xs text-slate-500">top_k {{ queryResult.usage.top_k }} / {{ queryResult.usage.embedding_model }}</span>
                </div>
                <div v-if="!queryResult" class="py-20 text-center text-sm text-slate-500">等待检索</div>
                <div v-else class="space-y-4">
                  <p class="text-sm text-slate-400">{{ queryResult.answer }}</p>
                  <article v-for="(source, index) in queryResult.sources" :key="source.chunk_id" class="border border-white/10 bg-black/20 p-4 transition hover:border-teal-300/30">
                    <div class="mb-3 flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <div class="flex flex-wrap items-center gap-2">
                          <span class="status-pill border-teal-300/30 bg-teal-300/10 text-teal-100">#{{ index + 1 }}</span>
                          <strong class="text-sm text-white">{{ source.title || documentLabel(source.document_id) }}</strong>
                        </div>
                        <p class="mt-2 text-xs text-slate-500">{{ sourceDebugLine(source) }}</p>
                      </div>
                      <span class="text-sm font-semibold text-teal-200">{{ source.score.toFixed(4) }}</span>
                    </div>
                    <a v-if="source.url" class="mb-3 block truncate text-xs text-teal-200/80 hover:text-teal-100" :href="source.url" target="_blank" rel="noreferrer">{{ source.url }}</a>
                    <p class="whitespace-pre-wrap break-words border-l border-white/10 pl-4 text-sm leading-7 text-slate-300">{{ source.content }}</p>
                  </article>
                  <div v-if="queryResult.sources.length === 0" class="py-16 text-center text-sm text-slate-500">没有命中 sources</div>
                </div>
              </div>
            </div>

            <section class="border border-white/10 bg-white/[0.03] p-6">
              <div class="mb-5 flex flex-wrap items-center justify-between gap-4">
                <div>
                  <div class="flex items-center gap-2 text-slate-300"><Layers :size="18" class="text-teal-200" />切分预览</div>
                  <p class="mt-2 text-sm text-slate-500">只做本地切分对比，不调用 Ollama，不写入数据库。</p>
                </div>
                <button class="primary-button" :disabled="operation === 'chunk-preview'" type="button" @click="runChunkPreview"><Search :size="16" />生成预览</button>
              </div>
              <div class="grid gap-5 xl:grid-cols-[360px_minmax(0,1fr)]">
                <div class="space-y-4">
                  <Field label="文档 ID"><input v-model="previewDocumentId" class="control" inputmode="numeric" pattern="[0-9]*" placeholder="优先使用已有文档正文" @input="sanitizePreviewDocumentId" /></Field>
                  <Field label="临时文本"><textarea v-model="previewText" class="control min-h-40 resize-y py-3" placeholder="不填文档 ID 时使用这里的文本" /></Field>
                  <div class="border border-white/10 bg-black/20 p-4">
                    <p class="mb-3 text-xs uppercase tracking-[0.18em] text-slate-500">Variants</p>
                    <div v-for="(variant, index) in previewVariants" :key="index" class="mb-3 grid grid-cols-2 gap-3 last:mb-0">
                      <input v-model="variant.chunk_size" class="control h-10" inputmode="numeric" placeholder="chunk_size" @input="sanitizeVariantNumber(variant, 'chunk_size')" />
                      <input v-model="variant.chunk_overlap" class="control h-10" inputmode="numeric" placeholder="overlap" @input="sanitizeVariantNumber(variant, 'chunk_overlap')" />
                    </div>
                  </div>
                </div>
                <div class="min-h-[360px] border border-white/10 bg-black/10">
                  <div v-if="!previewResult" class="py-24 text-center text-sm text-slate-500">输入文档 ID 或临时文本后生成切分预览</div>
                  <div v-else class="thin-scrollbar max-h-[620px] overflow-auto p-4">
                    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p class="text-sm text-white">{{ previewResult.title || (previewResult.source === 'document' ? '未命名文档' : '临时文本') }}</p>
                        <p class="mt-1 text-xs text-slate-500">{{ previewResult.source === 'document' ? '来自文档正文' : '来自临时文本' }}</p>
                      </div>
                      <span class="status-pill border-white/10 bg-white/[0.04] text-slate-300">{{ previewResult.variants.length }} 组参数</span>
                    </div>
                    <div class="space-y-5">
                      <article v-for="variant in previewResult.variants" :key="variant.chunk_size + '-' + variant.chunk_overlap" class="border border-white/10 bg-white/[0.025]">
                        <div class="grid gap-px border-b border-white/10 bg-white/10 sm:grid-cols-4">
                          <MetricCell label="参数" :value="variant.chunk_size" />
                          <MetricCell label="Chunks" :value="variant.chunk_count" />
                          <MetricCell label="最长" :value="variant.max_chars" />
                          <MetricCell label="平均" :value="Math.round(variant.avg_chars)" />
                        </div>
                        <div class="px-4 py-3 text-xs text-slate-500">overlap {{ variant.chunk_overlap }} / min {{ variant.min_chars }} / avg {{ formatAvg(variant.avg_chars) }}</div>
                        <div class="divide-y divide-white/10">
                          <details v-for="chunk in variant.chunks" :key="chunk.index" class="group px-4 py-3">
                            <summary class="cursor-pointer list-none text-sm text-slate-300 transition hover:text-teal-100">
                              <span class="text-teal-200">#{{ chunk.index }}</span>
                              <span class="ml-2 text-slate-500">{{ chunk.char_count }} chars</span>
                            </summary>
                            <p class="mt-3 whitespace-pre-wrap break-words text-sm leading-7 text-slate-400">{{ chunk.content }}</p>
                          </details>
                        </div>
                      </article>
                    </div>
                  </div>
                </div>
              </div>
            </section>
          </section>
        </transition>
        <p v-if="notice" class="fixed bottom-5 right-6 z-40 border border-teal-300/20 bg-black/80 px-4 py-3 text-sm text-teal-100 shadow-xl shadow-black/30">{{ notice }}</p>
      </section>
    </section>

    <div v-if="confirmTarget" class="fixed inset-0 z-50 grid place-items-center bg-black/70 px-6 backdrop-blur-sm">
      <section class="w-full max-w-md border border-white/10 bg-[#090e15] p-6 shadow-2xl shadow-black/50">
        <div class="mb-5 flex items-center gap-3">
          <div class="grid h-10 w-10 place-items-center border text-rose-200" :class="confirmTarget.kind === 'reindex' || confirmTarget.kind === 'batch-reindex' ? 'border-teal-300/30 bg-teal-300/10 text-teal-200' : 'border-rose-300/30 bg-rose-300/10'">
            <component :is="confirmTarget.kind === 'reindex' || confirmTarget.kind === 'batch-reindex' ? RotateCcw : AlertTriangle" :size="20" />
          </div>
          <div>
            <h3 class="text-lg font-semibold text-white">{{ confirmTarget.title }}</h3>
            <p class="mt-1 text-sm text-slate-500">{{ confirmTarget.label }}</p>
          </div>
        </div>
        <p class="text-sm leading-6 text-slate-400">{{ confirmTarget.description }}</p>
        <div class="mt-6 flex justify-end gap-3">
          <button class="ghost-button" type="button" @click="confirmTarget = null">取消</button>
          <button class="danger-button" :class="confirmTarget.kind === 'reindex' || confirmTarget.kind === 'batch-reindex' ? 'reindex-confirm' : ''" type="button" @click="confirmAction">{{ confirmTarget.confirmText }}</button>
        </div>
      </section>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import {
  Activity,
  AlertTriangle,
  BookOpen,
  Check,
  ChevronDown,
  Database,
  FileText,
  Gauge,
  Layers,
  LogIn,
  LogOut,
  Pencil,
  RefreshCw,
  RotateCcw,
  Save,
  Search,
  Send,
  ShieldCheck,
  Sparkles,
  Trash2,
  UploadCloud,
  X,
} from 'lucide-vue-next'
import {
  authState,
  batchReindexDocuments,
  chunkPreview,
  createTextDocument,
  deleteChunk,
  deleteDocument,
  health,
  listChunks,
  listDocuments,
  login,
  logout,
  overview,
  queryRag,
  reindexDocument,
  runSync,
  retryFailedDocuments,
  syncStatus,
  updateChunk,
} from '../api'
import type {
  ChunkItem,
  ChunkPreviewResponse,
  DocumentItem,
  HealthInfo,
  Overview,
  PageResult,
  QueryResponse,
  QuerySource,
  SyncSourceStatus,
  SyncStatusResponse,
} from '../types'
import AiChatPanel from '../components/AiChatPanel.vue'

type MenuKey = 'overview' | 'documents' | 'sync' | 'ai' | 'search'
type DocumentTab = 'ingest' | 'list' | 'chunks'
type SyncSourceKey =
  | 'nav_sites'
  | 'game_details'
  | 'game_news'
  | 'all'
type ConfirmTarget =
  | { kind: 'document'; id: number; title: string; label: string; description: string; confirmText: string }
  | { kind: 'chunk'; id: number; title: string; label: string; description: string; confirmText: string }
  | { kind: 'reindex'; id: number; title: string; label: string; description: string; confirmText: string }
  | { kind: 'batch-reindex'; title: string; label: string; description: string; confirmText: string }
  | { kind: 'batch-retry'; title: string; label: string; description: string; confirmText: string }
type PendingFile = { id: string; name: string; title: string; size: number; type: string; lastModified: number; content: string }
type PreviewVariantForm = { chunk_size: string; chunk_overlap: string }

const maxFileSize = 10 * 1024 * 1024
const allowedExtensions = ['.txt', '.md', '.csv', '.json', '.yaml', '.yml', '.log', '.html', '.htm']
const fileAccept = `${allowedExtensions.join(',')},text/*`

const MetricCell = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: Number, required: true },
  },
  setup(props) {
    return () =>
      h('div', { class: 'bg-[#080d14] p-5' }, [
        h('p', { class: 'text-xs uppercase tracking-[0.18em] text-slate-500' }, props.label),
        h('p', { class: 'mt-3 text-3xl font-semibold text-white' }, String(props.value)),
      ])
  },
})

const StatusBar = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: Number, required: true },
    total: { type: Number, required: true },
    tone: { type: String, required: true },
  },
  setup(props) {
    return () => {
      const pct = props.total > 0 ? Math.round((props.value / props.total) * 100) : 0
      return h('div', [
        h('div', { class: 'mb-2 flex justify-between text-sm' }, [
          h('span', { class: 'text-slate-400' }, props.label),
          h('span', { class: 'text-slate-200' }, `${props.value} ${pct}%`),
        ]),
        h('div', { class: 'h-1.5 bg-white/10' }, [
          h('div', { class: `h-full ${props.tone}`, style: { width: `${pct}%` } }),
        ]),
      ])
    }
  },
})

const Field = defineComponent({
  props: {
    label: { type: String, required: true },
  },
  setup(props, { slots }) {
    return () =>
      h('label', { class: 'block' }, [
        h('span', { class: 'mb-2 block text-sm text-slate-400' }, props.label),
        slots.default?.(),
      ])
  },
})

const PaginationBar = defineComponent({
  props: {
    buttons: { type: Array<(number | string)>, required: true },
    currentPage: { type: Number, required: true },
    totalPages: { type: Number, required: true },
    jump: { type: String, required: true },
  },
  emits: ['go', 'jump', 'update:jump'],
  setup(props, { emit }) {
    return () =>
      h('div', { class: 'flex flex-wrap items-center justify-between gap-4 border-t border-white/10 px-4 py-4' }, [
        h('div', { class: 'flex flex-wrap gap-2' }, props.buttons.map((button, index) => {
          if (button === '...') {
            return h('span', { key: `ellipsis-${index}`, class: 'grid h-9 w-9 place-items-center text-slate-600' }, '...')
          }
          const page = Number(button)
          return h(
            'button',
            {
              key: page,
              class: page === props.currentPage ? 'page-button border-teal-300/40 bg-teal-300/10 text-teal-100' : 'page-button border-white/10 text-slate-400 hover:border-white/20 hover:text-slate-100',
              type: 'button',
              onClick: () => emit('go', page),
            },
            String(page),
          )
        })),
        h('div', { class: 'flex items-center gap-2 text-sm text-slate-500' }, [
          h('span', `共 ${props.totalPages} 页`),
          h('input', {
            class: 'control h-9 w-20',
            inputmode: 'numeric',
            value: props.jump,
            placeholder: '页码',
            onInput: (event: Event) => emit('update:jump', (event.target as HTMLInputElement).value.replace(/\D/g, '')),
            onKeyup: (event: KeyboardEvent) => {
              if (event.key === 'Enter') emit('jump')
            },
          }),
          h('button', { class: 'ghost-button h-9', type: 'button', onClick: () => emit('jump') }, '跳转'),
        ]),
      ])
  },
})

const menuItems = [
  { key: 'overview' as MenuKey, label: '整体态势', icon: Gauge },
  { key: 'documents' as MenuKey, label: '文档管理', icon: FileText },
  { key: 'sync' as MenuKey, label: '同步源', icon: RefreshCw },
  { key: 'ai' as MenuKey, label: 'AI 问答', icon: Sparkles },
  { key: 'search' as MenuKey, label: '文档检索', icon: Search },
]
const documentTabs = [
  { key: 'ingest' as DocumentTab, label: '文本入库' },
  { key: 'list' as DocumentTab, label: '文档' },
  { key: 'chunks' as DocumentTab, label: 'Chunks' },
]
const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'pending', label: '待处理' },
  { value: 'processing', label: '处理中' },
  { value: 'ready', label: '可检索' },
  { value: 'failed', label: '失败' },
]

const authenticated = ref(false)
const password = ref('')
const busy = ref(false)
const operation = ref('')
const notice = ref('')
const activeMenu = ref<MenuKey>('overview')
const documentTab = ref<DocumentTab>('ingest')
const healthState = reactive<HealthInfo>({ status: 'unknown' })
const overviewData = ref<Overview | null>(null)
const syncState = ref<SyncStatusResponse | null>(null)
const documents = reactive<PageResult<DocumentItem>>({ items: [], total: 0 })
const chunks = reactive<PageResult<ChunkItem>>({ items: [], total: 0 })
const chunkDocuments = reactive<PageResult<DocumentItem>>({ items: [], total: 0 })
const selectedDocument = ref<DocumentItem | null>(null)
const filters = reactive({ status: '', keyword: '', sourceType: '', category: '', language: '' })
const queryFilters = reactive({ sourceType: '', category: '', language: '', documentIds: '' })
const queryResult = ref<QueryResponse | null>(null)
const question = ref('gofurry 是个公益网站吗？')
const topKText = ref('6')
const previewDocumentId = ref('')
const previewText = ref('')
const previewResult = ref<ChunkPreviewResponse | null>(null)
const previewVariants = reactive<PreviewVariantForm[]>([
  { chunk_size: '500', chunk_overlap: '80' },
  { chunk_size: '700', chunk_overlap: '120' },
  { chunk_size: '900', chunk_overlap: '150' },
])
const sourceFieldsOpen = ref(false)
const statusOpen = ref(false)
const dragActive = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const pendingFiles = ref<PendingFile[]>([])
const fileIssues = ref<string[]>([])
const documentsPage = ref(1)
const documentJump = ref('')
const chunkDocumentPage = ref(1)
const chunkDocumentJump = ref('')
const chunkDocumentKeyword = ref('')
const editingChunkId = ref<number | null>(null)
const editingChunkContent = ref('')
const confirmTarget = ref<ConfirmTarget | null>(null)
const form = reactive({
  title: '',
  source_type: 'manual',
  source_id: '',
  url: '',
  content: '',
  category: '',
  language: '',
  tags: '',
  author: '',
  published_at: '',
})
let documentPoll: number | undefined
let overviewPoll: number | undefined
let syncPoll: number | undefined

const syncSourceMeta: Record<
  Exclude<SyncSourceKey, 'all'>,
  { label: string; badge: string; description: string; service: string }
> = {
  nav_sites: {
    label: '导航站点',
    badge: 'NAV',
    description: '按中英双语拉取导航站点、分组、详情与可选页面描述，适合回答站点是什么、属于什么分类。',
    service: 'gofurry-nav-backend',
  },
  game_details: {
    label: '游戏详情',
    badge: 'GAME',
    description: '按中英双语拉取游戏详情、标签、分组与官网信息，适合回答某个游戏是什么、支持什么平台。',
    service: 'gofurry-game-backend',
  },
  game_news: {
    label: '游戏新闻',
    badge: 'NEWS',
    description: '同步游戏更新公告与新闻正文，适合回答最近有什么动态、某个游戏更新了什么。',
    service: 'gofurry-game-backend',
  },
}

const databaseAddress = computed(() => {
  const host = healthState.database?.host
  const port = healthState.database?.port
  if (!host && !port) return '-'
  return port ? String(host) + ':' + port : String(host)
})
const selectedDocumentLabel = computed(() => (selectedDocument.value ? '#' + selectedDocument.value.id : ''))
const selectedStatusLabel = computed(() => statusOptions.find((item) => item.value === filters.status)?.label || '全部状态')
const documentTotalPages = computed(() => Math.max(1, Math.ceil(Number(documents.total || 0) / 6)))
const documentPageButtons = computed(() => buildPageButtons(documentsPage.value, documentTotalPages.value))
const chunkDocumentTotalPages = computed(() => Math.max(1, Math.ceil(Number(chunkDocuments.total || 0) / 7)))
const chunkDocumentPageButtons = computed(() => buildPageButtons(chunkDocumentPage.value, chunkDocumentTotalPages.value))
const currentTitle = computed(() => {
  if (activeMenu.value === 'documents') return '文档管理'
  if (activeMenu.value === 'sync') return '同步源'
  if (activeMenu.value === 'ai') return 'AI 问答'
  if (activeMenu.value === 'search') return '文档检索'
  return '整体态势'
})
const currentKicker = computed(() => {
  if (activeMenu.value === 'documents') return 'INGEST / DOCUMENTS / CHUNKS'
  if (activeMenu.value === 'sync') return 'SYNC / NAV / GAME'
  if (activeMenu.value === 'ai') return 'RAG / CHAT / TENCENT'
  if (activeMenu.value === 'search') return 'RETRIEVAL'
  return 'OBSERVABILITY'
})
const syncEnabledClass = computed(() => (syncState.value?.enabled ? 'border-teal-300/30 bg-teal-300/10 text-teal-100' : 'border-slate-400/20 bg-slate-400/10 text-slate-300'))
const syncEnabledLabel = computed(() => (syncState.value?.enabled ? 'auto on' : 'auto off'))
const syncRunningClass = computed(() => {
  if (syncState.value?.running) {
    return 'border-amber-300/30 bg-amber-300/10 text-amber-100'
  }
  return 'border-slate-400/20 bg-slate-400/10 text-slate-300'
})
const syncRunningLabel = computed(() => (syncState.value?.running ? 'running' : 'idle'))
const currentSyncSourceLabel = computed(() => syncSourceLabel(syncState.value?.current_source) || '暂无任务')
const currentSyncTriggerText = computed(() => {
  if (!syncState.value?.running) {
    return '当前没有进行中的同步任务'
  }
  return `${syncTriggerLabel(syncState.value.current_trigger)} · ${syncSourceLabel(syncState.value.current_source)}`
})
const syncCards = computed(() => {
  const sourceMap = new Map<string, SyncSourceStatus>((syncState.value?.sources || []).map((item) => [item.source, item]))
  return (
    Object.entries(syncSourceMeta) as Array<
      [Exclude<SyncSourceKey, 'all'>, { label: string; badge: string; description: string; service: string }]
    >
  ).map(([source, meta]) => {
    const sourceState = sourceMap.get(source)
    return {
      source,
      ...meta,
      service: sourceState?.service || meta.service,
      auto_enabled: sourceState?.auto_enabled ?? Boolean(syncState.value?.enabled),
      current_document_count: sourceState?.current_document_count ?? 0,
      last_run: sourceState?.last_run,
    }
    })
})

async function performLogin() {
  busy.value = true
  notice.value = ''
  try {
    await login(password.value)
    authenticated.value = true
    password.value = ''
    await loadInitialData()
  } catch (error) {
    notifyError(error)
  } finally {
    busy.value = false
  }
}

async function performLogout() {
  await logout()
  authenticated.value = false
  stopDocumentPolling()
  stopOverviewPolling()
  stopSyncPolling()
  activeMenu.value = 'overview'
  documents.items = []
  chunks.items = []
  chunkDocuments.items = []
  selectedDocument.value = null
}

async function loadInitialData() {
  await Promise.all([loadHealth(), loadOverview()])
}

async function loadHealth() {
  Object.assign(healthState, await health())
}

async function loadOverview() {
  overviewData.value = await overview()
}

async function loadSyncStatus() {
  syncState.value = await syncStatus()
}

async function refreshSyncStatus() {
  operation.value = 'sync-refresh'
  try {
    await loadSyncStatus()
  } catch (error) {
    notifyError(error)
  } finally {
    operation.value = ''
  }
}

async function runSyncNow(source: SyncSourceKey) {
  operation.value = `sync:${source}`
  notice.value = ''
  try {
    await runSync(source)
    notice.value = source === 'all' ? '已提交全部同步任务。' : `已提交 ${syncSourceLabel(source)} 同步任务。`
    await Promise.all([loadSyncStatus(), loadOverview()])
  } catch (error) {
    notifyError(error)
  } finally {
    operation.value = ''
  }
}

async function loadDocuments(page = documentsPage.value) {
  documentsPage.value = page
  const result = await listDocuments({
    page: documentsPage.value,
    page_size: 6,
    status: filters.status,
    keyword: filters.keyword,
    source_type: parseCSVInput(filters.sourceType),
    category: filters.category.trim(),
    language: filters.language.trim(),
  })
  documents.items = result.items
  documents.total = result.total
  const maxPage = Math.max(1, Math.ceil(Number(result.total || 0) / 6))
  if (documentsPage.value > maxPage) {
    documentsPage.value = maxPage
    await loadDocuments(maxPage)
    return
  }
  if (selectedDocument.value) {
    const fresh = result.items.find((item) => item.id === selectedDocument.value?.id)
    if (fresh) selectedDocument.value = fresh
  }
}

function reloadDocumentsFromFirstPage() {
  void loadDocuments(1)
}

async function submitText() {
  busy.value = true
  notice.value = ''
  try {
    await createTextDocument({
      title: form.title,
      content: form.content,
      source_type: form.source_type,
      source_id: form.source_id,
      url: form.url,
      metadata: buildMetadataPayload(),
    })
    form.content = ''
    form.title = ''
    notice.value = '文档已提交，等待后台入库。'
    switchDocumentTab('list')
    await Promise.all([loadDocuments(1), loadOverview()])
  } catch (error) {
    notifyError(error)
  } finally {
    busy.value = false
  }
}

async function addFiles(fileList: FileList | File[]) {
  const files = Array.from(fileList)
  const accepted: File[] = []
  const issues: string[] = []
  for (const file of files) {
    const ext = fileExtension(file.name)
    if (!allowedExtensions.includes(ext)) {
      issues.push(`${file.name}：不支持的文件类型`)
      continue
    }
    if (file.size > maxFileSize) {
      issues.push(`${file.name}：超过 10 MiB`)
      continue
    }
    accepted.push(file)
  }
  fileIssues.value = issues
  const loaded = await Promise.all(
    accepted.map(async (file) => ({
      id: `${file.name}-${file.size}-${file.lastModified}-${Math.random().toString(36).slice(2)}`,
      name: file.name,
      title: stripExtension(file.name),
      size: file.size,
      type: file.type,
      lastModified: file.lastModified,
      content: await file.text(),
    })),
  )
  pendingFiles.value = [...pendingFiles.value, ...loaded.filter((item) => item.content.trim())]
}

function onFileDrop(event: DragEvent) {
  dragActive.value = false
  if (event.dataTransfer?.files?.length) void addFiles(event.dataTransfer.files)
}

function onFileInputChange(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files?.length) void addFiles(input.files)
  input.value = ''
}

function removePendingFile(id: string) {
  pendingFiles.value = pendingFiles.value.filter((file) => file.id !== id)
}

async function submitFiles() {
  if (pendingFiles.value.length === 0) return
  busy.value = true
  notice.value = ''
  try {
    for (const file of pendingFiles.value) {
      await createTextDocument({
        title: file.title,
        content: file.content,
        source_type: 'file',
        source_id: file.name,
        url: '',
        metadata: {
          file_name: file.name,
          file_size: file.size,
          file_type: file.type,
          last_modified: file.lastModified,
        },
      })
    }
    const count = pendingFiles.value.length
    pendingFiles.value = []
    notice.value = `已提交 ${count} 个文件入库。`
    switchDocumentTab('list')
    await Promise.all([loadDocuments(1), loadOverview()])
  } catch (error) {
    notifyError(error)
  } finally {
    busy.value = false
  }
}

function switchDocumentTab(tab: DocumentTab) {
  documentTab.value = tab
  if (tab === 'list') void loadDocuments(documentsPage.value)
  if (tab === 'chunks') void searchChunkDocuments(chunkDocumentPage.value)
}

async function openChunksForDocument(doc: DocumentItem, switchTab = true) {
  selectedDocument.value = doc
  if (switchTab) documentTab.value = 'chunks'
  try {
    const result = await listChunks(doc.id, 1, 100)
    chunks.items = result.items
    chunks.total = result.total
  } catch (error) {
    notifyError(error)
  }
}

async function reloadChunks() {
  if (!selectedDocument.value) return
  operation.value = 'reload-chunks'
  try {
    await openChunksForDocument(selectedDocument.value, false)
  } catch (error) {
    notifyError(error)
  } finally {
    operation.value = ''
  }
}

async function searchChunkDocuments(page = 1) {
  chunkDocumentPage.value = page
  const result = await listDocuments({
    page: chunkDocumentPage.value,
    page_size: 7,
    status: '',
    keyword: chunkDocumentKeyword.value.trim(),
    source_type: [],
    category: '',
    language: '',
  })
  chunkDocuments.items = result.items
  chunkDocuments.total = result.total
  const maxPage = Math.max(1, Math.ceil(Number(result.total || 0) / 7))
  if (chunkDocumentPage.value > maxPage) {
    chunkDocumentPage.value = maxPage
    await searchChunkDocuments(maxPage)
  }
}

function askDeleteDocument(doc: DocumentItem) {
  confirmTarget.value = {
    kind: 'document',
    id: doc.id,
    title: '确认删除文档',
    label: `#${doc.id} ${doc.title || 'Untitled'}`,
    description: '删除后文档和所有 chunks 都无法从控制台恢复。',
    confirmText: '删除',
  }
}

function askReindexDocument(doc: DocumentItem) {
  confirmTarget.value = {
    kind: 'reindex',
    id: doc.id,
    title: '确认重新索引',
    label: `#${doc.id} ${doc.title || 'Untitled'}`,
    description: '系统会删除旧 chunks，把文档设为待处理，并由后台 worker 重新切分和向量化。期间该文档会短暂不可检索。',
    confirmText: '重新索引',
  }
}

function askBatchReindex() {
  confirmTarget.value = {
    kind: 'batch-reindex',
    title: '确认批量重新索引',
    label: currentFilterLabel(),
    description: '会删除命中文档的旧 chunks，并将这些文档重新投入后台切分与向量化队列。处理期间这些文档会短暂不可检索。',
    confirmText: '批量重建',
  }
}

function askRetryFailed() {
  confirmTarget.value = {
    kind: 'batch-retry',
    title: '确认重试失败文档',
    label: currentFilterLabel(),
    description: '只会重试命中过滤条件的失败文档，成功提交后会重新进入待处理队列。',
    confirmText: '重试失败',
  }
}

function askDeleteChunk(chunk: ChunkItem) {
  confirmTarget.value = {
    kind: 'chunk',
    id: chunk.id,
    title: '确认删除 Chunk',
    label: `Chunk #${chunk.chunk_index}`,
    description: '删除后这个片段不会再参与检索。需要恢复时可以重新索引所属文档。',
    confirmText: '删除',
  }
}

async function confirmAction() {
  if (!confirmTarget.value) return
  const target = confirmTarget.value
  confirmTarget.value = null
  operation.value = target.kind
  try {
    if (target.kind === 'document') {
      await deleteDocument(target.id)
      if (selectedDocument.value?.id === target.id) {
        selectedDocument.value = null
        chunks.items = []
      }
      notice.value = '文档已删除。'
      await Promise.all([loadDocuments(documentsPage.value), loadOverview(), searchChunkDocuments(chunkDocumentPage.value)])
      return
    }
    if (target.kind === 'reindex') {
      await reindexDocument(target.id)
      notice.value = '文档已提交重新索引。'
      await Promise.all([loadDocuments(documentsPage.value), loadOverview(), searchChunkDocuments(chunkDocumentPage.value)])
      return
    }
    if (target.kind === 'batch-reindex') {
      const result = await batchReindexDocuments(buildBatchRequest())
      notice.value = `已提交批量重建：${result.accepted_count} 个文档进入待处理，跳过 ${result.skipped_count} 个。`
      await Promise.all([loadDocuments(documentsPage.value), loadOverview(), searchChunkDocuments(chunkDocumentPage.value)])
      return
    }
    if (target.kind === 'batch-retry') {
      const result = await retryFailedDocuments(buildBatchRequest())
      notice.value = `已提交失败重试：${result.accepted_count} 个文档进入待处理，跳过 ${result.skipped_count} 个。`
      await Promise.all([loadDocuments(documentsPage.value), loadOverview(), searchChunkDocuments(chunkDocumentPage.value)])
      return
    }
    await deleteChunk(target.id)
    notice.value = 'Chunk 已删除。'
    await Promise.all([reloadChunks(), loadOverview(), searchChunkDocuments(chunkDocumentPage.value)])
  } catch (error) {
    notifyError(error)
  } finally {
    operation.value = ''
  }
}

function startEditChunk(chunk: ChunkItem) {
  editingChunkId.value = chunk.id
  editingChunkContent.value = chunk.content
}

async function saveChunk(id: number) {
  operation.value = 'save-chunk'
  try {
    const updated = await updateChunk(id, editingChunkContent.value)
    const index = chunks.items.findIndex((item) => item.id === id)
    if (index >= 0) chunks.items[index] = updated
    editingChunkId.value = null
    editingChunkContent.value = ''
    notice.value = 'Chunk 已保存并重新向量化。'
  } catch (error) {
    notifyError(error)
  } finally {
    operation.value = ''
  }
}

async function runQuery() {
  busy.value = true
  notice.value = ''
  try {
    const topK = Number(topKText.value || '6')
    queryResult.value = await queryRag(question.value, topK, {
      source_type: parseCSVInput(queryFilters.sourceType),
      category: parseCSVInput(queryFilters.category),
      language: parseCSVInput(queryFilters.language),
      document_ids: parseDocumentIDs(queryFilters.documentIds),
    })
  } catch (error) {
    notifyError(error)
  } finally {
    busy.value = false
  }
}

async function runChunkPreview() {
  operation.value = 'chunk-preview'
  notice.value = ''
  try {
    const documentId = Number(previewDocumentId.value || '0')
    const payload = {
      document_id: documentId > 0 ? documentId : undefined,
      text: documentId > 0 ? undefined : previewText.value,
      variants: previewVariants.map((variant) => ({
        chunk_size: Number(variant.chunk_size || '0'),
        chunk_overlap: Number(variant.chunk_overlap || '0'),
      })),
    }
    previewResult.value = await chunkPreview(payload)
  } catch (error) {
    notifyError(error)
  } finally {
    operation.value = ''
  }
}

function startDocumentPolling() {
  stopDocumentPolling()
  void loadDocuments(documentsPage.value)
  documentPoll = window.setInterval(() => {
    void loadDocuments(documentsPage.value)
  }, 3000)
}

function stopDocumentPolling() {
  if (documentPoll) {
    window.clearInterval(documentPoll)
    documentPoll = undefined
  }
}

function startOverviewPolling() {
  stopOverviewPolling()
  void loadInitialData()
  overviewPoll = window.setInterval(() => {
    void loadInitialData()
  }, 5000)
}

function stopOverviewPolling() {
  if (overviewPoll) {
    window.clearInterval(overviewPoll)
    overviewPoll = undefined
  }
}

function startSyncPolling() {
  stopSyncPolling()
  void loadSyncStatus()
  syncPoll = window.setInterval(() => {
    void loadSyncStatus()
  }, 3000)
}

function stopSyncPolling() {
  if (syncPoll) {
    window.clearInterval(syncPoll)
    syncPoll = undefined
  }
}

function selectStatus(status: string) {
  filters.status = status
  statusOpen.value = false
  reloadDocumentsFromFirstPage()
}

function sanitizeTopK(event: Event) {
  topKText.value = (event.target as HTMLInputElement).value.replace(/\D/g, '').slice(0, 2)
}

function sanitizePreviewDocumentId(event: Event) {
  previewDocumentId.value = (event.target as HTMLInputElement).value.replace(/\D/g, '')
}

function sanitizeVariantNumber(variant: PreviewVariantForm, key: keyof PreviewVariantForm) {
  variant[key] = variant[key].replace(/\D/g, '').slice(0, 4)
}

function goDocumentPage(page: number) {
  void loadDocuments(clampPage(page))
}

function jumpDocumentPage() {
  if (!documentJump.value) return
  void loadDocuments(clampPage(Number(documentJump.value)))
  documentJump.value = ''
}

function goChunkDocumentPage(page: number) {
  void searchChunkDocuments(clampChunkDocumentPage(page))
}

function jumpChunkDocumentPage() {
  if (!chunkDocumentJump.value) return
  void searchChunkDocuments(clampChunkDocumentPage(Number(chunkDocumentJump.value)))
  chunkDocumentJump.value = ''
}

function clampPage(page: number) {
  return Math.min(Math.max(1, page), documentTotalPages.value)
}

function clampChunkDocumentPage(page: number) {
  return Math.min(Math.max(1, page), chunkDocumentTotalPages.value)
}

function buildPageButtons(current: number, total: number) {
  const pages = new Set<number>([1, total, current, current - 2, current - 1, current + 1, current + 2])
  const normalized = Array.from(pages).filter((page) => page >= 1 && page <= total).sort((a, b) => a - b)
  const buttons: Array<number | string> = []
  normalized.forEach((page, index) => {
    const prev = normalized[index - 1]
    if (prev && page - prev > 1) buttons.push('...')
    buttons.push(page)
  })
  return buttons
}

function statusClass(status: string) {
  if (status === 'ready') return 'border-teal-300/30 bg-teal-300/10 text-teal-100'
  if (status === 'processing') return 'border-amber-300/30 bg-amber-300/10 text-amber-100'
  if (status === 'failed') return 'border-rose-300/30 bg-rose-300/10 text-rose-100'
  return 'border-slate-400/20 bg-slate-400/10 text-slate-300'
}

function statusLabel(status: string) {
  return statusOptions.find((item) => item.value === status)?.label || status || '-'
}

function documentSourceLine(doc: DocumentItem) {
  const pieces = [doc.source_type, doc.source_id].filter(Boolean)
  return pieces.length ? pieces.join(' · ') : 'manual'
}

function documentLabel(id: number) {
  return 'Document #' + id
}

function syncSourceLabel(source?: string) {
  if (!source) return ''
  if (source === 'all') return '全部同步源'
  return syncSourceMeta[source as Exclude<SyncSourceKey, 'all'>]?.label || source
}

function syncTriggerLabel(trigger?: string) {
  if (trigger === 'auto') return '自动'
  if (trigger === 'manual') return '手动'
  return trigger || '暂无记录'
}

function syncLastRunLabel(status?: string) {
  if (status === 'success') return 'success'
  if (status === 'partial') return 'partial'
  if (status === 'failed') return 'failed'
  return '未运行'
}

function syncLastRunClass(status?: string) {
  if (status === 'success') return 'border-teal-300/30 bg-teal-300/10 text-teal-100'
  if (status === 'partial') return 'border-amber-300/30 bg-amber-300/10 text-amber-100'
  if (status === 'failed') return 'border-rose-300/30 bg-rose-300/10 text-rose-100'
  return 'border-slate-400/20 bg-slate-400/10 text-slate-300'
}

function syncRunDisabled(source: SyncSourceKey) {
  if (operation.value === `sync:${source}` || operation.value === 'sync-refresh') {
    return true
  }
  if (source === 'all') {
    return Boolean(syncState.value?.running)
  }
  return Boolean(syncState.value?.running)
}

function sourceDebugLine(source: QuerySource) {
  const pieces = [
    `doc #${source.document_id}`,
    `chunk #${source.chunk_index}`,
    `chunk_id ${source.chunk_id}`,
    `${source.token_count || 0} chars`,
    source.source_type,
    source.source_id,
  ].filter(Boolean)
  return pieces.join(' / ')
}

function documentMetaLine(doc: DocumentItem) {
  const metadata = doc.metadata || {}
  const pieces = [metadata.category, metadata.language, metadata.author]
    .map((value) => (typeof value === 'string' ? value : ''))
    .filter(Boolean)
  return pieces.join(' / ')
}

function buildMetadataPayload() {
  const metadata: Record<string, unknown> = {}
  if (form.category.trim()) metadata.category = form.category.trim()
  if (form.language.trim()) metadata.language = form.language.trim()
  const tags = parseCSVInput(form.tags)
  if (tags.length) metadata.tags = tags
  if (form.author.trim()) metadata.author = form.author.trim()
  if (form.published_at.trim()) metadata.published_at = form.published_at.trim()
  return metadata
}

function buildBatchRequest() {
  const filterPayload = {
    source_type: parseCSVInput(filters.sourceType),
    category: parseCSVInput(filters.category),
    language: parseCSVInput(filters.language),
    status: filters.status ? [filters.status] : [],
  }
  if (filterPayload.source_type.length || filterPayload.category.length || filterPayload.language.length || filterPayload.status.length) {
    return {
      scope: 'filters' as const,
      filters: filterPayload,
    }
  }
  return { scope: 'all' as const }
}

function parseCSVInput(value: string) {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

function parseDocumentIDs(value: string) {
  return value
    .split(',')
    .map((item) => Number(item.trim()))
    .filter((item) => Number.isInteger(item) && item > 0)
}

function currentFilterLabel() {
  const pieces = [
    filters.status ? `状态 ${statusLabel(filters.status)}` : '',
    filters.sourceType.trim() ? `来源 ${filters.sourceType.trim()}` : '',
    filters.category.trim() ? `分类 ${filters.category.trim()}` : '',
    filters.language.trim() ? `语言 ${filters.language.trim()}` : '',
  ].filter(Boolean)
  if (pieces.length === 0) {
    return '当前范围：全部文档'
  }
  return `当前范围：${pieces.join(' / ')}`
}

function stripExtension(name: string) {
  return name.replace(/\.[^/.]+$/, '') || name
}

function fileExtension(name: string) {
  const dot = name.lastIndexOf('.')
  return dot >= 0 ? name.slice(dot).toLowerCase() : ''
}

function formatBytes(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

function formatAvg(value: number) {
  return value.toFixed(1)
}

function formatDate(value?: string) {
  if (!value) return '暂无记录'
  return new Date(value).toLocaleString()
}

function formatDuration(value?: number) {
  if (value === undefined || value === null) return '-'
  if (value < 1000) return `${value} ms`
  return `${(value / 1000).toFixed(1)} s`
}

const workerStatusClass = computed(() => {
  if (overviewData.value?.worker_state === 'processing') {
    return 'border-amber-300/30 bg-amber-300/10 text-amber-100'
  }
  if (overviewData.value?.worker_state === 'failed') {
    return 'border-rose-300/30 bg-rose-300/10 text-rose-100'
  }
  return 'border-teal-300/30 bg-teal-300/10 text-teal-100'
})

const ollamaQueueClass = computed(() => {
  const queue = overviewData.value?.ollama_queue
  if (!queue) {
    return 'border-slate-500/30 bg-slate-500/10 text-slate-200'
  }
  if (queue.rejected > 0) {
    return 'border-rose-300/30 bg-rose-300/10 text-rose-100'
  }
  if (queue.active >= queue.max_concurrency || queue.queued_query > 0 || queue.queued_ingest > 0) {
    return 'border-amber-300/30 bg-amber-300/10 text-amber-100'
  }
  return 'border-teal-300/30 bg-teal-300/10 text-teal-100'
})

const ollamaQueueLabel = computed(() => {
  const queue = overviewData.value?.ollama_queue
  if (!queue) {
    return 'idle'
  }
  if (queue.rejected > 0) {
    return 'rejected'
  }
  if (queue.active >= queue.max_concurrency || queue.queued_query > 0 || queue.queued_ingest > 0) {
    return 'busy'
  }
  return 'idle'
})

function notifyError(error: unknown) {
  notice.value = (error as Error).message || '操作失败'
}

watch([activeMenu, documentTab, authenticated], ([menu, tab, loggedIn]) => {
  if (loggedIn && menu === 'overview') {
    startOverviewPolling()
  } else {
    stopOverviewPolling()
  }

  if (loggedIn && menu === 'documents' && tab === 'list') {
    startDocumentPolling()
  } else {
    stopDocumentPolling()
  }

  if (loggedIn && menu === 'sync') {
    startSyncPolling()
  } else {
    stopSyncPolling()
  }
})

onMounted(async () => {
  try {
    const state = await authState()
    authenticated.value = state.authenticated
    if (state.authenticated) {
      await loadInitialData()
      if (activeMenu.value === 'overview') {
        startOverviewPolling()
      }
    }
  } catch {
    authenticated.value = false
  }
})

onUnmounted(() => {
  stopDocumentPolling()
  stopOverviewPolling()
  stopSyncPolling()
})
</script>

<style scoped>
.control {
  width: 100%;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(0, 0, 0, 0.24);
  padding: 0 0.9rem;
  color: white;
  outline: none;
  transition:
    border-color 180ms ease,
    background-color 180ms ease;
}

.control:focus {
  border-color: rgba(94, 234, 212, 0.72);
  background: rgba(0, 0, 0, 0.38);
}

input.control,
select.control {
  height: 2.75rem;
}

.primary-button,
.ghost-button,
.danger-button,
.icon-button,
.custom-select {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  transition:
    border-color 180ms ease,
    background-color 180ms ease,
    color 180ms ease,
    opacity 180ms ease,
    transform 180ms ease;
}

.primary-button {
  height: 2.75rem;
  background: #5eead4;
  padding: 0 1rem;
  font-size: 0.875rem;
  font-weight: 700;
  color: #061015;
}

.primary-button:hover {
  background: #99f6e4;
}

.primary-button:disabled,
.ghost-button:disabled,
.icon-button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.ghost-button {
  height: 2.5rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 0 0.8rem;
  color: #cbd5e1;
}

.ghost-button:hover {
  border-color: rgba(94, 234, 212, 0.38);
  background: rgba(94, 234, 212, 0.08);
  color: #ccfbf1;
}

.danger-button {
  height: 2.5rem;
  border: 1px solid rgba(253, 164, 175, 0.34);
  background: rgba(244, 63, 94, 0.14);
  padding: 0 0.9rem;
  color: #fecdd3;
}

.danger-button:hover {
  background: rgba(244, 63, 94, 0.22);
}

.reindex-confirm {
  border-color: rgba(94, 234, 212, 0.34);
  background: rgba(94, 234, 212, 0.12);
  color: #ccfbf1;
}

.reindex-confirm:hover {
  background: rgba(94, 234, 212, 0.2);
}

.icon-button {
  height: 2rem;
  width: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #cbd5e1;
}

.icon-button:hover {
  border-color: rgba(94, 234, 212, 0.38);
  background: rgba(94, 234, 212, 0.08);
  color: #ccfbf1;
}

.custom-select {
  height: 2.5rem;
  width: 100%;
  justify-content: space-between;
  border: 1px solid rgba(94, 234, 212, 0.18);
  background: linear-gradient(180deg, rgba(94, 234, 212, 0.08), rgba(0, 0, 0, 0.24));
  padding: 0 0.8rem;
  color: #ccfbf1;
}

.drop-zone {
  display: grid;
  min-height: 15rem;
  place-items: center;
  gap: 1rem;
  border: 1px dashed;
  padding: 2rem;
  text-align: center;
  transition:
    border-color 180ms ease,
    background-color 180ms ease,
    transform 180ms ease;
}

.drop-zone:hover {
  transform: translateY(-1px);
}

.status-pill {
  display: inline-flex;
  border-width: 1px;
  padding: 0.22rem 0.5rem;
  font-size: 0.72rem;
}

.page-button {
  display: grid;
  height: 2.25rem;
  min-width: 2.25rem;
  place-items: center;
  border-width: 1px;
  padding: 0 0.65rem;
  font-size: 0.875rem;
  transition:
    border-color 180ms ease,
    background-color 180ms ease,
    color 180ms ease;
}

.fade-slide-enter-active,
.fade-slide-leave-active {
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(6px);
}
</style>
