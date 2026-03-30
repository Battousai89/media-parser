<template>
  <div class="new-request">
    <n-card title="Новый запрос на парсинг">
      <n-form :model="form" :rules="rules" ref="formRef" label-placement="top">
        <n-form-item label="URL (один в строке, макс. 100)" path="urlsText">
          <n-input
            v-model:value="form.urlsText"
            type="textarea"
            :rows="8"
            placeholder="https://example.com/page1&#10;https://example.com/page2"
          />
        </n-form-item>

        <n-form-item label="Типы медиа" path="mediaTypeIds">
          <MediaTypeSelect v-model:value="form.mediaTypeIds" placeholder="Все типы" />
          <div v-if="form.mediaTypeIds && form.mediaTypeIds.length > 0" class="selected-types">
            <n-tag
              v-for="id in form.mediaTypeIds"
              :key="id"
              size="small"
              style="margin-right: 8px; margin-top: 8px;"
            >
              {{ getMediaTypeName(id) }}
            </n-tag>
          </div>
        </n-form-item>

        <n-form-item label="Лимит" path="limit">
          <n-input-number
            v-model:value="form.limit"
            :min="1"
            :max="1000"
            style="width: 150px"
          />
        </n-form-item>

        <n-form-item label="Offset" path="offset">
          <n-input-number
            v-model:value="form.offset"
            :min="0"
            style="width: 150px"
          />
        </n-form-item>

        <n-form-item label="Приоритет (0-10)" path="priority">
          <div class="priority-slider">
            <n-slider
              v-model:value="form.priority"
              :min="0"
              :max="10"
              :marks="{
                0: 'Низкий',
                5: '',
                10: 'Высокий'
              }"
              :step="1"
              style="width: 100%"
            />
          </div>
          <div class="priority-value">{{ form.priority }}</div>
        </n-form-item>

        <n-form-item>
          <n-button type="primary" @click="handleSubmit" :loading="loading">
            Запустить
          </n-button>
        </n-form-item>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import {
  NCard,
  NForm,
  NFormItem,
  NInputNumber,
  NSlider,
  NButton,
  NTag,
  useMessage,
} from 'naive-ui';
import type { FormRules, FormInst } from 'naive-ui';
import { parseBatch } from '../api/endpoints';
import MediaTypeSelect from '../components/MediaTypeSelect.vue';
import { useDictionariesStore } from '../stores/dictionaries';

const router = useRouter();
const message = useMessage();
const formRef = ref<FormInst | null>(null);
const loading = ref(false);
const dictStore = useDictionariesStore();

const form = reactive({
  urlsText: '',
  mediaTypeIds: [] as number[],
  limit: 10,
  offset: 0,
  priority: 5,
});

const rules: FormRules = {
  urlsText: {
    required: true,
    trigger: 'blur',
    validator: (_rule, value) => {
      const urls = value.split('\n').filter((u: string) => u.trim());
      if (urls.length === 0) return new Error('Введите хотя бы один URL');
      if (urls.length > 100) return new Error('Максимум 100 URL');
      return true;
    },
  },
  limit: {
    required: true,
    type: 'number',
    trigger: 'blur',
  },
  offset: {
    required: true,
    type: 'number',
    trigger: 'blur',
  },
  priority: {
    required: true,
    type: 'number',
    min: 0,
    max: 10,
    trigger: 'blur',
  },
};

const getMediaTypeName = computed(() => {
  return (id: number) => {
    const mt = dictStore.getMediaTypeById(id);
    return mt?.name || `Type #${id}`;
  };
});

async function handleSubmit() {
  await formRef.value?.validate();
  loading.value = true;
  try {
    const urls = form.urlsText.split('\n').filter((u) => u.trim());

    const result = await parseBatch({
      urls,
      media_type_ids: form.mediaTypeIds.length > 0 ? form.mediaTypeIds : undefined,
      limit: form.limit,
      offset: form.offset,
      priority: form.priority,
    });

    message.success('Запрос создан');
    router.push({ name: 'RequestDetail', params: { id: result.request_id } });
  } catch (e) {
    message.error(e instanceof Error ? e.message : 'Ошибка создания запроса');
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  await dictStore.loadDictionaries();
});
</script>

<style scoped>
.new-request {
  max-width: 600px;
  margin: 0 auto;
}

.priority-slider {
  width: 100%;
  padding: 8px 0;
}

.priority-value {
  text-align: center;
  font-weight: 600;
  color: #18a058;
  margin-top: 4px;
}

.selected-types {
  margin-top: 8px;
}
</style>
