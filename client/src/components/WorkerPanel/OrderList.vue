<template>
  <v-list class="list" v-if="ready">
    <v-dialog v-model="isOpen">
      <order-detail
        :orderDoc="selected"
        :onClickClose="closeDetail"
        :onClickFinishTrade="()=>finishTrade(selected.document_id)"
        :onClickUnfinishTrade="()=>unfinishTrade(selected.document_id)"
      />
    </v-dialog>
    <div v-for="(orderDoc, i) in orders" :key="i">
      <div v-if="!(orderDoc.order.in_progress)">
        <v-list-item>
          <v-list-item-icon>
            <v-icon color="gray" v-if="orderDoc.order.in_trade">mdi-run-fast</v-icon>
            <v-icon color="red" v-else>mdi-check-bold</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title v-text="orderDoc.order.user.display_name"></v-list-item-title>
          </v-list-item-content>
          <v-item-group>
            <router-link :to="chatRoomUrl(orderDoc)" class="link">
              <v-icon color="gray" class="mr-3" large>mdi-chat</v-icon>
            </router-link>
            <v-icon @click="() => openDetail(orderDoc)" large>mdi-file-outline</v-icon>
          </v-item-group>
        </v-list-item>
        <v-divider></v-divider>
      </div>
    </div>
  </v-list>
</template>

<script>
import OrderDetail from "./OrderDetail";

export default {
  components: {
    OrderDetail
  },
  props: {
    userID: {
      type: String
    }
  },
  data() {
    return {
      ready: false,
      isOpen: false
    };
  },
  methods: {
    openDetail: function(orderDoc) {
      this.$store.dispatch("selectOrder", orderDoc);
      this.isOpen = true;
    },
    closeDetail: function() {
      this.isOpen = false;
    },
    finishTrade: function(document_id) {
      this.$store.dispatch("finishTrade", document_id);
    },
    unfinishTrade: function(document_id) {
      this.$store.dispatch("unfinishTrade", document_id);
    },
    sortTrade: function(orderDocA, orderDocB) {
      return orderDocA.order.in_trade === orderDocB.order.in_trade
        ? 0
        : orderDocA.order.in_trade
        ? -1
        : 1;
    },
    chatRoomUrl: function(orderDoc) {
      // eslint-disable-next-line no-console
      console.log("userID is ", this.userID);
      let orderID = orderDoc.document_id;
      let chatID = orderDoc.order.chat_id;
      return `/user/${this.userID}/order/${orderID}/chats/${chatID}`;
    }
  },
  computed: {
    selected: function() {
      return this.$store.state.workerPanel.selectedOrderDoc;
    },
    orders: function() {
      let orders = this.$store.state.workerPanel.orders;
      return orders.sort(this.sortTrade);
    }
  },
  async mounted() {
    await this.$store.dispatch("fetchOrders");
    this.ready = true;
  }
};
</script>

<style scoped>
.list {
  margin-top: 15%;
}
.link {
  color: transparent;
}
</style>
