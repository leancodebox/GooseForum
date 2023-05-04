<script setup>
import {NList, NListItem, useMessage} from 'naive-ui';
import {ref} from "vue";
import {addDays, isYesterday} from "date-fns";


const message = useMessage()


function success() {
  message.success(
      `还挺大`
  )
}


const value = ref(addDays(Date.now(), 1).valueOf())

function handleUpdateValue(_, {year, month, date}) {
  message.success(`${year}-${month}-${date}`)
}

function isDateDisabled(timestamp) {
  return isYesterday(timestamp);
}
</script>
<template>
  <n-list class="thh-index" :show-divider="false">
    <n-list-item>
      <div class="greetings" style=" text-align: center;">
        <h1 class="green">本场的辩题是</h1>
        <h1 class="green">「故事的结局重不重要」</h1>
        <p>故事的结局重不重要呢</p>
        <p>如果有一个绝好的故事</p>
        <p>却以一个烂透了的结局结束</p>
        <p>那这到底是一个好故事还是烂故事呢</p>
        <p>比如说有一场无疾而终的恋爱</p>
        <p>或者是一次没有目标的旅程</p>
        <p>所以很好奇</p>
        <p>一个故事的结局重要吗</p>
      </div>
    </n-list-item>
  </n-list>
</template>

<style scoped>
h1 {
  font-weight: 500;
  font-size: 2.6rem;
  top: -10px;
}

h3 {
  font-size: 1.2rem;
}

.greetings {
  max-width: 1280px;
  margin: 0 auto;
}

.greetings h1,
.greetings h3 {
  text-align: center;
}

/*@media (min-width: 1024px) {*/
/*  .greetings h1,*/
/*  .greetings h3 {*/
/*    text-align: left;*/
/*  }*/
/*}*/
.thh-index .n-list-item {
  padding-left: 20px;
  padding-right: 20px;
}

.green {
  text-decoration: none;
  /*color: #5BACFCFF;*/
  transition: 0.4s;
}

</style>