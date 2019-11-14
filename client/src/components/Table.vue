<template>
  <div>
    <h1 class="border-bottom title">注文票</h1>
    <b-card-group deck v-for="(orderDocument,i) in orderDocuments" :key="i" class="card_group">
      <card
        :orderDocument="orderDocument"
        :onClickToggleFinishedStatus="onClickToggleFinishedStatus"
        :title="`#`+i"
      />
    </b-card-group>
  </div>
</template>

<script>
import axios from 'axios'
import { BCardGroup, BCardText, BButton } from 'bootstrap-vue'
import Card from './Card'
export default {
  components: {
    Card,
    BCardGroup,
    BCardText,
    BButton
  },
  data () {
    return {
      orderDocuments: []
    }
  },
  methods: {
    onClickToggleFinishedStatus: function (documentID) {
      let selected = this.orderDocuments.find(orderDocument => {
        return orderDocument.documentID === documentID
      })
      selected.order.finished = !selected.order.finished
      if (selected) {
        axios
          .post(`http://localhost:1964/order`, selected)
          .then(function () {
            this.orderDocuments = this.orderDocuments.map(function (
              orderDocument
            ) {
              if (orderDocument.documentID === documentID) {
                return selected
              }
              return orderDocument
            })
          })
          .catch(function (e) {
            console.error(e)
          })
      }
    }
  },
  mounted () {
    axios
      .get('http://localhost:1964/order_list')
      .then(r => {
        this.orderDocuments = r.data
      })
      .catch(e => {
        console.error(e)
      })
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
.card_group {
  justify-content: center;
}
</style>
