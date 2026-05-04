<template>
  <main class="shell">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-mark">R</span>
        <div>
          <strong>gofurry-rag</strong>
          <small>Knowledge Console</small>
        </div>
      </div>

      <el-form label-position="top" class="token-form">
        <el-form-item label="Admin Token">
          <el-input v-model="tokenInput" type="password" show-password @change="saveToken" />
        </el-form-item>
        <el-button :icon="Refresh" @click="loadAll">刷新</el-button>
      </el-form>

      <div class="health">
        <span :class="['health-dot', healthState.status === 'ok' ? 'ok' : 'warn']" />
        <span>{{ healthState.status || 'unknown' }}</span>
      </div>
    </aside>

    <section class="workspace">
      <header class="toolbar">
        <div>
          <h1>RAG 控制台</h1>
          <p>文本入库、chunk 检查和检索测试</p>
        </div>
        <div class="filters">
          <el-select v-model="filters.status" placeholder="状态" clearable @change="loadDocuments">
            <el-option label="pending" value="pending" />
            <el-option label="processing" value="processing" />
            <el-option label="ready" value="ready" />
            <el-option label="failed" value="failed" />
          </el-select>
          <el-input v-model="filters.keyword" placeholder="标题关键词" clearable @keyup.enter="loadDocuments" />
          <el-button :icon="Search" type="primary" @click="loadDocuments">查询</el-button>
        </div>
      </header>

      <section class="grid">
        <el-card class="panel input-panel" shadow="never">
          <template #header>文本入库</template>
          <el-form label-position="top">
            <el-form-item label="标题">
              <el-input v-model="form.title" placeholder="GoFurry 网站介绍" />
            </el-form-item>
            <el-form-item label="来源">
              <div class="inline-fields">
                <el-input v-model="form.source_type" placeholder="manual" />
                <el-input v-model="form.source_id" placeholder="about-page" />
              </div>
            </el-form-item>
            <el-form-item label="URL">
              <el-input v-model="form.url" placeholder="https://example.com/about" />
            </el-form-item>
            <el-form-item label="正文">
              <el-input v-model="form.content" type="textarea" :rows="8" resize="none" />
            </el-form-item>
            <el-button :icon="Upload" type="primary" :loading="submitting" @click="submitText">提交入库</el-button>
          </el-form>
        </el-card>

        <el-card class="panel query-panel" shadow="never">
          <template #header>检索测试</template>
          <el-form label-position="top">
            <el-form-item label="问题">
              <el-input v-model="question" type="textarea" :rows="4" resize="none" />
            </el-form-item>
            <el-form-item label="Top K">
              <el-input-number v-model="topK" :min="1" :max="20" />
            </el-form-item>
            <el-button :icon="Cpu" type="success" :loading="querying" @click="runQuery">检索</el-button>
          </el-form>
          <div v-if="queryResult" class="sources">
            <p>{{ queryResult.answer }}</p>
            <article v-for="source in queryResult.sources" :key="source.chunk_id" class="source">
              <strong>{{ source.title || `Document #${source.document_id}` }}</strong>
              <small>score {{ source.score.toFixed(4) }} · chunk {{ source.chunk_id }}</small>
              <p>{{ source.content }}</p>
            </article>
          </div>
        </el-card>
      </section>

      <el-card class="panel table-panel" shadow="never">
        <template #header>文档</template>
        <el-table :data="documents.items" height="360" @row-click="selectDocument">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="title" label="标题" min-width="180" />
          <el-table-column prop="source_type" label="来源" width="120" />
          <el-table-column prop="status" label="状态" width="130">
            <template #default="{ row }">
              <el-tag :type="statusType(row.status)">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="chunk_count" label="Chunks" width="100" />
          <el-table-column prop="error_message" label="错误" min-width="160" />
          <el-table-column label="操作" width="110">
            <template #default="{ row }">
              <el-button :icon="Delete" text type="danger" @click.stop="removeDocument(row.id)" />
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <el-card class="panel table-panel" shadow="never">
        <template #header>Chunks {{ selectedDocument ? `#${selectedDocument.id}` : '' }}</template>
        <el-table :data="chunks.items" height="320">
          <el-table-column prop="chunk_index" label="#" width="70" />
          <el-table-column prop="content" label="内容" min-width="360" show-overflow-tooltip />
          <el-table-column prop="token_count" label="长度" width="90" />
          <el-table-column prop="embedding_dim" label="维度" width="90" />
          <el-table-column prop="has_embedding" label="Embedding" width="120">
            <template #default="{ row }">
              <el-tag :type="row.has_embedding ? 'success' : 'warning'">
                {{ row.has_embedding ? 'ready' : 'missing' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Cpu, Delete, Refresh, Search, Upload } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createTextDocument,
  deleteDocument,
  getToken,
  health,
  listChunks,
  listDocuments,
  queryRag,
  setToken,
} from '../api'
import type { ChunkItem, DocumentItem, PageResult, QueryResponse } from '../types'

const tokenInput = ref(getToken())
const healthState = reactive<Record<string, unknown>>({ status: 'unknown' })
const filters = reactive({ status: '', keyword: '' })
const documents = reactive<PageResult<DocumentItem>>({ items: [], total: 0 })
const chunks = reactive<PageResult<ChunkItem>>({ items: [], total: 0 })
const selectedDocument = ref<DocumentItem | null>(null)
const submitting = ref(false)
const querying = ref(false)
const question = ref('GoFurry 是做什么的？')
const topK = ref(6)
const queryResult = ref<QueryResponse | null>(null)
const form = reactive({
  title: '',
  source_type: 'manual',
  source_id: '',
  url: '',
  content: '',
})

function saveToken() {
  setToken(tokenInput.value)
  ElMessage.success('Token 已保存')
}

async function loadHealth() {
  Object.assign(healthState, await health())
}

async function loadDocuments() {
  const result = await listDocuments({ page: 1, page_size: 50, status: filters.status, keyword: filters.keyword })
  documents.items = result.items
  documents.total = result.total
}

async function loadAll() {
  saveToken()
  try {
    await Promise.all([loadHealth(), loadDocuments()])
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function submitText() {
  submitting.value = true
  try {
    await createTextDocument({ ...form, metadata: {} })
    ElMessage.success('已创建 pending 文档')
    form.content = ''
    await loadDocuments()
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    submitting.value = false
  }
}

async function selectDocument(row: DocumentItem) {
  selectedDocument.value = row
  try {
    const result = await listChunks(row.id)
    chunks.items = result.items
    chunks.total = result.total
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function removeDocument(id: number) {
  await ElMessageBox.confirm(`删除文档 #${id}？`, '确认删除', { type: 'warning' })
  await deleteDocument(id)
  ElMessage.success('已删除')
  chunks.items = []
  await loadDocuments()
}

async function runQuery() {
  querying.value = true
  try {
    queryResult.value = await queryRag(question.value, topK.value)
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    querying.value = false
  }
}

function statusType(status: string) {
  if (status === 'ready') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'processing') return 'warning'
  return 'info'
}

onMounted(loadAll)
</script>
