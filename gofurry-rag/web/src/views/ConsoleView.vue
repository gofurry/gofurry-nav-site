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
            <div class="grid gap-px overflow-hidden border border-white/10 bg-white/10 md:grid-cols-4">
              <MetricCell label="文档" :value="overviewData?.document_total ?? 0" />
              <MetricCell label="Chunks" :value="overviewData?.chunk_total ?? 0" />
              <MetricCell label="已向量化" :value="overviewData?.embedded_chunk_total ?? 0" />
              <MetricCell label="可检索" :value="overviewData?.ready_documents ?? 0" />
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
                <Field label="标题"><input v-model="form.title" class="control" placeholder="GoFurry 网站介绍" /></Field>
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
              <div class="flex flex-wrap items-center gap-3 border-b border-white/10 p-4">
                <div class="relative w-44">
                  <button class="custom-select" type="button" @click="statusOpen = !statusOpen">
                    <span>{{ selectedStatusLabel }}</span><ChevronDown :size="16" :class="statusOpen ? 'rotate-180 text-teal-200' : 'text-slate-500'" />
                  </button>
                  <div v-if="statusOpen" class="absolute left-0 top-12 z-20 w-full border border-white/10 bg-[#090e15] p-1 shadow-2xl shadow-black/40">
                    <button v-for="option in statusOptions" :key="option.value || 'all'" class="flex h-9 w-full items-center justify-between px-3 text-left text-sm transition hover:bg-white/[0.06]" :class="filters.status === option.value ? 'text-teal-100' : 'text-slate-400'" type="button" @click="selectStatus(option.value)">
                      {{ option.label }}<Check v-if="filters.status === option.value" :size="14" />
                    </button>
                  </div>
                </div>
                <input v-model="filters.keyword" class="control h-10 w-64" placeholder="标题关键字" @keyup.enter="reloadDocumentsFromFirstPage" />
                <button class="ghost-button" @click="reloadDocumentsFromFirstPage"><RefreshCw :size="16" />刷新</button>
                <span class="ml-auto text-xs text-slate-500">每 3 秒自动刷新</span>
              </div>
              <div class="min-h-[452px]">
                <table class="w-full border-collapse text-sm">
                  <thead class="bg-[#080d14] text-left text-xs uppercase tracking-[0.16em] text-slate-500">
                    <tr><th class="px-4 py-3">ID</th><th class="px-4 py-3">标题</th><th class="px-4 py-3">状态</th><th class="px-4 py-3">Chunks</th><th class="px-4 py-3">更新</th><th class="px-4 py-3"></th></tr>
                  </thead>
                  <tbody>
                    <tr v-for="doc in documents.items" :key="doc.id" class="border-t border-white/10 transition hover:bg-white/[0.04]">
                      <td class="px-4 py-4 text-slate-500">#{{ doc.id }}</td>
                      <td class="px-4 py-4"><p class="font-medium text-slate-100">{{ doc.title || 'Untitled' }}</p><p class="mt-1 text-xs text-slate-500">{{ documentSourceLine(doc) }}</p></td>
                      <td class="px-4 py-4"><span class="status-pill" :class="statusClass(doc.status)">{{ statusLabel(doc.status) }}</span><p v-if="doc.error_message" class="mt-2 text-xs text-rose-300">{{ doc.error_message }}</p></td>
                      <td class="px-4 py-4 text-slate-300">{{ doc.chunk_count }}</td>
                      <td class="px-4 py-4 text-slate-500">{{ formatDate(doc.updated_at) }}</td>
                      <td class="px-4 py-4">
                        <div class="flex justify-end gap-2">
                          <button class="ghost-button h-9" title="查看 Chunks" @click="openChunksForDocument(doc)"><Layers :size="15" />查看</button>
                          <button class="ghost-button h-9" title="重新索引" @click="askReindexDocument(doc)"><RotateCcw :size="15" />重建</button>
                          <button class="icon-button text-rose-200 hover:border-rose-300/40 hover:bg-rose-300/10" title="删除" @click="askDeleteDocument(doc)"><Trash2 :size="16" /></button>
                        </div>
                      </td>
                    </tr>
                    <tr v-if="documents.items.length === 0"><td class="px-4 py-16 text-center text-sm text-slate-500" colspan="6">暂无文档</td></tr>
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

          <section v-else key="search" class="grid gap-6 xl:grid-cols-[520px_1fr]">
            <form class="border border-white/10 bg-white/[0.035] p-6" @submit.prevent="runQuery">
              <Field label="问题"><textarea v-model="question" class="control min-h-36 resize-none py-3" /></Field>
              <Field label="Top K"><input v-model="topKText" class="control" inputmode="numeric" pattern="[0-9]*" @input="sanitizeTopK" /></Field>
              <button class="primary-button mt-5" :disabled="busy" type="submit"><Search :size="17" />检索</button>
            </form>
            <div class="border border-white/10 bg-white/[0.03] p-6">
              <div class="mb-5 flex items-center gap-2 text-slate-300"><BookOpen :size="18" class="text-teal-200" />Sources</div>
              <div v-if="!queryResult" class="py-20 text-center text-sm text-slate-500">等待检索</div>
              <div v-else class="space-y-4">
                <p class="text-slate-300">{{ queryResult.answer }}</p>
                <article v-for="source in queryResult.sources" :key="source.chunk_id" class="border-l border-teal-300/40 bg-black/20 p-4">
                  <div class="mb-2 flex items-center justify-between gap-4"><strong class="text-sm text-white">{{ source.title || documentLabel(source.document_id) }}</strong><span class="text-xs text-teal-200">{{ source.score.toFixed(4) }}</span></div>
                  <p class="whitespace-pre-wrap break-words text-sm leading-6 text-slate-400">{{ source.content }}</p>
                </article>
              </div>
            </div>
          </section>
        </transition>
        <p v-if="notice" class="fixed bottom-5 right-6 z-40 border border-teal-300/20 bg-black/80 px-4 py-3 text-sm text-teal-100 shadow-xl shadow-black/30">{{ notice }}</p>
      </section>
    </section>

    <div v-if="confirmTarget" class="fixed inset-0 z-50 grid place-items-center bg-black/70 px-6 backdrop-blur-sm">
      <section class="w-full max-w-md border border-white/10 bg-[#090e15] p-6 shadow-2xl shadow-black/50">
        <div class="mb-5 flex items-center gap-3">
          <div class="grid h-10 w-10 place-items-center border text-rose-200" :class="confirmTarget.kind === 'reindex' ? 'border-teal-300/30 bg-teal-300/10 text-teal-200' : 'border-rose-300/30 bg-rose-300/10'">
            <component :is="confirmTarget.kind === 'reindex' ? RotateCcw : AlertTriangle" :size="20" />
          </div>
          <div>
            <h3 class="text-lg font-semibold text-white">{{ confirmTarget.title }}</h3>
            <p class="mt-1 text-sm text-slate-500">{{ confirmTarget.label }}</p>
          </div>
        </div>
        <p class="text-sm leading-6 text-slate-400">{{ confirmTarget.description }}</p>
        <div class="mt-6 flex justify-end gap-3">
          <button class="ghost-button" type="button" @click="confirmTarget = null">取消</button>
          <button class="danger-button" :class="confirmTarget.kind === 'reindex' ? 'reindex-confirm' : ''" type="button" @click="confirmAction">{{ confirmTarget.confirmText }}</button>
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
  Trash2,
  UploadCloud,
  X,
} from 'lucide-vue-next'
import {
  authState,
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
  updateChunk,
} from '../api'
import type { ChunkItem, DocumentItem, HealthInfo, Overview, PageResult, QueryResponse } from '../types'

type MenuKey = 'overview' | 'documents' | 'search'
type DocumentTab = 'ingest' | 'list' | 'chunks'
type ConfirmTarget =
  | { kind: 'document'; id: number; title: string; label: string; description: string; confirmText: string }
  | { kind: 'chunk'; id: number; title: string; label: string; description: string; confirmText: string }
  | { kind: 'reindex'; id: number; title: string; label: string; description: string; confirmText: string }
type PendingFile = { id: string; name: string; title: string; size: number; type: string; lastModified: number; content: string }

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
const documents = reactive<PageResult<DocumentItem>>({ items: [], total: 0 })
const chunks = reactive<PageResult<ChunkItem>>({ items: [], total: 0 })
const chunkDocuments = reactive<PageResult<DocumentItem>>({ items: [], total: 0 })
const selectedDocument = ref<DocumentItem | null>(null)
const filters = reactive({ status: '', keyword: '' })
const queryResult = ref<QueryResponse | null>(null)
const question = ref('GoFurry 是个公益网站吗？')
const topKText = ref('6')
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
})
let documentPoll: number | undefined
let overviewPoll: number | undefined

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
  if (activeMenu.value === 'search') return '文档检索'
  return '整体态势'
})
const currentKicker = computed(() => {
  if (activeMenu.value === 'documents') return 'INGEST / DOCUMENTS / CHUNKS'
  if (activeMenu.value === 'search') return 'RETRIEVAL'
  return 'OBSERVABILITY'
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

async function loadDocuments(page = documentsPage.value) {
  documentsPage.value = page
  const result = await listDocuments({ page: documentsPage.value, page_size: 6, status: filters.status, keyword: filters.keyword })
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
    await createTextDocument({ ...form, metadata: {} })
    form.content = ''
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
  const result = await listDocuments({ page: chunkDocumentPage.value, page_size: 7, status: '', keyword: chunkDocumentKeyword.value.trim() })
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
    queryResult.value = await queryRag(question.value, topK)
  } catch (error) {
    notifyError(error)
  } finally {
    busy.value = false
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

function selectStatus(status: string) {
  filters.status = status
  statusOpen.value = false
  reloadDocumentsFromFirstPage()
}

function sanitizeTopK(event: Event) {
  topKText.value = (event.target as HTMLInputElement).value.replace(/\D/g, '').slice(0, 2)
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

function formatDate(value?: string) {
  if (!value) return '暂无记录'
  return new Date(value).toLocaleString()
}

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
