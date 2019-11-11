<template>
  <div>
    <h1 class="border-bottom title">注文票</h1>
    <b-card-group deck>
      <b-card
        class="mb-2"
        :title="`#`+(i+1)"
        img-alt="Image"
        img-top
        tag="article"
        style="max-width: 20rem;"
        v-for="(item,i) in items"
        :key="i"
      >
        <b-card-text class="card-text">注文者: {{item.orderBy}}</b-card-text>
        <b-card-text class="card-text">日程: {{item.date}}</b-card-text>
        <b-card-text class="card-text">時間: {{item.time}}</b-card-text>
        <b-button href="#" variant="primary">注文詳細</b-button>
      </b-card>
    </b-card-group>
  </div>
</template>

<script>
import { BCard, BCardGroup, BCardText, BButton } from 'bootstrap-vue'
export default {
  components: {
    BCard,
    BCardGroup,
    BCardText,
    BButton
  },
  data () {
    return {
      items: [
        {
          orderBy: 'mizuno',
          date: '本日 11/9 (土)',
          time: '12:00AM~12:30AM',
          detail: '注文詳細はこちら'
        },
        {
          orderBy: 'mizuno',
          date: '本日 11/9 (土)',
          time: '12:00AM~12:30AM',
          detail: '注文詳細はこちら'
        },
        {
          orderBy: 'mizuno',
          date: '本日 11/9 (土)',
          time: '12:00AM~12:30AM',
          detail: '注文詳細はこちら'
        },
        {
          orderBy: 'mizuno',
          date: '本日 11/9 (土)',
          time: '12:00AM~12:30AM',
          detail: '注文詳細はこちら'
        }
      ]
    }
  }
}
</script>

<style scoped>
.table {
  width: 100%;
  margin: auto;
}
.title {
  text-align: center;
  margin-bottom: 20px;
  font-size: 1.5rem;
}
.mb-2 {
  margin: auto;
}
.card-text {
  text-align: left;
}
.card-title {
  text-align: left;
}
</style>
