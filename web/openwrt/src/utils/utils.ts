import { ElMessage, ElMessageBox } from 'element-plus'

export function showWarmDialog(title: string, ok: any, cancel: any) {
  ElMessageBox.confirm(title, 'Warning', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(() => {
      ok()
    })
    .catch(() => {
      cancel()
    })
}
export function showErrorTips(message: string) {
  ElMessage({
    showClose: true,
    message: message,
    type: 'error',
  })
}
export function showTips(code: any, message: string) {
  if (code === 0) {
    showSucessTips(message)
  } else {
    showWarmTips(message)
  }
}

export function showSucessTips(message: string) {
  ElMessage({
    showClose: true,
    message: message,
    type: 'success',
  })
}

export function showWarmTips(message: string) {
  ElMessage({
    showClose: true,
    message: message,
    type: 'warning',
  })
}

export function syntaxHighlight(json: string): string {
  // 转义特殊字符防止 XSS
  json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

  // 正则匹配 JSON 元素并分配类名
  return json.replace(
    /("(\\u[\dA-Fa-f]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+-]?\d+)?)/g,
    (match) => {
      let cls = 'number'
      if (/^"/.test(match)) {
        cls = match.endsWith(':') ? 'key' : 'string' // 键名与字符串区分
      } else if (/true|false/.test(match)) {
        cls = 'boolean'
      } else if (/null/.test(match)) {
        cls = 'null'
      }
      return `<span class="${cls}">${match}</span>` // 直接内联类名判断[1,6](@ref)
    },
  )
}