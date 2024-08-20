<script setup lang="ts">
import * as monaco from 'monaco-editor'
import { onMounted } from 'vue'

const props = defineProps<{
  language: { type: String; required: true }
  value: String
}>()

self.MonacoEnvironment = {
  getWorker: function (workerId, label) {
    const getWorkerModule = (moduleUrl, label) => {
      return new Worker(self.MonacoEnvironment.getWorkerUrl(moduleUrl), {
        name: label,
        type: 'module'
      })
    }

    switch (label) {
      case 'json':
        return getWorkerModule('/monaco-editor/esm/vs/language/json/json.worker?worker', label)
      case 'css':
      case 'scss':
      case 'less':
        return getWorkerModule('/monaco-editor/esm/vs/language/css/css.worker?worker', label)
      case 'html':
      case 'handlebars':
      case 'razor':
        return getWorkerModule('/monaco-editor/esm/vs/language/html/html.worker?worker', label)
      case 'typescript':
      case 'javascript':
        return getWorkerModule('/monaco-editor/esm/vs/language/typescript/ts.worker?worker', label)
      default:
        return getWorkerModule('/monaco-editor/esm/vs/editor/editor.worker?worker', label)
    }
  }
}

onMounted(() => {
  monaco.editor.create(document.getElementById('editor'), {
    value: props.value,
    language: props.language,
    automaticLayout: true,
    fontFamily: 'JetBrains Mono'
  })
})
</script>

<template>
  <div id="editor" class="h-full"></div>
</template>
