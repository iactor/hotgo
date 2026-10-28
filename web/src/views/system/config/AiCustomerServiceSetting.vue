<template>
  <n-spin :show="loading" description="正在加载客服配置...">
    <n-space vertical :size="20">
      <n-alert type="info" :show-icon="true">
        聊天、Embedding、Rerank 分别配置。API Key 在 Python 客服后端的 MODEL_API_KEY、EMBEDDING_API_KEY、RERANK_API_KEY 中保存。
        参数保存后最多 30 秒应用到新的问答请求。未启用向量和重排序时使用关键词检索。
      </n-alert>
      <n-alert v-if="loadFailed" type="error" title="配置加载失败">
        <n-space align="center">
          <span>请检查服务配置后重试。</span>
          <n-button @click="load">重新加载</n-button>
        </n-space>
      </n-alert>
      <n-form ref="formRef" :model="formValue" :rules="rules" :label-width="120">
        <n-divider title-placement="left">聊天模型 · 理解需求与回答</n-divider>
        <n-form-item label="模型 Base URL" path="model_base_url">
          <n-input v-model:value="formValue.model_base_url" placeholder="https://模型服务地址/v1" />
        </n-form-item>
        <n-form-item label="模型名称" path="model_name">
          <n-input v-model:value="formValue.model_name" placeholder="请输入服务支持的模型名称" />
        </n-form-item>
        <n-form-item label="思考等级" path="model_reasoning_effort">
          <n-select
            v-model:value="formValue.model_reasoning_effort"
            :options="reasoningOptions"
            placeholder="请选择模型支持的思考等级"
          />
        </n-form-item>
        <n-alert type="info">
          默认跟随模型；关闭仅适用于支持关闭思考的模型。不同模型支持的等级不同，较高等级通常增加耗时和用量。
          服务不支持思考等级时请选择默认。此设置同时用于需求判断、工具选择和回答生成。
        </n-alert>
        <n-form-item label="请求超时（秒）" path="model_timeout_seconds">
          <n-input-number v-model:value="formValue.model_timeout_seconds" :max="120" :step="1" />
        </n-form-item>
        <n-divider title-placement="left">Embedding · 向量召回</n-divider>
        <n-form-item label="启用向量检索"><n-switch v-model:value="formValue.embedding_enabled" /></n-form-item>
        <n-form-item label="Embedding 地址" path="embedding_base_url">
          <n-input v-model:value="formValue.embedding_base_url" placeholder="基础地址，例如 https://服务地址/v1；自动追加 /embeddings" />
        </n-form-item>
        <n-form-item label="Embedding 模型" path="embedding_model"><n-input v-model:value="formValue.embedding_model" placeholder="服务支持的向量模型名称" /></n-form-item>
        <n-form-item label="请求超时（秒）" path="embedding_timeout_seconds"><n-input-number v-model:value="formValue.embedding_timeout_seconds" :max="120" :step="1" /></n-form-item>
        <n-alert type="info">切换向量模型或修改知识文件后会按新内容建立缓存。当前适用于小型知识库，首次检索会生成知识向量。</n-alert>
        <n-divider title-placement="left">Rerank · 候选精排</n-divider>
        <n-form-item label="启用重排序"><n-switch v-model:value="formValue.rerank_enabled" /></n-form-item>
        <n-form-item label="Rerank 端点" path="rerank_url"><n-input v-model:value="formValue.rerank_url" placeholder="完整端点，例如 https://服务地址/v1/rerank" /></n-form-item>
        <n-form-item label="Rerank 模型" path="rerank_model"><n-input v-model:value="formValue.rerank_model" placeholder="服务支持的重排序模型名称" /></n-form-item>
        <n-form-item label="请求超时（秒）" path="rerank_timeout_seconds"><n-input-number v-model:value="formValue.rerank_timeout_seconds" :max="120" :step="1" /></n-form-item>
        <n-divider title-placement="left">检索质量参数</n-divider>
        <n-form-item label="召回候选数量" path="retrieval_candidates"><n-input-number v-model:value="formValue.retrieval_candidates" :min="1" :max="50" :precision="0" /></n-form-item>
        <n-form-item label="最终片段数量" path="retrieval_limit">
          <n-input-number v-model:value="formValue.retrieval_limit" :min="1" :max="10" :precision="0" />
        </n-form-item>
        <n-form-item label="向量最低相似度" path="vector_min_score"><n-input-number v-model:value="formValue.vector_min_score" :min="-1" :max="1" :step="0.05" /></n-form-item>
        <n-form-item label="重排序最低分" path="rerank_min_score"><n-input-number v-model:value="formValue.rerank_min_score" :min="0" :max="1" :step="0.05" /></n-form-item>
        <n-alert type="info">阈值需要结合实际模型和测试问题调整；提高阈值会减少返回内容。重排序接口需返回 results，其中包含 index 和 0–1 范围的 relevance_score。</n-alert>
        <n-button
          v-if="hasPermission(['/config/update'])"
          type="primary"
          :loading="saving"
          :disabled="!loaded || loading"
          @click="formSubmit"
        >
          保存更新
        </n-button>
      </n-form>
    </n-space>
  </n-spin>
</template>

<script lang="ts" setup>
  import { onMounted, ref } from 'vue';
  import {
    NAlert,
    NButton,
    NForm,
    NFormItem,
    NInput,
    NInputNumber,
    NSelect,
    NSpace,
    NSpin,
    NDivider,
    NSwitch,
    useMessage,
    type FormInst,
  } from 'naive-ui';
  import { getConfig, updateConfig } from '@/api/sys/config';
  import { usePermission } from '@/hooks/web/usePermission';
  import { useDictStore } from '@/store/modules/dict';
  import { AiCustomerServiceState, loadOptions, retrievalRules } from './aiCustomerServiceModel';

  const group = 'ai_customer_service';
  const reasoningOptions = useDictStore().getOption('AiReasoningEffortOptions');
  const { hasPermission } = usePermission();
  const message = useMessage();
  const formRef = ref<FormInst | null>(null);
  const formValue = ref(new AiCustomerServiceState());
  const rules = retrievalRules(() => formValue.value);
  const loading = ref(false);
  const saving = ref(false);
  const loaded = ref(false);
  const loadFailed = ref(false);

  async function load() {
    loadOptions();
    loading.value = true;
    loadFailed.value = false;
    loaded.value = false;
    try {
      const { list } = await getConfig({ group });
      const next = new AiCustomerServiceState();
      for (const key of Object.keys(next)) {
        if (Object.prototype.hasOwnProperty.call(list, key)) Object.assign(next, { [key]: list[key] });
      }
      formValue.value = next;
      loaded.value = true;
    } catch {
      loadFailed.value = true;
    } finally {
      loading.value = false;
    }
  }

  async function formSubmit() {
    if (!formRef.value || !loaded.value || saving.value) return;
    try {
      await formRef.value.validate();
    } catch {
      return;
    }
    saving.value = true;
    try {
      await updateConfig({ group, list: formValue.value });
      message.success('更新成功，新的问答请求将在 30 秒内使用新配置');
      await load();
    } catch {
      // HTTP 封装已显示服务端错误，保留当前输入以便修正。
    } finally {
      saving.value = false;
    }
  }

  onMounted(load);
</script>
