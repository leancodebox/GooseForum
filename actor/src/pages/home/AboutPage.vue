<script setup>
import {NCard, NCountdown, NGi, NGrid, NSpace,NTimeline, NTimelineItem} from 'naive-ui'
import moment from "moment/moment";
const date1 = new Date('2023/12/24')
const time1 = parseInt(date1-(new Date()))
function renderCountdown({  hours,
                           minutes,
                           seconds}){
  let day = parseInt(hours/24);
  let dayHours = hours%24
  return `至 2023/12/24 ${String(day).padStart(3, "0")} 天 ${String(dayHours).padStart(2, "0")} 小时 ${String(minutes).padStart(2, "0")} 分 ${String(seconds).padStart(2, "0")} 秒`;
}

let dayInfoList = [];
let nowT = moment()
let t = moment(moment().format("YYYY-01-01"))
for (let i = 1; i < 12; i++) {
  t.add(1, "months")
  let type = 'warning'
  let lineType = 'dashed'
  if (parseInt(t.format('M')) > parseInt(nowT.format('M'))) {
    type = 'success'
    lineType = 'default'
  }
  let timeInfo = t.format('YYYY-MM-DD')
  dayInfoList.push({
    title: timeInfo,
    time: timeInfo,
    // content: timeInfo,
    type: type,
    lineType: lineType
  })
}
dayInfoList.sort(function (item1, item2) {
  return item1.time > item2.time ? -1 : 1
})

dayInfoList.push({title: "start"})
dayInfoList.unshift({title: "end", type: "success"});
</script>
<template>
  <n-card style="margin:0 auto">
    <n-space>
      <n-grid>
        <n-gi span="24">
          <p>青山不改绿水长流</p>
        </n-gi>

        <n-gi span="24">
          <n-countdown ref="countdown" :render="renderCountdown" :duration="time1" :active="true"/>
        </n-gi>
        <n-gi span="24">

          <n-timeline :size="'large'">
            <n-timeline-item v-for="timeInfo in dayInfoList" :type="timeInfo.type"
                             :title="timeInfo.title"
                             :content="timeInfo.content"
                             :time="timeInfo.time"
                             :line-type="timeInfo.lineType"
            />
          </n-timeline>
        </n-gi>
      </n-grid>

    </n-space>
  </n-card>
</template>


