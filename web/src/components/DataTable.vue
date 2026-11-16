<template>
  <div class="data-table">
    <table>
      <thead>
        <tr>
          <th v-for="(column, colIndex) in columns" :key="column" :class="getColumnClass(colIndex)">
            {{ column }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(row, index) in data"
          :key="row.id ?? index"
          :class="{ 'clickable-row': onRowClick }"
          @click="handleRowClick(row, index)"
        >
          <td v-for="(column, colIndex) in columns" :key="column" :class="getColumnClass(colIndex)">
            {{ row[column] }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
interface Props {
  columns: string[]
  data: { id?: string | number | undefined; [key: string]: unknown }[]
  onRowClick?: (row: { [key: string]: unknown }, index: number) => void
}

const props = defineProps<Props>()

const getColumnClass = (index: number) => {
  if (index >= 2 && index < 4) return 'hide-mobile'
  if (index >= 4) return 'hide-landscape'
  return ''
}

const handleRowClick = (row: { [key: string]: unknown }, index: number) => {
  if (props.onRowClick) {
    props.onRowClick(row, index)
  }
}
</script>

<style scoped>
.data-table {
  width: 100%;
  background-color: var(--surface-background);
  border-radius: var(--control-radius);
  border: 1px solid var(--border-color);
  color: var(--muted-text-color);
  overflow: hidden;
  font-family: 'Roboto Mono', monospace;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th {
  background-color: var(--subtle-background);
  padding: 12px 15px;
  text-align: left;
  font-weight: bold;
  font-size: 12px;
  color: var(--text-color);
  border-bottom: 1px solid var(--border-color);
}

td {
  padding: 12px 15px;
  font-size: 12px;
  color: var(--muted-text-color);
  border-bottom: 1px solid var(--subtle-border-color);
}

tbody tr:last-child td {
  border-bottom: none;
}

.clickable-row {
  cursor: pointer;
  transition: background-color 0.2s;
}

.clickable-row:hover {
  background-color: var(--subtle-background);
}

@media (max-width: 480px) {
  .hide-mobile,
  .hide-landscape {
    display: none;
  }
}

@media (min-width: 481px) and (max-width: 768px) {
  .hide-landscape {
    display: none;
  }
}
</style>
