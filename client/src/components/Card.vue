<template>
  <b-card
    class="mb-2"
    :title="title"
    img-alt="Image"
    img-top
    tag="article"
    style="max-width: 20rem;"
    v-if="!(order.in_progress)"
  >
    <b-card-text class="card-text">注文者: {{order.user.display_name}}</b-card-text>
    <b-card-text class="card-text">日程: {{order.date}}</b-card-text>
    <b-card-text class="card-text">時間: {{order.time}}</b-card-text>
    <b-card-text class="card-text">場所: {{order.location}}</b-card-text>
    <b-card-text class="card-text in_trade" v-if="order.in_trade">対応未</b-card-text>
    <b-card-text class="card-text" v-else>対応済</b-card-text>
    <div class="card-button-group">
      <b-button href="#" variant="primary" class="card-button-datail">注文詳細</b-button>
      <b-button
        href="#"
        variant="primary"
        class="card-button-yet"
        v-on:click="onClick"
        v-if="order.in_trade"
      >対応済みにする</b-button>
      <b-button
        href="#"
        variant="primary"
        class="card-button-done"
        v-else
        v-on:click="onClick"
      >対応済みを解除する</b-button>
    </div>
  </b-card>
</template>

<script>
import { BCard, BCardText, BButton, BButtonGroup } from 'bootstrap-vue'
export default {
  components: {
    BCard,
    BButton,
    BCardText,
    BButtonGroup
  },
  props: {
    orderDocument: {
      type: Object,
      required: true
    },
    title: {
      type: String,
      required: true
    },
    onClickToggleFinishedStatus: {
      type: Function,
      required: true
    }
  },
  methods: {
    onClick: function () {
      this.onClickToggleFinishedStatus(this.orderDocument.document_id)
    }
  },
  computed: {
    order: function () {
      return this.orderDocument.order
    }
  }
}
</script>

<style>
.mb-2 {
  margin: auto;
}
.card-text {
  text-align: left;
}
.card-title {
  text-align: left;
}
.in_trade {
  color: red;
}
.card-button-group {
  display: flex;
  flex-direction: column;
}
.card-button-datail {
  margin-bottom: 15px;
}
.card-button-yet {
  border-color: red;
  background-color: red;
}
</style>
