<template>
    <div class="dashboard">
        <n-card title="Запросы на парсинг" content-style="padding: 16px;">
            <template #header-extra>
                <div class="header-actions">
                    <span class="last-updated"
                        >Обновлено: {{ lastUpdated }}</span
                    >
                    <n-button @click="refresh" :loading="loading">
                        <template #icon>
                            <n-icon
                                :component="Refresh"
                                :class="{ spinning: loading }"
                            />
                        </template>
                        Обновить
                    </n-button>
                </div>
            </template>

            <div class="table-wrapper">
                <n-data-table
                    :columns="columns"
                    :data="requests"
                    :loading="loading"
                    :row-key="(row) => row.id"
                    @click-row="handleRowClick"
                    class="requests-table"
                    :single-line="false"
                    :scroll-x="1400"
                />
            </div>

            <div class="pagination-wrapper" v-if="total > 0">
                <n-pagination
                    v-model:page="currentPage"
                    v-model:page-size="pageSize"
                    :page-count="pageCount"
                    :page-sizes="[10, 20, 50, 100]"
                    show-size-picker
                    @update:page="handlePageChange"
                    @update:page-size="handlePageSizeChange"
                />
            </div>
        </n-card>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, h, computed } from "vue";
import { useRouter } from "vue-router";
import {
    NCard,
    NDataTable,
    NButton,
    NIcon,
    NPagination,
    NTooltip,
    NTag,
    useMessage,
} from "naive-ui";
import type { DataTableColumns } from "naive-ui";
import { Refresh } from "@vicons/ionicons5";
import type { RequestListResponse } from "../types";
import { getRequests } from "../api/endpoints";
import StatusBadge from "../components/StatusBadge.vue";
import { useDictionariesStore } from "../stores/dictionaries";

const router = useRouter();
const message = useMessage();
const dictStore = useDictionariesStore();

const requests = ref<RequestListResponse[]>([]);
const loading = ref(false);
const total = ref(0);
const limit = ref(20);
const offset = ref(0);
const currentPage = ref(1);
const pageSize = ref(20);
const lastUpdated = ref("");

const getMediaTypesNames = computed(() => {
  return (ids: number[]) => {
    return ids.map(id => {
      const mt = dictStore.getMediaTypeById(id);
      return mt?.name || `Type #${id}`;
    });
  };
});

const pageCount = computed(() => {
  const count = Math.ceil(total.value / pageSize.value);
  console.log('pageCount:', count, 'total:', total.value, 'pageSize:', pageSize.value);
  return count;
});

function renderTitle(text: string, tooltip: string) {
    return () =>
        h(
            NTooltip,
            { trigger: "hover" },
            {
                trigger: () => h("span", { style: "cursor: help;" }, text),
                default: () => tooltip,
            }
        );
}

const columns: DataTableColumns<RequestListResponse> = [
    {
        title: "№",
        key: "rowNum",
        width: 60,
        render: (_row, index) =>
            (currentPage.value - 1) * pageSize.value + index + 1,
    },
    {
        title: "UUID",
        key: "id",
        width: 280,
        ellipsis: { tooltip: true },
        sorter: "default",
        render: (row) =>
            h(
                "a",
                {
                    style: "color: #18a058; cursor: pointer;",
                    onClick: (e: Event) => {
                        e.stopPropagation();
                        router.push({
                            name: "RequestDetail",
                            params: { id: String(row.id) },
                        });
                    },
                },
                row.id,
            ),
    },
    {
        title: renderTitle("Статус", "Статус выполнения запроса"),
        key: "status_id",
        width: 120,
        sorter: "default",
        render: (row) => h(StatusBadge, { statusId: row.status_id }),
    },
    {
        title: renderTitle(
            "Типы медиа",
            "Типы медиафайлов для парсинга"
        ),
        key: "media_type_ids",
        width: 180,
        render: (row) => {
            if (!row.media_type_ids || row.media_type_ids.length === 0) {
                return h(NTag, { type: "default", size: "small" }, { default: () => "Все" });
            }
            return h("div", { style: "display: flex; gap: 4px; flex-wrap: wrap;" }, [
                row.media_type_ids.map(id => 
                    h(NTag, { key: id, type: "info", size: "small" }, { default: () => getMediaTypesNames.value([id])[0] })
                )
            ]);
        },
    },
    {
        title: renderTitle(
            "Лимит",
            "Максимальное количество медиафайлов для парсинга"
        ),
        key: "limit_count",
        width: 80,
        sorter: "default",
    },
    {
        title: renderTitle(
            "Смещение",
            "Количество пропущенных медиафайлов с начала"
        ),
        key: "offset_count",
        width: 110,
        sorter: "default",
    },
    {
        title: renderTitle(
            "Приоритет",
            "Приоритет запроса (0-10): чем выше, тем раньше обрабатывается"
        ),
        key: "priority",
        width: 110,
        sorter: "default",
    },
    {
        title: renderTitle(
            "Источников",
            "Количество источников, добавленных к запросу"
        ),
        key: "sources_count",
        width: 110,
        sorter: "default",
    },
    {
        title: renderTitle(
            "Найдено",
            "Общее количество найденных медиафайлов по всем источникам"
        ),
        key: "parsed_count",
        width: 100,
        sorter: "default",
    },
    {
        title: renderTitle(
            "Создан",
            "Дата и время создания запроса"
        ),
        key: "created_at",
        width: 180,
        render: (row) => new Date(row.created_at).toLocaleString(),
        sorter: (a, b) =>
            new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
    },
];

let pollingInterval: number | undefined;

async function fetchRequests() {
    loading.value = true;
    try {
        const data = await getRequests({
            limit: pageSize.value,
            offset: (currentPage.value - 1) * pageSize.value,
        });
        console.log('Fetched requests:', data);
        requests.value = data.items;
        total.value = data.total;
        limit.value = data.limit;
        offset.value = data.offset;
        console.log('total:', total.value, 'pageSize:', pageSize.value, 'currentPage:', currentPage.value);
        lastUpdated.value = new Date().toLocaleTimeString();
    } catch {
        message.error("Ошибка загрузки запросов");
    } finally {
        loading.value = false;
    }
}

function refresh() {
    currentPage.value = 1;
    fetchRequests();
}

function handleRowClick(_e: Event, row: RequestListResponse) {
    console.log("Clicked row:", row.id);
    router.push({ name: "RequestDetail", params: { id: String(row.id) } });
}

function handlePageChange(page: number) {
    currentPage.value = page;
    fetchRequests();
}

function handlePageSizeChange(size: number) {
    pageSize.value = size;
    currentPage.value = 1;
    fetchRequests();
}

onMounted(async () => {
    await dictStore.loadDictionaries();
    fetchRequests();
    pollingInterval = window.setInterval(fetchRequests, 5000);
});

onUnmounted(() => {
    if (pollingInterval) {
        clearInterval(pollingInterval);
    }
});
</script>

<style scoped>
.dashboard {
    max-width: 1200px;
    margin: 0 auto;
}

.header-actions {
    display: flex;
    align-items: center;
    gap: 16px;
}

.last-updated {
    color: #999;
    font-size: 13px;
}

.requests-table :deep(tr) {
    cursor: pointer;
}

.requests-table :deep(tr:hover) {
    background-color: #f5f5f5;
}

.pagination-wrapper {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
}

.spinning {
    animation: spin 1s linear infinite;
}

@keyframes spin {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(360deg);
    }
}
</style>
