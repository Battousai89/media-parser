<template>
    <n-card class="source-media-card" :title="sourceName">
        <template #header-extra>
            <div class="source-header">
                <div class="source-info">
                    <a
                        v-if="sourceUrl"
                        :href="sourceUrl"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="source-name-link"
                        :title="sourceUrl"
                    >
                        {{ sourceUrl }}
                    </a>
                    <span v-else class="source-name">{{ sourceName }}</span>
                </div>
                <div class="source-stats">
                    <n-tooltip
                        v-if="source.error_message"
                        trigger="hover"
                        :style="{ maxWidth: '300px' }"
                    >
                        <template #trigger>
                            <n-tag
                                :type="statusType"
                                size="small"
                                style="cursor: help"
                            >
                                {{ statusName }}
                            </n-tag>
                        </template>
                        <div class="error-tooltip">
                            {{ source.error_message }}
                        </div>
                    </n-tooltip>
                    <n-tag v-else :type="statusType" size="small">
                        {{ statusName }}
                    </n-tag>
                    <span class="stat-item retry-info">
                        Повторных попыток: {{ source.retry_count }}/{{
                            source.max_retries
                        }}
                    </span>
                    <n-tooltip trigger="hover">
                        <template #trigger>
                            <span class="stat-item">
                                Найдено: <strong>{{ parsedCount }}</strong> /
                                {{ mediaCount }}
                            </span>
                        </template>
                        Найдено медиа из запрошенного количества
                    </n-tooltip>
                </div>
            </div>
        </template>

        <n-data-table
            :columns="columns"
            :data="media"
            v-model:page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            show-size-picker
            :row-key="(row: MediaItem) => row.id"
            size="small"
        />
    </n-card>
</template>

<script setup lang="ts">
import { h, computed, ref } from "vue";
import { NTag, NButton, NTooltip } from "naive-ui";
import type { DataTableColumns, DataTableBaseColumn } from "naive-ui";
import type { MediaItem, RequestSourceItem } from "../types";
import { useDictionariesStore } from "../stores/dictionaries";

const props = defineProps<{
    source: RequestSourceItem;
    media: MediaItem[];
}>();

const emit = defineEmits<{
    download: [id: string];
}>();

const dictStore = useDictionariesStore();

const sourceName = computed(
    () => props.source.source_name || `Source #${props.source.source_id}`,
);
const sourceUrl = computed(() => props.source.base_url || undefined);

const statusType = computed(() => {
    const status = dictStore.getRequestStatusById(props.source.status_id);
    switch (status?.code) {
        case "pending":
            return "warning";
        case "processing":
            return "info";
        case "completed":
            return "success";
        case "failed":
            return "error";
        default:
            return "default";
    }
});

const statusName = computed(() => {
    const status = dictStore.getRequestStatusById(props.source.status_id);
    return status?.name || "Unknown";
});

const parsedCount = computed(() => props.source.parsed_count);
const mediaCount = computed(() => props.source.media_count);

const getMediaTypeName = (mediaTypeId: number): string => {
    const mt = dictStore.getMediaTypeById(mediaTypeId);
    return mt?.name || `Type #${mediaTypeId}`;
};

const getMediaTypeColor = (
    mediaTypeId: number,
): "default" | "info" | "success" | "warning" | "error" | "primary" => {
    const mt = dictStore.getMediaTypeById(mediaTypeId);
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
    return colors[mt?.code || "other"] || "default";
};

const columns: DataTableColumns<MediaItem> = [
    {
        title: "Тип",
        key: "media_type_id",
        width: 120,
        render: (row) => {
            const typeName = getMediaTypeName(row.media_type_id);
            const color = getMediaTypeColor(row.media_type_id);
            return h(
                NTag,
                { type: color, size: "small" },
                { default: () => typeName },
            );
        },
    } as DataTableBaseColumn<MediaItem>,
    {
        title: "URL",
        key: "url",
        ellipsis: { tooltip: true },
        render: (row) =>
            h(
                "a",
                {
                    href: row.url,
                    target: "_blank",
                    rel: "noopener noreferrer",
                    style: "color: #18a058; text-decoration: none;",
                },
                { default: () => row.url },
            ),
    },
    {
        title: "Название",
        key: "title",
        ellipsis: { tooltip: true },
        render: (row) => row.title || "-",
    },
    {
        title: "Действия",
        key: "actions",
        width: 100,
        render: (row) =>
            h(
                NButton,
                {
                    size: "small",
                    type: "primary",
                    onClick: () => emit("download", row.id),
                },
                { default: () => "Скачать" },
            ),
    },
];

const currentPage = ref(1);
const pageSize = ref(20);
</script>

<style scoped>
.source-media-card {
    margin-bottom: 16px;
}

.source-header {
    display: flex;
    flex-direction: column;
    gap: 8px;
    align-items: flex-end;
}

.source-info {
    display: flex;
    align-items: center;
    gap: 8px;
}

.source-name-link {
    font-weight: 600;
    font-size: 14px;
    color: #18a058;
    text-decoration: none;
    max-width: 250px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.source-name-link:hover {
    text-decoration: underline;
}

.source-name {
    font-weight: 600;
    font-size: 14px;
    color: #333;
}

.source-stats {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    justify-content: flex-end;
}

.stat-item {
    font-size: 12px;
    color: #666;
}

.stat-item strong {
    color: #333;
}

.retry-info {
    color: #f59e42;
    font-weight: 500;
}

.error-tooltip {
    white-space: pre-wrap;
    word-break: break-word;
    max-width: 300px;
}
</style>
