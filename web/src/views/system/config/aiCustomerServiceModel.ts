import type { FormRules } from 'naive-ui';
import { useDictStore } from '@/store/modules/dict';

export function loadOptions() {
  useDictStore().loadOptions(['AiReasoningEffortOptions']);
}

export class AiCustomerServiceState {
  model_base_url = '';
  model_name = '';
  model_reasoning_effort = '';
  model_timeout_seconds: number | null = 30;
  retrieval_limit: number | null = 3;
  embedding_enabled = false;
  embedding_base_url = '';
  embedding_model = '';
  embedding_timeout_seconds: number | null = 30;
  rerank_enabled = false;
  rerank_url = '';
  rerank_model = '';
  rerank_timeout_seconds: number | null = 30;
  retrieval_candidates: number | null = 12;
  vector_min_score: number | null = 0.3;
  rerank_min_score: number | null = 0.1;
}

export const rules: FormRules = {
  model_base_url: {
    required: true,
    trigger: ['blur', 'input'],
    validator: (_rule, value: string) => {
      try {
        const url = new URL(value.trim());
        if (
          ['http:', 'https:'].includes(url.protocol) &&
          url.hostname &&
          !/[\s\u0000-\u001f\u007f]/.test(value.trim()) &&
          !url.username &&
          !url.password &&
          !value.includes('?') &&
          !value.includes('#')
        ) {
          return true;
        }
      } catch {
        // 非法地址统一显示下面的提示。
      }
      return new Error('请输入完整 HTTP(S) 地址，不能包含凭据、查询参数或片段');
    },
  },
  model_name: {
    required: true,
    trigger: ['blur', 'input'],
    validator: (_rule, value: string) =>
      [...value.trim()].length >= 1 && [...value.trim()].length <= 200
        ? true
        : new Error('请输入 1 至 200 个字符的模型名称'),
  },
  model_timeout_seconds: {
    required: true,
    type: 'number',
    trigger: ['blur', 'change'],
    validator: (_rule, value: number | null) =>
      value !== null && Number.isFinite(value) && value > 0 && value <= 120
        ? true
        : new Error('请求超时必须大于 0 且不超过 120 秒'),
  },
  retrieval_limit: {
    required: true,
    type: 'number',
    trigger: ['blur', 'change'],
    validator: (_rule, value: number | null) =>
      value !== null && Number.isInteger(value) && value >= 1 && value <= 10
        ? true
        : new Error('检索片段数必须是 1 至 10 的整数'),
  },
};

export function retrievalRules(current: () => AiCustomerServiceState): FormRules {
  const optionalURL = (enabled: () => boolean) => ({
    trigger: ['blur', 'input'],
    validator: (_rule, value: string) => {
      if (!value.trim()) return enabled() ? new Error('启用时需要填写服务地址') : true;
      try {
        const url = new URL(value.trim());
        if (['http:', 'https:'].includes(url.protocol) && url.hostname && url.port !== '0' && !url.username && !url.password && !/[\s\u0000-\u001f\u007f?#]/.test(value.trim())) return true;
      } catch { /* The backend applies the same validation. */ }
      return new Error('请输入不含凭据、查询参数的完整 HTTP(S) 地址');
    },
  });
  const optionalModel = (enabled: () => boolean) => ({
    trigger: ['blur', 'input'],
    validator: (_rule, value: string) => [...value.trim()].length <= 200 && (!enabled() || !!value.trim()) ? true : new Error('启用时需填写 1 至 200 个字符的模型名称'),
  });
  const score = (min: number) => ({
    trigger: ['blur', 'change'], type: 'number' as const,
    validator: (_rule, value: number | null) => value !== null && Number.isFinite(value) && value >= min && value <= 1 ? true : new Error(`阈值必须在 ${min} 至 1 之间`),
  });
  return {
    ...rules,
    embedding_base_url: optionalURL(() => current().embedding_enabled),
    embedding_model: optionalModel(() => current().embedding_enabled),
    embedding_timeout_seconds: rules.model_timeout_seconds,
    rerank_url: optionalURL(() => current().rerank_enabled),
    rerank_model: optionalModel(() => current().rerank_enabled),
    rerank_timeout_seconds: rules.model_timeout_seconds,
    retrieval_candidates: {
      trigger: ['blur', 'change'], type: 'number',
      validator: (_rule, value: number | null) => value !== null && Number.isInteger(value) && value >= (current().retrieval_limit ?? 1) && value <= 50 ? true : new Error('候选数应不小于最终片段数，且不超过 50'),
    },
    vector_min_score: score(-1),
    rerank_min_score: score(0),
  };
}
