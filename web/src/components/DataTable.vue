<template>
  <div class="data-table">
    <table>
      <thead>
        <tr>
          <th v-for="column in columns" :key="column">
            {{ column }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(row, index) in data"
          :key="index"
          :class="{ 'clickable-row': onRowClick }"
          @click="handleRowClick(row, index)"
        >
          <td v-for="column in columns" :key="column">
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
  data: { [key: string]: unknown }[]
  onRowClick?: (row: { [key: string]: unknown }, index: number) => void
}

const props = defineProps<Props>()

const handleRowClick = (row: { [key: string]: unknown }, index: number) => {
  if (props.onRowClick) {
    props.onRowClick(row, index)
  }
}
</script>

<style scoped>
.data-table {
  width: 100%;
  background-color: #f8f9fa;
  border-radius: 4px;
  border: 1px solid #ddd;
  color: #212529;
  overflow: hidden;
  font-family: 'Roboto Mono', monospace;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th {
  padding: 12px 15px;
  text-align: left;
  font-weight: bold;
  font-size: 12px;
  color: #333;
  border-bottom: 1px solid #ddd;
}

td {
  padding: 12px 15px;
  font-size: 12px;
  color: #666;
  border-bottom: 1px solid #f0f0f0;
}

tbody tr:last-child td {
  border-bottom: none;
}

.clickable-row {
  cursor: pointer;
  transition: background-color 0.2s;
}

.clickable-row:hover {
  background-color: #f0f0f0;
}
</style>
