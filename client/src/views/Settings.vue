<template>
    <div class="settings-page">
        <n-card title="Настройки">
            <n-form :model="form" :rules="rules" ref="formRef" label-placement="top">
                <n-form-item label="API Token" path="apiToken">
                    <n-input
                        v-model:value="form.apiToken"
                        placeholder="Введите ваш API токен"
                        type="password"
                        show-password-on="click"
                    />
                </n-form-item>

                <n-form-item label="Backend URL" path="apiUrl">
                    <n-input
                        v-model:value="form.apiUrl"
                        placeholder="http://localhost:8080"
                    />
                </n-form-item>

                <n-form-item>
                    <n-button type="primary" @click="handleSave" :loading="saving">
                        Сохранить
                    </n-button>
                </n-form-item>
            </n-form>
        </n-card>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { NForm, NFormItem, NInput, NButton, useMessage } from "naive-ui";
import type { FormRules, FormInst } from "naive-ui";
import { useSettingsStore } from "../stores/settings";

const settingsStore = useSettingsStore();
const message = useMessage();
const formRef = ref<FormInst | null>(null);
const saving = ref(false);

const form = reactive({
    apiToken: "",
    apiUrl: "http://localhost:8080",
});

const rules: FormRules = {
    apiToken: {
        required: true,
        message: "Введите API токен",
        trigger: "blur",
    },
    apiUrl: {
        required: true,
        message: "Введите URL бэкенда",
        trigger: "blur",
        validator: (_rule, value) => {
            try {
                new URL(value);
                return true;
            } catch {
                return new Error("Некорректный URL");
            }
        },
    },
};

onMounted(() => {
    form.apiToken = settingsStore.apiToken;
    form.apiUrl = settingsStore.apiUrl;
});

async function handleSave() {
    await formRef.value?.validate();
    saving.value = true;
    try {
        settingsStore.setToken(form.apiToken);
        settingsStore.setUrl(form.apiUrl);
        message.success("Настройки сохранены");
    } catch {
        message.error("Ошибка сохранения");
    } finally {
        saving.value = false;
    }
}
</script>

<style scoped>
.settings-page {
    max-width: 600px;
    margin: 0 auto;
}
</style>
