<template>
  <el-progress
    v-if="globalProgress > 0 && globalProgress < 100"
    :percentage="globalProgress"
    :stroke-width="2"
    :show-text="false"
    :color="customColors"
    class="global-progress-bar"
  />
  <div id="app">
    <header class="grid-content header-color">
      <div class="header-content">
        <div class="brand">
          <el-dropdown trigger="click">
            <a href="#">{{ title }}</a>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleShowCheckVersionDialog"
                  >版本检测
                </el-dropdown-item>
                <el-dropdown-item @click="manusForm.show = true"
                  >手动升级
                </el-dropdown-item>
                <el-dropdown-item @click="handleClearData"
                  >清空数据
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
        <div class="dark-switch">
          <el-switch
            v-model="darkmodeSwitch"
            inline-prompt
            active-text="Dark"
            inactive-text="Light"
            @change="toggleDark"
            style="
              --el-switch-on-color: #444452;
              --el-switch-off-color: #589ef8;
            "
          />
        </div>
      </div>
    </header>
    <section>
      <el-main>
        <el-table
          :data="paginatedTableData"
          style="width: 100%"
          :border="true"
          :preserve-expanded-content="true"
          :cell-style="{ padding: mobileLayout ? '4px' : '8px' }"
        >
          <el-table-column type="expand">
            <template #default="props">
              <ViewExpand :row="props.row" />
            </template>
            <!--            <template #default="props">-->
            <!--              <div m="4">-->
            <!--                <p m="t-0 b-2">接口: {{ props.row.phy }}</p>-->
            <!--                            <el-timeline style="max-width: 200px">-->
            <!--                              <el-timeline-item-->
            <!--                                v-for="(activity, index) in props.row.statusList"-->
            <!--                                :key="index"-->
            <!--                                :color="activity.connected ? '#55f604' : 'red'"-->
            <!--                                :hollow="false"-->
            <!--                                :timestamp="activity.timestamp"-->
            <!--                              >-->
            <!--                                <span-->
            <!--                                  :style="{ color: activity.connected ? '#55f604' : 'red' }"-->
            <!--                                >-->
            <!--                                  {{ activity.connected ? '在线' : '离线' }}-->
            <!--                                </span>-->
            <!--                              </el-timeline-item>-->
            <!--                            </el-timeline>-->
            <!--              </div>-->
            <!--            </template>-->
          </el-table-column>
          <el-table-column prop="hostname" label="名称" sortable>
            <template #default="scope">
              {{
                scope.row.nickName === ''
                  ? scope.row.hostname
                  : scope.row.hostname === '*'
                    ? scope.row.nickName
                    : `${scope.row.hostname}(${scope.row.nickName})`
              }}
            </template>
          </el-table-column>
          <el-table-column prop="ip" label="IP地址" sortable />
          <el-table-column prop="mac" label="Mac地址" sortable />
          <el-table-column prop="starTime" label="连接时间" sortable />
          <el-table-column prop="online" label="状态" sortable>
            <template #default="scope">
              <el-tag v-if="scope.row.online" type="success">在线</el-tag>
              <el-tag v-else type="danger">离线</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template #default="{ row }">
              <el-dropdown trigger="click">
                <el-button size="small" type="text">功能操作</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleChangeNickName(row)"
                      >修改名称
                    </el-dropdown-item>
                    <el-dropdown-item @click="handleGoToTimeLineDialog(row)"
                      >查看设备
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <el-pagination
          style="margin-top: 20px"
          background
          layout="prev, pager, next"
          :total="filteredTableData.length"
          :page-size="pageSize"
          :current-page="currentPage"
          :pager-count="mobileLayout ? 3 : 7"
          @current-change="handlePageChange"
        />
      </el-main>
    </section>
    <footer></footer>
  </div>

  <!--  客户端程序升级-->
  <el-dialog v-model="manusForm.show" align-center width="500">
    <template #header><span>程序升级</span></template>
    <el-input
      v-model="manusForm.binUrl"
      autocomplete="off"
      placeholder="请输入程序Url地址～"
    />

    <template #footer>
      <div class="dialog-footer">
        <el-upload
          class="upload-demo"
          :http-request="handleUploadUpgradeBin"
          :limit="1"
        >
          <template #trigger>
            <el-button type="primary" :disabled="manusForm.binUrl.length > 0"
              >上传文件升级
            </el-button>
          </template>
          <!-- 添加额外按钮 -->
          <el-button
            style="margin-left: 10px"
            type="danger"
            @click="handleUpdate"
          >
            文件url升级
          </el-button>
        </el-upload>
      </div>
    </template>
  </el-dialog>

  <UpgradeDialog ref="upgradeRef" />
  <ClientTimeLineDialog ref="clientTimeLineDialogRef" />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useDark, useToggle } from '@vueuse/core'
import { Client } from './utils/type.ts'
import ClientTimeLineDialog from './components/ClientTimeLineDialog.vue'
import {
  showErrorTips,
  showLoading,
  showSucessTips,
  showTips,
  showWarmDialog,
  showWarmTips,
  xhrPromise,
} from './utils/utils.ts'
import { EventAwareSSEClient } from './utils/sseclient.ts'
import { ElMessageBox } from 'element-plus'
import ViewExpand from './components/expand/ViewExpand.vue'
import UpgradeDialog from './components/expand/UpgradeDialog.vue'

const title = ref<string>('客户端列表')
const clientTimeLineDialogRef = ref<InstanceType<
  typeof ClientTimeLineDialog
> | null>(null)

const manusForm = ref({
  show: false,
  binUrl: '',
})
const customColors = [
  { color: '#f56c6c', percentage: 20 },
  { color: '#e6a23c', percentage: 40 },
  { color: '#5cb87a', percentage: 60 },
  { color: '#1989fa', percentage: 80 },
  { color: '#6f7ad3', percentage: 100 },
]
const appinfo = ref<any>()
const globalProgress = ref(0)
const isDark = useDark()
const darkmodeSwitch = ref(isDark)
const toggleDark = useToggle(isDark)
const source = ref<EventAwareSSEClient | null>()
// 搜索关键字
const searchKeyword = ref<string>('')
const pageSize = ref<number>(50)
const currentPage = ref<number>(1)
const tableData = ref<Client[]>([])
// 分页后的表格数
const paginatedTableData = computed<Client[]>(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredTableData.value.slice(start, end)
})
// 过滤后的表格数据（根据搜索关键字）
const filteredTableData = computed<Client[]>(() => {
  return tableData.value.filter(() => !searchKeyword.value)
})

function renderTable(data: any) {
  tableData.value = data as Client[]
}

const getVersion = () => {
  // versionDialogVisible.value = true
  fetch('../api/version', { credentials: 'include', method: 'GET' })
    .then((res) => {
      return res.json()
    })
    .then((json) => {
      if (json && json.code === 0 && json.data) {
        appinfo.value = json.data
        if (json.data && json.data.appVersion) {
          title.value = `客户端列表 ${json.data.appVersion}`
        }
      }
    })
    .catch(() => {
      showErrorTips('失败')
    })
}

const upgradeRef = ref<InstanceType<typeof UpgradeDialog> | null>(null)

const handleShowCheckVersionDialog = () => {
  if (upgradeRef.value) {
    upgradeRef.value.openUpgradeDialog()
  }
}
// 自定义上传函数
const handleUploadUpgradeBin = (options: any) => {
  const { file } = options
  const formData = new FormData()
  formData.append('file', file)
  const loading = showLoading('程序更新中...')
  globalProgress.value = 0
  manusForm.value.show = false
  xhrPromise({
    url: '../api/upgrade',
    method: 'POST',
    data: formData,
    onUploadProgress: (progress: string) => {
      console.log(`上传进度：${progress}`)
      loading.setText(`程序更新中...${progress}%`)
      globalProgress.value = parseInt(progress)
    },
  })
    .then((data: any) => {
      console.log('请求成功', data)
      // 上传成功的回调
      const json = JSON.parse(data.data)
      if (json.code !== 0) {
        if (json.msg !== '') {
          showErrorTips(json.msg)
        }
      } else {
        if (json.msg !== '') {
          showSucessTips(json.msg)
        }
      }
    })
    .catch((error) => {
      console.error('请求失败', error)
      // 上传失败的回调
      //showErrorTips('上传失败的回调')
    })
    .finally(() => {
      setTimeout(function () {
        loading.close()
        globalProgress.value = 0
        manusForm.value.show = false
        window.location.reload()
      }, 4000)
    })
}

const handleUpdate = () => {
  if (manusForm.value.binUrl.length > 0) {
    const binUrl = manusForm.value.binUrl
    console.log('upgradeByUrl', binUrl)
    const loading = showLoading('程序升级中...')
    manusForm.value.show = false
    fetch('../api/upgrade', {
      credentials: 'include',
      method: 'PUT',
      body: binUrl,
    })
      .then((res) => {
        return res.json()
      })
      .then((json) => {
        showTips(json.code, json.msg)
      })
      .catch(() => {
        showWarmTips('更新失败')
      })
      .finally(() => {
        setTimeout(function () {
          loading.close()
          window.location.reload()
        }, 4000)
      })
  } else {
    showWarmTips('请正确输入url地址')
  }
}

const fetchData = () => {
  console.log('fetchData')
  fetch(`../api/get/clients`, {
    credentials: 'include',
    method: 'GET',
  })
    .then((res) => res.json())
    .then((json) => {
      console.log('fetchData', json)
      if (json && json.code === 0 && json.data) {
        console.log(json)
        renderTable(json.data)
      }
    })
    .catch(() => {
      // showErrorTips('获取服务器信息失败')
    })
}

const handleClearData = () => {
  showWarmDialog(
    `确定清空临时数据吗？`,
    () => {
      fetch('../api/clear', { credentials: 'include', method: 'DELETE' })
        .then((res) => {
          return res.json()
        })
        .then((json) => {
          showTips(json.code, json.msg)
        })
        .catch(() => {
          showErrorTips('清空失败')
        })
    },
    () => {},
  )
}

const handleChangeNickName = (row: Client) => {
  console.log('handleChangeNickName', row)
  ElMessageBox.prompt('请输入设备昵称', '修改昵称', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputValue: row.nickName,
  }).then(({ value }) => {
    row.nickName = value
    fetch('../api/nick/set', {
      credentials: 'include',
      method: 'POST',
      body: JSON.stringify(row),
    })
      .then((res) => {
        return res.json()
      })
      .then((json) => {
        console.log('handleChangeNickName', json)
        showTips(json.code, json.msg)
      })
      .catch((error) => {
        console.log('error', error)
        showErrorTips('修改昵称失败')
      })
  })
}

// 调整详情
const handleGoToTimeLineDialog = (row: Client) => {
  console.log('handleGoToTimeLineDialog', row)
  if (clientTimeLineDialogRef.value) {
    clientTimeLineDialogRef.value.openClientDialog(row)
  }
}
// 分页切换
const handlePageChange = (page: number) => {
  currentPage.value = page
}

// 响应式布局相关
const mobileLayout = ref(false)
const checkMobile = () => {
  mobileLayout.value = window.innerWidth < 768
}

// 弹窗宽度控制
const dialogWidth = ref('500px')
const updateDialogWidth = () => {
  checkMobile()
  dialogWidth.value = mobileLayout.value ? '90%' : '500px'
}

const connectSSE = () => {
  try {
    const sseUrl = `../api/client/sse`
    console.log('connectSSE', sseUrl)
    source.value = new EventAwareSSEClient(sseUrl)
    source.value.addEventListener('update', (data) => {
      console.log('update', data)
      renderTable(data)
    })
    source.value.connect()
  } catch (e) {
    console.error('connectSSE err', e)
  }
}

// 初始化监听
onMounted(() => {
  window.addEventListener('resize', updateDialogWidth)
  updateDialogWidth()
})

onUnmounted(() => {
  window.removeEventListener('resize', updateDialogWidth)
})
getVersion()
connectSSE()
fetchData()
</script>

<style>
body {
  margin: 0px;
  font-family:
    -apple-system,
    BlinkMacSystemFont,
    Helvetica Neue,
    sans-serif;
}

header {
  width: 100%;
  height: 60px;
}

.header-color {
  background: #58b7ff;
}

html.dark .header-color {
  background: #395c74;
}

.header-content {
  display: flex;
  align-items: center;
}

#content {
  margin-top: 20px;
  padding-right: 40px;
}

.brand {
  display: flex;
  justify-content: flex-start;
}

.brand a {
  color: #fff;
  background-color: transparent;
  margin-left: 20px;
  line-height: 25px;
  font-size: 25px;
  padding: 15px 15px;
  height: 30px;
  text-decoration: none;
}

.dark-switch {
  display: flex;
  justify-content: flex-end;
  flex-grow: 1;
  padding-right: 40px;
}

.global-progress-bar {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 9999;
  width: 100%;
}
</style>
