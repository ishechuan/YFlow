<template>
  <div class="languages-view">
    <el-card class="box-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <h1 class="page-title">语言管理</h1>
          <el-button type="primary" icon="Plus" @click="openAddDialog"> 添加语言 </el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="languages" style="width: 100%" border stripe>
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="name" label="语言名称" min-width="150" />
        <el-table-column prop="code" label="语言代码" width="120" align="center">
          <template #default="{ row }">
            <el-tag>{{ row.code }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">
              {{ row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_default" label="默认" width="80" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.is_default" color="#67C23A" :size="20"><Check /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link icon="Edit" @click="openEditDialog(row)">
              编辑
            </el-button>
            <el-button type="danger" link icon="Delete" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑语言' : '添加语言'"
      width="500px"
      @closed="resetForm"
    >
      <el-form ref="formRef" :model="formData" :rules="rules" label-width="100px">
        <el-form-item label="语言名称" prop="name">
          <el-input v-model="formData.name" placeholder="例如: 简体中文" />
        </el-form-item>
        <el-form-item label="语言代码" prop="code">
          <el-input v-model="formData.code" placeholder="例如: zh-CN" />
        </el-form-item>
        <el-form-item label="默认语言" prop="is_default">
          <el-switch v-model="formData.is_default" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitting"> 确定 </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import { getLanguages, createLanguage, updateLanguage, deleteLanguage } from '@/services/language'
import type { Language, CreateLanguageRequest } from '@/types/translation'
import type { FormInstance, FormRules } from 'element-plus'

const loading = ref(false)
const languages = ref<Language[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const formData = reactive<CreateLanguageRequest>({
  name: '',
  code: '',
  is_default: false,
})

const rules = reactive<FormRules>({
  name: [{ required: true, message: '请输入语言名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入语言代码', trigger: 'blur' }],
})

const editingId = ref<number | null>(null)

onMounted(() => {
  fetchLanguages()
})

const fetchLanguages = async () => {
  loading.value = true
  try {
    languages.value = await getLanguages()
  } catch {
    ElMessage.error('获取语言列表失败')
  } finally {
    loading.value = false
  }
}

const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString()
}

const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  formData.name = ''
  formData.code = ''
  formData.is_default = false
  editingId.value = null
  isEdit.value = false
}

const openAddDialog = () => {
  isEdit.value = false
  dialogVisible.value = true
}

const openEditDialog = (row: Language) => {
  isEdit.value = true
  editingId.value = row.id
  formData.name = row.name
  formData.code = row.code
  formData.is_default = row.is_default
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (isEdit.value && editingId.value) {
          await updateLanguage(editingId.value, formData)
          ElMessage.success('更新成功')
        } else {
          await createLanguage(formData)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchLanguages()
      } catch (error: unknown) {
        ElMessage.error(error instanceof Error ? error.message : '操作失败')
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleDelete = (row: Language) => {
  ElMessageBox.confirm(`确定要删除语言 "${row.name}" 吗？`, '删除确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await deleteLanguage(row.id)
      ElMessage.success('删除成功')
      fetchLanguages()
    } catch (error: unknown) {
      ElMessage.error(error instanceof Error ? error.message : '删除失败')
    }
  })
}
</script>

<style scoped>
.languages-view {
  background-color: #f3f4f6;
}

.box-card {
  border-radius: 8px;
  border: none;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}
</style>
