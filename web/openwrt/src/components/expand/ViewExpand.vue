<template>
  <div class="main">
    <el-form
      label-position="left"
      label-width="auto"
      inline
      class="proxy-table-expand"
    >
      <el-form-item label="名称" v-if="isMobile()">
        <span>{{
          row.nickName === ''
            ? row.hostname
            : row.hostname === '*'
              ? row.nickName
              : `${row.hostname}(${row.nickName})`
        }}</span>
      </el-form-item>
      <el-form-item label="网络接口">
        <span>{{ row.phy }}</span>
      </el-form-item>
      <el-form-item label="连接时间" v-if="isMobile()">
        <span>{{ row.starTime }}</span>
      </el-form-item>
      <el-form-item label="Mac地址" v-if="isMobile()">
        <span>{{ row.mac }}</span>
      </el-form-item>
    </el-form>

    <!--    <el-timeline style="max-width: 200px">-->
    <!--      <el-timeline-item-->
    <!--        v-for="(activity, index) in row.statusList"-->
    <!--        :key="index"-->
    <!--        :color="activity.connected ? '#55f604' : 'red'"-->
    <!--        :hollow="false"-->
    <!--        :timestamp="activity.timestamp"-->
    <!--      >-->
    <!--        <span :style="{ color: activity.connected ? '#55f604' : 'red' }">-->
    <!--          {{ activity.connected ? '在线' : '离线' }}-->
    <!--        </span>-->
    <!--      </el-timeline-item>-->
    <!--    </el-timeline>-->
  </div>
</template>

<script setup lang="ts">
import { isMobile } from '../../utils/utils.ts'

defineProps<{
  row: any
}>()
</script>

<style>
ul {
  list-style-type: none;
  padding: 5px;
}

ul li {
  justify-content: space-between;
  padding: 5px;
}

ul .annotation-key {
  width: 300px;
  display: inline-block;
  vertical-align: middle;
}

.title-text {
  color: #99a9bf;
}

.main {
  margin-left: 100px;
}
</style>
