<script setup>
import {getSysInfo} from "@/service/remote";
import {NGrid, NGridItem, NCard, NProgress} from "naive-ui"
import {ref, onUnmounted} from "vue"

function reload() {
  getSysInfo().then(r => {
    sysInfo.value = r.data.result
  })
}

reload()
const timer = ref(null)
timer.value = setInterval(() => {
  reload()
}, 1000 * 2)

onUnmounted(() => {
  clearInterval(timer.value)
  timer.value = null
})
const sysInfo = ref({
  "os": {
    "goos": "darwin",
    "numCpu": 8,
    "compiler": "gc",
    "goVersion": "go1.20",
    "numGoroutine": 13
  },
  "cpu": {
    "cpus": [
      0.00,
      0.00,
      0.00,
      0.00,
      0.00,
      0.00,
      0.00,
      0.00
    ],
    "cores": 8
  },
  "ram": {
    "usedMB": 0,
    "totalMB": 1,
    "usedPercent": 0
  },
  "disk": {
    "usedMB": 0,
    "usedGB": 0,
    "totalMB": 1,
    "totalGB": 1,
    "usedPercent": 0
  }
})
</script>
<template>
  <n-grid cols="1 600:2">
    <n-grid-item>
      <n-card title="cpu" style="height: 240px">
        cpu-cores:{{ sysInfo.cpu.cores }}
        <n-progress type="line" v-for="percentage in sysInfo.cpu.cpus "
                    :status="percentage>80?'error':(percentage>50?'warning':'success')"
                    :percentage="percentage.toFixed(2)"
                    :indicator-placement="'inside'"/>
      </n-card>
    </n-grid-item>
    <n-grid-item>
      <n-card title="ram" style="height: 240px">
        {{ sysInfo.ram.usedMB }} MB
        /
        {{ sysInfo.ram.totalMB }} MB

        <n-progress type="circle"
                    :status="sysInfo.ram.usedPercent>80?'error':(sysInfo.ram.usedPercent>50?'warning':'success')"
                    :percentage="sysInfo.ram.usedPercent.toFixed(2)" :indicator-placement="'inside'">
          {{ sysInfo.ram.usedPercent.toFixed(2) }}%
        </n-progress>
      </n-card>
    </n-grid-item>
    <n-grid-item>
      <n-card title="disk" style="height: 240px">
        {{ sysInfo.disk.usedGB }} GB
        /
        {{ sysInfo.disk.totalGB }} GB
        <n-progress type="circle"
                    :status="sysInfo.disk.usedPercent>88?'error':(sysInfo.disk.usedPercent>50?'warning':'success')"
                    :percentage="sysInfo.disk.usedPercent"
                    :indicator-placement="'inside'">
          {{ sysInfo.disk.usedPercent.toFixed(2) }}%
        </n-progress>
      </n-card>
    </n-grid-item>
    <n-grid-item>
      <n-card title="os" style="height: 240px">
        compiler:{{ sysInfo.os.compiler }} <br>
        goos:{{ sysInfo.os.goos }} <br>
        goVersion:{{ sysInfo.os.goVersion }} <br>
        numGoroutine:{{ sysInfo.os.numGoroutine }} <br>
        numCpu:{{ sysInfo.os.numCpu }} <br>
      </n-card>
    </n-grid-item>
  </n-grid>

</template>