<template>
  <v-content class="container">
    <chat :chatID="chatID" :orderID="orderID" :ownID="userID" />
  </v-content>
</template>

<script>
import Chat from "./Chat";
export default {
  components: {
    Chat
  },
  props: {
    chatID: {
      type: String,
      required: true
    },
    orderID: {
      type: String,
      required: true
    },
    userID: {
      type: String,
      required: true
    }
  },
  computed: {
    isOpen() {
      return this.$store.state.order.isOpen;
    }
  },
  methods: {
    closeOrderDetail() {
      this.$store.dispatch("closeOrderDetail");
    }
  },
  mounted() {
    this.$store.dispatch("fetchChats", this.chatID);
  }
};
</script>

<style scoped>
.container {
  height: 100%;
}
</style>
