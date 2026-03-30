<template>
    <n-card class="media-card" size="small">
        <div class="media-preview">
            <n-icon v-if="icon" size="48" :component="icon" />
        </div>
        <div class="media-info">
            <div class="media-title">{{ title || "Без названия" }}</div>
            <div class="media-meta">
                <n-tag size="small" :type="typeTag">
                    {{ mediaTypeName }}
                </n-tag>
                <span v-if="fileSize" class="media-size">
                    {{ formatSize(fileSize) }}
                </span>
            </div>
            <div class="media-url" :title="url">
                {{ url }}
            </div>
        </div>
        <template #action>
            <n-button size="small" @click="$emit('download')">
                <n-icon :component="DownloadOutline" />
                Скачать
            </n-button>
        </template>
    </n-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { NCard, NIcon, NTag, NButton } from "naive-ui";
import type { Component } from "vue";
import {
    ImageOutline,
    VideocamOutline,
    MusicalNotesOutline,
    DocumentTextOutline,
    ArchiveOutline,
    EllipsisHorizontalOutline,
    DownloadOutline,
} from "@vicons/ionicons5";
import { useDictionariesStore } from "../stores/dictionaries";

const props = defineProps<{
    url: string;
    mediaTypeId: number;
    title?: string | null;
    fileSize?: number | null;
}>();

defineEmits<{
    download: [];
}>();

const dictStore = useDictionariesStore();

const mediaType = computed(() => {
    const mt = dictStore.getMediaTypeById(props.mediaTypeId);
    return mt?.code || "other";
});

const mediaTypeName = computed(() => {
    const mt = dictStore.getMediaTypeById(props.mediaTypeId);
    return mt?.name || "Другое";
});

const iconMap: Record<string, Component> = {
    image: ImageOutline,
    video_audio: VideocamOutline,
    audio: MusicalNotesOutline,
    document: DocumentTextOutline,
    archive: ArchiveOutline,
    other: EllipsisHorizontalOutline,
};

const icon = computed(() => iconMap[mediaType.value] || iconMap.other);

const typeTag = computed(() => {
    const colors: Record<
        string,
        "default" | "info" | "success" | "warning" | "error" | "primary"
    > = {
        image: "success",
        video_audio: "info",
        audio: "warning",
        document: "default",
        archive: "error",
        other: "default",
    };
    return colors[mediaType.value] || "default";
});

function formatSize(bytes: number): string {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
}
</script>

<style scoped>
.media-card {
    display: flex;
    flex-direction: column;
}

.media-preview {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 80px;
    background: #f5f5f5;
    border-radius: 4px;
}

.media-info {
    flex: 1;
    padding: 12px 0;
}

.media-title {
    font-weight: 600;
    margin-bottom: 8px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.media-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
}

.media-size {
    font-size: 12px;
    color: #666;
}

.media-url {
    font-size: 12px;
    color: #999;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
