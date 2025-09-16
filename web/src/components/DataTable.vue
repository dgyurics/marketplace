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
        <tr v-for="(row, index) in data" :key="index">
          <td v-for="column in columns" :key="column">
            {{ row[column] }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  columns: string[]
  data: { [key: string]: unknown }[]
}>()
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
</style>
